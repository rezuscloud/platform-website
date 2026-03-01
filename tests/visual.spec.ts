import { test, expect, Page } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';

const SECTIONS = [
  { id: 'hero', name: 'Hero' },
  { id: 'challenge', name: 'Challenge' },
  { id: 'architecture', name: 'Architecture' },
  { id: 'features', name: 'Features' },
  { id: 'networking', name: 'Networking' },
  { id: 'edge', name: 'Edge' },
  { id: 'services', name: 'Services' },
  { id: 'comparison', name: 'Comparison' },
  { id: 'usecases', name: 'Use Cases' },
  { id: 'techstack', name: 'Tech Stack' },
  { id: 'getstarted', name: 'Get Started' },
];

const VIEWPORTS = {
  desktop: { width: 1280, height: 720 },
  mobile: { width: 375, height: 812 },
};

const THEMES = ['light', 'dark'] as const;

async function setTheme(page: Page, theme: 'light' | 'dark') {
  await page.evaluate((t) => {
    if (t === 'dark') {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  }, theme);
}

async function waitForStable(page: Page) {
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500);
}

test.describe('Visual Regression Tests', () => {
  for (const [viewportName, viewport] of Object.entries(VIEWPORTS)) {
    for (const theme of THEMES) {
      test.describe(`${viewportName} - ${theme} theme`, () => {
        test.use({ viewport });

        test.beforeEach(async ({ page }) => {
          await page.goto(BASE_URL);
          await setTheme(page, theme);
          await page.reload();
          await waitForStable(page);
        });

        for (const section of SECTIONS) {
          test(`${section.name} section`, async ({ page }) => {
            const sectionLocator = page.locator(`#${section.id}`);
            await sectionLocator.scrollIntoViewIfNeeded();
            await page.waitForTimeout(300);
            
            await expect(sectionLocator).toHaveScreenshot(
              `${section.id}-${viewportName}-${theme}.png`,
              {
                maxDiffPixels: 100,
                threshold: 0.1,
              }
            );
          });
        }

        test('full page', async ({ page }) => {
          await expect(page).toHaveScreenshot(
            `fullpage-${viewportName}-${theme}.png`,
            {
              fullPage: true,
              maxDiffPixels: 500,
              threshold: 0.1,
            }
          );
        });
      });
    }
  }
});

test.describe('Navigation Visual Tests', () => {
  test.use({ viewport: VIEWPORTS.desktop });

  test('navigation bar - light theme', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'light');
    await page.reload();
    await waitForStable(page);

    const nav = page.locator('nav');
    await expect(nav).toHaveScreenshot('nav-desktop-light.png', {
      maxDiffPixels: 50,
    });
  });

  test('navigation bar - dark theme', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'dark');
    await page.reload();
    await waitForStable(page);

    const nav = page.locator('nav');
    await expect(nav).toHaveScreenshot('nav-desktop-dark.png', {
      maxDiffPixels: 50,
    });
  });
});

test.describe('Footer Visual Tests', () => {
  test.use({ viewport: VIEWPORTS.desktop });

  test('footer - light theme', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'light');
    await page.reload();
    await page.locator('footer').scrollIntoViewIfNeeded();
    await waitForStable(page);

    const footer = page.locator('footer');
    await expect(footer).toHaveScreenshot('footer-desktop-light.png', {
      maxDiffPixels: 100,
    });
  });

  test('footer - dark theme', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'dark');
    await page.reload();
    await page.locator('footer').scrollIntoViewIfNeeded();
    await waitForStable(page);

    const footer = page.locator('footer');
    await expect(footer).toHaveScreenshot('footer-desktop-dark.png', {
      maxDiffPixels: 100,
    });
  });
});

test.describe('Mobile Menu Visual Tests', () => {
  test.use({ viewport: VIEWPORTS.mobile });

  test('mobile menu closed', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'light');
    await page.reload();
    await waitForStable(page);

    const nav = page.locator('nav');
    await expect(nav).toHaveScreenshot('mobile-nav-closed.png', {
      maxDiffPixels: 50,
    });
  });

  test('mobile menu open', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'light');
    await page.reload();
    await waitForStable(page);

    await page.click('#mobile-menu-btn');
    await page.waitForTimeout(300);

    const nav = page.locator('nav');
    await expect(nav).toHaveScreenshot('mobile-nav-open.png', {
      maxDiffPixels: 50,
    });
  });
});
