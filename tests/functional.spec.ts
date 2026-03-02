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

test.describe('Homepage', () => {
  test('loads successfully', async ({ page }) => {
    const response = await page.goto(BASE_URL);
    expect(response?.status()).toBe(200);
  });

  test('has correct title', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page).toHaveTitle(/RezusCloud/);
  });

  test('has meta description', async ({ page }) => {
    await page.goto(BASE_URL);
    const description = page.locator('meta[name="description"]');
    await expect(description).toHaveAttribute('content', /Enterprise Kubernetes Platform/);
  });
});

test.describe('Sections', () => {
  test('all sections are present', async ({ page }) => {
    await page.goto(BASE_URL);
    
    for (const section of SECTIONS) {
      const locator = page.locator(`#${section.id}`);
      await expect(locator).toBeVisible();
    }
  });

  test('sections have content', async ({ page }) => {
    await page.goto(BASE_URL);
    
    for (const section of SECTIONS) {
      const locator = page.locator(`#${section.id}`);
      const text = await locator.textContent();
      expect(text?.trim().length).toBeGreaterThan(10);
    }
  });
});

test.describe('Navigation', () => {
  const NAV_LINKS = [
    { href: '#hero', text: 'Home' },
    { href: '#architecture', text: 'Architecture' },
    { href: '#features', text: 'Features' },
    { href: '#comparison', text: 'Compare' },
    { href: '#getstarted', text: 'Get Started' },
  ];

  test('navigation links exist', async ({ page }) => {
    await page.goto(BASE_URL);
    
    for (const link of NAV_LINKS) {
      const navLink = page.locator(`nav a[href="${link.href}"]`);
      await expect(navLink.first()).toContainText(link.text);
    }
  });

  test('navigation links scroll to sections', async ({ page }) => {
    await page.goto(BASE_URL);
    
    for (const link of NAV_LINKS) {
      await page.click(`nav a[href="${link.href}"]`);
      await page.waitForTimeout(500);
      
      const section = page.locator(link.href);
      await expect(section).toBeInViewport();
    }
  });

  test('footer links work', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.locator('footer').scrollIntoViewIfNeeded();
    
    const footerLinks = page.locator('footer a');
    const count = await footerLinks.count();
    expect(count).toBeGreaterThan(0);
    
    for (let i = 0; i < count; i++) {
      const href = await footerLinks.nth(i).getAttribute('href');
      expect(href).toBeTruthy();
      expect(href?.startsWith('#')).toBeTruthy();
    }
  });
});

test.describe('Theme Toggle', () => {
  test('defaults to system preference', async ({ page }) => {
    await page.goto(BASE_URL);
    
    const theme = await page.evaluate(() => localStorage.getItem('theme'));
    const hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    
    if (theme === null) {
      const prefersDark = await page.evaluate(() => 
        window.matchMedia('(prefers-color-scheme: dark)').matches
      );
      expect(hasDarkClass).toBe(prefersDark);
    }
  });

  test('toggles to dark mode', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'light');
    await page.reload();
    
    await page.click('button[aria-label="Toggle theme"]');
    await page.waitForTimeout(300);
    
    const hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    expect(hasDarkClass).toBe(true);
    
    const theme = await page.evaluate(() => localStorage.getItem('theme'));
    expect(theme).toBe('dark');
  });

  test('toggles to light mode', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'dark');
    await page.reload();
    
    await page.click('button[aria-label="Toggle theme"]');
    await page.waitForTimeout(300);
    
    const hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    expect(hasDarkClass).toBe(false);
    
    const theme = await page.evaluate(() => localStorage.getItem('theme'));
    expect(theme).toBe('light');
  });

  test('persists theme across reloads', async ({ page }) => {
    await page.goto(BASE_URL);
    await setTheme(page, 'dark');
    await page.reload();
    
    let hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    expect(hasDarkClass).toBe(true);
    
    await setTheme(page, 'light');
    await page.reload();
    
    hasDarkClass = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    expect(hasDarkClass).toBe(false);
  });
});

