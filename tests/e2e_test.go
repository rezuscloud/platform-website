//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
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

func TestE2EHomePageLoads(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
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

func TestE2EAllSectionsExist(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	sections := []string{
		"hero", "challenge", "architecture", "features",
		"networking", "edge", "services", "comparison",
		"usecases", "techstack", "getstarted",
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("#hero"),
	)
	require.NoError(t, err)

	for _, section := range sections {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.ScrollIntoView("#"+section),
			chromedp.Sleep(100*time.Millisecond),
			chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					const el = document.getElementById('%s');
					return el !== null;
				})()
			`, section), &exists),
		)
		require.NoError(t, err)
		assert.True(t, exists, "Section %s should exist", section)
	}
}

func TestE2EThemeToggle(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
	)
	require.NoError(t, err)

	var initialHasDark bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.documentElement.classList.contains('dark')`, &initialHasDark),
	)
	require.NoError(t, err)

	err = chromedp.Run(ctx,
		chromedp.Click("button[aria-label='Toggle theme']"),
		chromedp.Sleep(300*time.Millisecond),
	)
	require.NoError(t, err)

	var afterToggleHasDark bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.documentElement.classList.contains('dark')`, &afterToggleHasDark),
	)
	require.NoError(t, err)

	assert.NotEqual(t, initialHasDark, afterToggleHasDark,
		"Theme should toggle between light and dark")
}

func TestE2ENavigationScroll(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	navLinks := []struct {
		selector    string
		target      string
		description string
	}{
		{"nav a[href='#architecture']", "#architecture", "Architecture"},
		{"nav a[href='#features']", "#features", "Features"},
		{"nav a[href='#getstarted']", "#getstarted", "Get Started"},
	}

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("nav"),
	)
	require.NoError(t, err)

	for _, link := range navLinks {
		var isInViewport bool
		err := chromedp.Run(ctx,
			chromedp.Click(link.selector),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					const el = document.querySelector('%s');
					const rect = el.getBoundingClientRect();
					return rect.top >= 0 && rect.top < window.innerHeight;
				})()
			`, link.target), &isInViewport),
		)
		require.NoError(t, err)
		assert.True(t, isInViewport, "%s section should be in viewport after clicking nav", link.description)
	}
}

func TestE2EMobileMenu(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(375, 812),
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("#mobile-menu-btn"),
	)
	require.NoError(t, err)

	var menuInitiallyVisible bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				const el = document.getElementById('mobile-menu');
				if (!el) return false;
				const style = window.getComputedStyle(el);
				return style.display !== 'none' && style.visibility !== 'hidden';
			})()
		`, &menuInitiallyVisible),
	)
	require.NoError(t, err)
	assert.False(t, menuInitiallyVisible, "Mobile menu should be hidden initially")

	err = chromedp.Run(ctx,
		chromedp.Click("#mobile-menu-btn"),
		chromedp.Sleep(300*time.Millisecond),
	)
	require.NoError(t, err)

	var menuVisibleAfterClick bool
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				const el = document.getElementById('mobile-menu');
				if (!el) return false;
				const style = window.getComputedStyle(el);
				return style.display !== 'none' && style.visibility !== 'hidden';
			})()
		`, &menuVisibleAfterClick),
	)
	require.NoError(t, err)
	assert.True(t, menuVisibleAfterClick, "Mobile menu should be visible after clicking button")

	var linkCount int
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelectorAll('#mobile-menu a').length`, &linkCount),
	)
	require.NoError(t, err)
	assert.Equal(t, 5, linkCount, "Mobile menu should have 5 links")
}

func TestE2EPerformance(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
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
	ctx, cancel := chromedp.NewContext(context.Background())
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

	assert.Contains(t, bodyText, "Enterprise Kubernetes",
		"Section endpoint should contain expected content")
}

func TestE2EConsoleErrors(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var consoleErrors []string

	taskCtx, taskCancel := chromedp.NewContext(ctx)
	defer taskCancel()

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *chromedp.EventLogConsoleAPICalled:
			if ev.Type == "error" {
				for _, arg := range ev.Args {
					consoleErrors = append(consoleErrors, string(arg.Value))
				}
			}
		}
	})

	err := chromedp.Run(ctx,
		chromedp.Navigate(getBaseURL()),
		chromedp.WaitVisible("body"),
		chromedp.Sleep(2*time.Second),
	)
	require.NoError(t, err)

	filteredErrors := []string{}
	for _, errMsg := range consoleErrors {
		if !strings.Contains(errMsg, "favicon") && !strings.Contains(errMsg, "404") {
			filteredErrors = append(filteredErrors, errMsg)
		}
	}

	assert.Empty(t, filteredErrors, "Page should not have unexpected console errors")
}
