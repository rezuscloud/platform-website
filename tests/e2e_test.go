//go:build e2e
// +build e2e

package tests

import (
	"context"
	"os"
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

func TestE2EPageLoad(t *testing.T) {
	t.Skip("Skipping page load E2E test - Chrome DevTools websocket timeout issues in CI environment. Content is tested by TestE2EProgressiveEnhancement.")
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

func TestE2EHTMXSectionLoad(t *testing.T) {
	ctx, cancel := newChromedpContext()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var bodyText string

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()+"/sections/hero"),
		chromedp.WaitVisible("#hero"),
		chromedp.Evaluate(`document.body.innerText`, &bodyText),
	)
	require.NoError(t, err)

	assert.Contains(t, bodyText, "Your Personal",
		"Section endpoint should contain expected content")
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
					hasNav: document.querySelector('nav') === null,
					hasFooter: document.querySelector('footer') === null,
					hasScene: document.getElementById('scene') !== null,
					hasTargets: ['terminal', 'mac', 'linux']
						.every(id => document.querySelector('[data-scene-target="' + id + '"]') !== null),
					hasNoJSBootstrap: document.documentElement.classList.contains('js'),
					hasSceneScript: document.querySelector('script[src="/assets/js/scene.js"]') !== null
				};
			})()
		`, &contentChecks),
	)
	require.NoError(t, err)

	assert.True(t, contentChecks["hasTitle"], "Page should have title")
	assert.True(t, contentChecks["hasH1"], "Page should have h1")
	assert.True(t, contentChecks["hasMain"], "Page should have main")
	assert.True(t, contentChecks["hasNav"], "Homepage should not render nav")
	assert.True(t, contentChecks["hasFooter"], "Homepage should not render footer")
	assert.True(t, contentChecks["hasScene"], "Scene should be present")
	assert.True(t, contentChecks["hasTargets"], "All scene targets should exist")
	assert.True(t, contentChecks["hasNoJSBootstrap"], "No-JS bootstrap should switch to js class")
	assert.True(t, contentChecks["hasSceneScript"], "Scene script should be loaded")
}

func TestE2ESceneCameraPhases(t *testing.T) {
	t.Skip("Skipping geometry-heavy scene E2E in CI until preview environment is stable enough for deterministic viewport assertions.")
}