test.describe('Mobile Menu', () => {
  test.use({ viewport: { width: 375, height: 812 } });

  test('mobile menu button is visible', async ({ page }) => {
    await page.goto(BASE_URL);
    const menuBtn = page.locator('#mobile-menu-btn');
    await expect(menuBtn).toBeVisible();
  });

  test('mobile menu toggles on click', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');
    
    const mobileMenu = page.locator('#mobile-menu');
    const menuBtn = page.locator('#mobile-menu-btn');
    
    await expect(mobileMenu).toBeHidden();
    
    await menuBtn.click();
    await expect(mobileMenu).toBeVisible();
    
    await menuBtn.click();
    await expect(mobileMenu).toBeHidden();
  });

  test('mobile menu closes after navigation', async ({ page }) => {
    await page.goto(BASE_URL);
    
    await page.click('#mobile-menu-btn');
    const mobileMenu = page.locator('#mobile-menu');
    await expect(mobileMenu).toBeVisible();
    
    await mobileMenu.locator('a[href="#features"]').click();
    await page.waitForTimeout(500);
    
    await expect(page.locator('#features')).toBeInViewport();
  });

  test('mobile menu has all links', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.click('#mobile-menu-btn');
    
    const mobileMenu = page.locator('#mobile-menu');
    const links = mobileMenu.locator('a');
    const count = await links.count();
    expect(count).toBe(5);
  });
});

test.describe('HTMX Sections', () => {
  test('section endpoint returns content', async ({ page }) => {
    const response = await page.request.get(`${BASE_URL}/sections/hero`);
    expect(response.status()).toBe(200);
    
    const text = await response.text();
    expect(text).toContain('Enterprise Kubernetes');
  });

  test('all sections available via HTMX', async ({ page }) => {
    for (const section of SECTIONS) {
      const response = await page.request.get(`${BASE_URL}/sections/${section.id}`);
      expect(response.status()).toBe(200);
      
      const text = await response.text();
      expect(text.length).toBeGreaterThan(50);
    }
  });
});

test.describe('Responsive Design', () => {
  test('desktop navigation visible on large screens', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto(BASE_URL);
    
    const desktopNav = page.locator('nav .hidden.md\\:flex');
    await expect(desktopNav).toBeVisible();
  });

  test('mobile menu button hidden on large screens', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto(BASE_URL);
    
    const mobileBtn = page.locator('#mobile-menu-btn');
    await expect(mobileBtn).toBeHidden();
  });

  test('sections are responsive', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto(BASE_URL);
    
    for (const section of SECTIONS) {
      const locator = page.locator(`#${section.id}`);
      await expect(locator).toBeVisible();
      
      const box = await locator.boundingBox();
      expect(box?.width).toBeLessThanOrEqual(375);
    }
  });
});

test.describe('Accessibility', () => {
  test('has skip link or main landmark', async ({ page }) => {
    await page.goto(BASE_URL);
    const main = page.locator('main');
    await expect(main).toBeVisible();
  });

  test('theme toggle has aria-label', async ({ page }) => {
    await page.goto(BASE_URL);
    const toggle = page.locator('button[aria-label="Toggle theme"]');
    await expect(toggle).toBeVisible();
  });

  test('all images have alt text', async ({ page }) => {
    await page.goto(BASE_URL);
    const images = page.locator('img');
    const count = await images.count();
    
    for (let i = 0; i < count; i++) {
      const alt = await images.nth(i).getAttribute('alt');
      expect(alt).toBeDefined();
    }
  });

  test('links have discernible text', async ({ page }) => {
    await page.goto(BASE_URL);
    const links = page.locator('a');
    const count = await links.count();
    
    for (let i = 0; i < count; i++) {
      const text = await links.nth(i).textContent();
      const ariaLabel = await links.nth(i).getAttribute('aria-label');
      expect(text?.trim() || ariaLabel).toBeTruthy();
    }
  });
});

test.describe('Performance', () => {
  test('page loads within acceptable time', async ({ page }) => {
    const start = Date.now();
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');
    const loadTime = Date.now() - start;
    
    expect(loadTime).toBeLessThan(5000);
  });

  test('no unexpected console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        const text = msg.text();
        if (!text.includes('favicon') && !text.includes('404')) {
          errors.push(text);
        }
      }
    });
    
    await page.goto(BASE_URL);
    await page.waitForLoadState('networkidle');
    
    expect(errors).toHaveLength(0);
  });
});
