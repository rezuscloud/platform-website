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

func readBoolChecks(ctx context.Context, script string) (map[string]bool, error) {
	var checks map[string]bool
	err := chromedp.Run(ctx, chromedp.Evaluate(script, &checks))
	return checks, err
}

func waitForBoolChecks(t *testing.T, ctx context.Context, script string, predicate func(map[string]bool) bool) map[string]bool {
	t.Helper()

	var checks map[string]bool
	require.Eventually(t, func() bool {
		var err error
		checks, err = readBoolChecks(ctx, script)
		if err != nil {
			return false
		}

		return predicate(checks)
	}, 10*time.Second, 100*time.Millisecond)

	return checks
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
		chromedp.WaitVisible("#linux-panel"),
		chromedp.WaitVisible("#terminal-panel"),
	)
	require.NoError(t, err)

	contentChecks := waitForBoolChecks(t, ctx, `
			(function() {
				const linux = document.getElementById('linux-panel');
				const mac = document.getElementById('mac-panel');
				const terminal = document.getElementById('terminal-panel');
				return {
					hasTitle: document.title.includes('RezusCloud'),
					hasH1: document.querySelector('h1') !== null,
					hasMain: document.querySelector('main') !== null,
					hasSummary: document.getElementById('shell-summary') !== null,
					hasTerminal: terminal !== null,
					hasMac: mac !== null,
					hasLinux: linux !== null,
					hasSceneRoot: document.querySelector('[data-scene-root]') !== null,
					hasSceneScript: document.querySelector('script[src="/assets/js/scene.js"]') !== null,
					hasNestedTerminal: !!(mac && terminal && mac.contains(terminal)),
					hasNestedMac: !!(linux && mac && linux.contains(mac)),
					hasNoJSBootstrap: document.documentElement.classList.contains('js'),
					hasHTMX: document.querySelector('script[src="/assets/js/htmx.min.js"]') !== null
				};
			})()
		`, func(checks map[string]bool) bool {
			return checks["hasTitle"] &&
				checks["hasH1"] &&
				checks["hasMain"] &&
				checks["hasSummary"] &&
				checks["hasTerminal"] &&
				checks["hasMac"] &&
				checks["hasLinux"] &&
				checks["hasSceneRoot"] &&
				checks["hasSceneScript"] &&
				checks["hasNestedTerminal"] &&
				checks["hasNestedMac"] &&
				checks["hasNoJSBootstrap"] &&
				checks["hasHTMX"]
		})

	assert.True(t, contentChecks["hasTitle"], "Page should have title")
	assert.True(t, contentChecks["hasH1"], "Page should have h1")
	assert.True(t, contentChecks["hasMain"], "Page should have main")
	assert.True(t, contentChecks["hasSummary"], "Shell summary should be present")
	assert.True(t, contentChecks["hasTerminal"], "Terminal surface should be present")
	assert.True(t, contentChecks["hasMac"], "Mac surface should be present")
	assert.True(t, contentChecks["hasLinux"], "Linux surface should be present")
	assert.True(t, contentChecks["hasSceneRoot"], "Scene root should be present")
	assert.True(t, contentChecks["hasSceneScript"], "Scene script should be loaded")
	assert.True(t, contentChecks["hasNestedTerminal"], "Terminal should be nested inside Mac")
	assert.True(t, contentChecks["hasNestedMac"], "Mac should be nested inside Linux")
	assert.True(t, contentChecks["hasNoJSBootstrap"], "No-JS bootstrap should switch to js class")
	assert.True(t, contentChecks["hasHTMX"], "HTMX should be loaded")
}

func TestE2EDirectRoutesPreserveNesting(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		script  string
		expects map[string]bool
	}{
		{
			name: "terminal route stays leaf only",
			path: "/apps/terminal",
			script: `(() => {
				const linux = document.getElementById('linux-panel');
				const mac = document.getElementById('mac-panel');
				const terminal = document.getElementById('terminal-panel');
				return {
					hasLinux: !!linux,
					hasMac: !!mac,
					hasTerminal: !!terminal
				};
			})()`,
			expects: map[string]bool{"hasLinux": false, "hasMac": false, "hasTerminal": true},
		},
		{
			name: "mac route keeps nested terminal",
			path: "/apps/mac",
			script: `(() => {
				const mac = document.getElementById('mac-panel');
				const terminal = document.getElementById('terminal-panel');
				return {
					hasMac: !!mac,
					hasTerminal: !!terminal,
					hasNestedTerminal: !!(mac && terminal && mac.contains(terminal)),
					hasLinux: !!document.getElementById('linux-panel')
				};
			})()`,
			expects: map[string]bool{"hasMac": true, "hasTerminal": true, "hasNestedTerminal": true, "hasLinux": false},
		},
		{
			name: "linux route keeps nested mac and terminal",
			path: "/apps/linux",
			script: `(() => {
				const linux = document.getElementById('linux-panel');
				const mac = document.getElementById('mac-panel');
				const terminal = document.getElementById('terminal-panel');
				return {
					hasLinux: !!linux,
					hasMac: !!mac,
					hasTerminal: !!terminal,
					hasNestedMac: !!(linux && mac && linux.contains(mac)),
					hasNestedTerminal: !!(mac && terminal && mac.contains(terminal))
				};
			})()`,
			expects: map[string]bool{"hasLinux": true, "hasMac": true, "hasTerminal": true, "hasNestedMac": true, "hasNestedTerminal": true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := newChromedpContext()
			defer cancel()

			ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
			defer cancel()

			require.NoError(t, chromedp.Run(ctx,
				chromedp.Navigate(getBaseURL()+tc.path),
				chromedp.WaitVisible("body"),
			))

			checks := waitForBoolChecks(t, ctx, tc.script, func(checks map[string]bool) bool {
				for key, expected := range tc.expects {
					if checks[key] != expected {
						return false
					}
				}

				return true
			})

			for key, expected := range tc.expects {
				assert.Equal(t, expected, checks[key], key)
			}
		})
	}
}

func TestE2ECrossAppFlow(t *testing.T) {
	ctx, cancel := newChromedpContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("#terminal-panel"),
		chromedp.Evaluate(`(() => {
			const button = document.querySelector('button[name="preset"][value="rezus sync demo"]');
			if (!button) return false;
			button.click();
			return true;
		})()`, nil),
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
