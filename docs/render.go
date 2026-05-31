package docs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// md is the shared goldmark instance with syntax highlighting and extras.
var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.NewTable(),
			extension.Strikethrough,
			extension.TaskList,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

// Heading represents a section heading extracted from rendered HTML.
type Heading struct {
	Level int    // 2 or 3
	ID    string // HTML anchor ID
	Text  string // Heading text
}

// ExtractHeadings pulls H2 and H3 headings from rendered HTML.
func ExtractHeadings(html string) []Heading {
	var result []Heading
	i := 0
	for i < len(html) {
		// Look for <h2 ...> or <h3 ...>
		idx := strings.Index(html[i:], "<h")
		if idx == -1 {
			break
		}
		pos := i + idx
		if pos+3 >= len(html) {
			break
		}
		ch := html[pos+2]
		if ch != '2' && ch != '3' {
			i = pos + 3
			continue
		}
		level := int(ch - '0')

		// Find end of opening tag
		tagEnd := strings.Index(html[pos:], ">")
		if tagEnd == -1 {
			break
		}
		contentStart := pos + tagEnd + 1

		// Find closing tag
		closeTag := "</h" + string(ch) + ">"
		closeIdx := strings.Index(html[contentStart:], closeTag)
		if closeIdx == -1 {
			break
		}
		content := html[contentStart : contentStart+closeIdx]

		// Extract id="..." from opening tag
		openTag := html[pos : pos+tagEnd+1]
		id := extractAttr(openTag, "id")

		// Strip any HTML tags inside the heading text
		text := stripTags(content)

		result = append(result, Heading{
			Level: level,
			ID:    id,
			Text:  text,
		})
		i = contentStart + closeIdx + len(closeTag)
	}
	return result
}

func extractAttr(tag, attr string) string {
	seek := attr + "=\""
	idx := strings.Index(tag, seek)
	if idx == -1 {
		return ""
	}
	start := idx + len(seek)
	end := strings.Index(tag[start:], "\"")
	if end == -1 {
		return ""
	}
	return tag[start : start+end]
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// Render converts markdown source to HTML with heading anchor IDs.
func Render(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("markdown render: %w", err)
	}
	return buf.String(), nil
}

// ExtractTitle returns the first H1 heading text from markdown source.
func ExtractTitle(source []byte) string {
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		if heading, ok := node.(*ast.Heading); ok && heading.Level == 1 {
			var sb strings.Builder
			for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
				if seg, ok := c.(*ast.Text); ok {
					sb.Write(seg.Segment.Value(source))
				}
			}
			title := sb.String()
			if title != "" {
				return title
			}
		}
	}
	return ""
}
