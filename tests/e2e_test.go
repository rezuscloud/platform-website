//go:build e2e
// +build e2e

package tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getBaseURL() string {
	if url := os.Getenv("BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:3000"
}

func newChromedpContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("remote-debugging-port", "9222"),
	)
	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx)
	return ctx, func() {
		ctxCancel()
		cancel()
	}
}

type panelTexts struct {
	Terminal string `json:"terminal"`
	Shell    string `json:"shell"`
	Mac      string `json:"mac"`
	Linux    string `json:"linux"`
}

func readPanelTexts(ctx context.Context) (panelTexts, error) {
	var texts panelTexts
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const text = (id) => document.getElementById(id)?.innerText || "";
				return {
					terminal: text("terminal-panel"),
					shell: text("shell-summary"),
					mac: text("mac-panel"),
					linux: text("linux-panel")
				};
			})()
		`, &texts),
	)

	return texts, err
}

func TestE2EPerformance(t *testing.T) {
	ctx, cancel := newChromedpContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var loadTime float64
	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Evaluate(`
			(function() {
				const timing = performance.timing;
				return timing.loadEventEnd - timing.navigationStart;
			})()
		`, &loadTime),
	)
	require.NoError(t, err)

	assert.Less(t, loadTime, float64(5000),
		"Page should load within 5 seconds, took %fms", loadTime)
}

func TestE2EProgressiveEnhancement(t *testing.T) {
	ctx, cancel := newChromedpContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(700*time.Millisecond),
	)
	require.NoError(t, err)

	var contentChecks map[string]bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				return {
					hasTitle: document.title.includes('RezusCloud'),
					hasH1: document.querySelector('h1') !== null,
					hasMain: document.querySelector('main') !== null,
					hasSummary: document.getElementById('shell-summary') !== null,
					hasTerminal: document.getElementById('terminal-panel') !== null,
					hasMac: document.getElementById('mac-panel') !== null,
					hasLinux: document.getElementById('linux-panel') !== null,
					hasNoJSBootstrap: document.documentElement.classList.contains('js'),
					hasHTMX: document.querySelector('script[src="/assets/js/htmx.min.js"]') !== null
				};
			})()
		`, &contentChecks),
	)
	require.NoError(t, err)

	assert.True(t, contentChecks["hasTitle"], "Page should have title")
	assert.True(t, contentChecks["hasH1"], "Page should have h1")
	assert.True(t, contentChecks["hasMain"], "Page should have main")
	assert.True(t, contentChecks["hasSummary"], "Shell summary should be present")
	assert.True(t, contentChecks["hasTerminal"], "Terminal surface should be present")
	assert.True(t, contentChecks["hasMac"], "Mac surface should be present")
	assert.True(t, contentChecks["hasLinux"], "Linux surface should be present")
	assert.True(t, contentChecks["hasNoJSBootstrap"], "No-JS bootstrap should switch to js class")
	assert.True(t, contentChecks["hasHTMX"], "HTMX should be loaded")
}

func TestE2ECrossAppFlow(t *testing.T) {
	ctx, cancel := newChromedpContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("#terminal-panel"),
		chromedp.Click(`button[name="preset"][value="rezus sync demo"]`),
	)
	require.NoError(t, err)

	var texts panelTexts
	require.Eventually(t, func() bool {
		var readErr error
		texts, readErr = readPanelTexts(ctx)
		if readErr != nil {
			return false
		}

		terminalText := strings.ToLower(texts.Terminal)
		shellText := strings.ToLower(texts.Shell)
		macText := strings.ToLower(texts.Mac)
		linuxText := strings.ToLower(texts.Linux)

		return strings.Contains(terminalText, "artifact.published") &&
			strings.Contains(shellText, "one command moved through three services") &&
			strings.Contains(macText, "deployment dossier") &&
			strings.Contains(linuxText, "reconciled") &&
			strings.Contains(linuxText, "artifact.published")
	}, 10*time.Second, 100*time.Millisecond)

	assert.Contains(t, strings.ToLower(texts.Terminal), "artifact.published")
	assert.Contains(t, strings.ToLower(texts.Shell), "one command moved through three services")
	assert.Contains(t, strings.ToLower(texts.Mac), "deployment dossier")
	assert.Contains(t, strings.ToLower(texts.Linux), "reconciled")
	assert.Contains(t, strings.ToLower(texts.Linux), "artifact.published")
}
