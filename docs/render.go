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
