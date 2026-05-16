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

func newChromedpContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		// Disable remote debugging to avoid port conflicts in CI
		chromedp.Flag("remote-debugging-port", "0"),
		// Suppress verbose logging
		chromedp.Flag("log-level", "2"),
	)
	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Use a shorter timeout to fail fast on CI
	ctx, ctxCancel := chromedp.NewContext(allocCtx, chromedp.WithBrowserOption(
		chromedp.WithBrowserTimeout(30*time.Second),
	))

	return ctx, func() {
		ctxCancel()
		cancel()
	}
}

func TestE2EPageLoad(t *testing.T) {
	ctx, cancel := newChromedpContext(t)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Title(&title),
	)
	require.NoError(t, err)
	assert.Contains(t, title, "RezusCloud")
}

func TestE2EThemeToggle(t *testing.T) {
	ctx, cancel := newChromedpContext(t)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
	)
	require.NoError(t, err)

	var initialDark bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.documentElement.classList.contains('dark')`, &initialDark),
	)
	require.NoError(t, err)

	err = chromedp.Run(ctx,
		chromedp.Click("button[aria-label='Toggle theme']"),
		chromedp.Sleep(500*time.Millisecond),
	)
	require.NoError(t, err)

	var afterClickDark bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.documentElement.classList.contains('dark')`, &afterClickDark),
	)
	require.NoError(t, err)
	assert.NotEqual(t, initialDark, afterClickDark, "Theme should toggle after clicking button")
}

func TestE2EAlpineJSInitialization(t *testing.T) {
	ctx, cancel := newChromedpContext(t)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
	)
	require.NoError(t, err)

	var alpineLoaded bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`typeof Alpine !== 'undefined'`, &alpineLoaded),
	)
	require.NoError(t, err)
	assert.True(t, alpineLoaded, "Alpine.js should be loaded and initialized")

	// Verify Alpine store is reactive (this would fail if CSP blocks new Function())
	var themeStoreExists bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`typeof Alpine.store('theme') !== 'undefined'`, &themeStoreExists),
	)
	require.NoError(t, err)
	assert.True(t, themeStoreExists, "Alpine theme store should exist (CSP allows Alpine)")
}

func TestE2EHTMXSectionLoad(t *testing.T) {
	ctx, cancel := newChromedpContext(t)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var bodyText string
	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()+"/sections/hero"),
		chromedp.WaitVisible("#hero"),
		chromedp.Evaluate(`document.body.innerText`, &bodyText),
	)
	require.NoError(t, err)

	assert.Contains(t, bodyText, "PERSONAL",
		"Section endpoint should contain expected content")
}

func TestE2ESectionsPresent(t *testing.T) {
	ctx, cancel := newChromedpContext(t)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
	)
	require.NoError(t, err)

	var contentChecks map[string]bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				return {
					hasTitle: document.title.includes('RezusCloud'),
					hasNav: document.querySelector('nav') !== null,
					hasFooter: document.querySelector('footer') !== null,
					allSectionsPresent: ['hero', 'architecture', 'live', 'features', 'getstarted']
						.every(id => document.getElementById(id) !== null)
				};
			})()
		`, &contentChecks),
	)
	require.NoError(t, err)

	assert.True(t, contentChecks["hasTitle"], "Page should have title")
	assert.True(t, contentChecks["hasNav"], "Page should have nav")
	assert.True(t, contentChecks["hasFooter"], "Page should have footer")
	assert.True(t, contentChecks["allSectionsPresent"], "All 5 sections should be present")
}
