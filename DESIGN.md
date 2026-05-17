---
name: RezusCloud Platform Website
description: "Dual-era retro computing visual system. Mac System 1 (1984) for light mode, NeXTSTEP (1988) with teal accent for dark mode. 5 sections: hero, architecture, live platform, features, get started."
colors:
  # Mac mode: warm-tinted neutrals + amber gold accent
  paper: "oklch(99.5% 0.004 85)"
  surface: "oklch(95.5% 0.004 85)"
  surface-strong: "oklch(88.5% 0.006 85)"
  ink: "oklch(14% 0.008 65)"
  ink-muted: "oklch(30% 0.008 65)"
  rule: "oklch(72% 0.005 85)"
  accent-gold: "oklch(78% 0.16 75)"
  accent-gold-dark: "oklch(65% 0.14 75)"
  # NeXT mode: cool-tinted grayscale + teal accent
  next-black: "oklch(6% 0.005 270)"
  next-dark: "oklch(20% 0.006 270)"
  next-mid: "oklch(34% 0.006 270)"
  next-light: "oklch(58% 0.006 270)"
  next-white: "oklch(88% 0.004 270)"
  next-subtle: "oklch(72% 0.012 85)"
  next-teal: "oklch(60% 0.08 170)"
  next-teal-hi: "oklch(70% 0.07 170)"
  # Bevel system
  next-bevel-hi: "oklch(55% 0.006 270)"
  next-bevel-lo: "oklch(2% 0.004 270)"
  # Semantic
  positive: "oklch(55% 0.12 150)"
  positive-next: "oklch(75% 0.14 150)"
  negative: "oklch(50% 0.12 25)"
  negative-next: "oklch(65% 0.14 25)"
typography:
  display:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 800
    fontSize: "clamp(3rem, 8vw, 6rem)"
    lineHeight: 0.88
    letterSpacing: "-0.05em"
  headline:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    fontSize: "1.875rem"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    fontSize: "1.125rem"
    letterSpacing: "normal"
  body:
    fontFamily: "system-ui, -apple-system, Segoe UI, sans-serif"
    fontWeight: 400
    fontSize: "0.875rem"
    lineHeight: 1.625
  label:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    fontSize: "0.75rem"
    letterSpacing: "0.1em"
  mono:
    fontFamily: "VT323, Courier New, monospace"
    fontWeight: 400
    fontSize: "0.75rem"
rounded:
  none: "0px"
spacing:
  section-hero: "5rem 4rem"
  section-md: "4rem 6rem"
  section-lg: "5rem 7rem"
  section-sm: "3rem 4rem"
  nav-px: "0.75rem"
  nav-py: "0.5rem"
  btn-px: "1.5rem"
  btn-py: "0.75rem"
components:
  nav-link-active:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    padding: "{spacing.nav-px} {spacing.nav-py}"
  nav-link-active-dark:
    backgroundColor: "{colors.next-light}"
    textColor: "{colors.next-black}"
    padding: "{spacing.nav-px} {spacing.nav-py}"
  nav-link-inactive:
    textColor: "{colors.ink-muted}"
    padding: "{spacing.nav-px} {spacing.nav-py}"
  nav-link-inactive-dark:
    textColor: "{colors.next-subtle}"
    padding: "{spacing.nav-px} {spacing.nav-py}"
  cta-button:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.paper}"
    padding: "{spacing.btn-px} {spacing.btn-py}"
  cta-button-dark:
    backgroundColor: "{colors.next-teal-hi}"
    textColor: "{colors.next-black}"
    padding: "{spacing.btn-px} {spacing.btn-py}"
  icon-square-md:
    size: "2.5rem"
    backgroundColor: "{colors.accent-gold}"
  icon-square-md-dark:
    size: "2.5rem"
    backgroundColor: "{colors.next-teal-hi}"
  icon-square-lg:
    size: "3rem"
    backgroundColor: "{colors.accent-gold}"
  icon-square-lg-dark:
    size: "3rem"
    backgroundColor: "{colors.next-teal-hi}"
  accent-bar:
    width: "3rem"
    height: "0.25rem"
    backgroundColor: "{colors.accent-gold}"
  accent-bar-dark:
    width: "3rem"
    height: "0.25rem"
    backgroundColor: "{colors.next-teal-hi}"
  accent-dot:
    size: "0.375rem"
    backgroundColor: "{colors.accent-gold}"
  accent-dot-dark:
    size: "0.375rem"
    backgroundColor: "{colors.next-teal-hi}"
  service-box:
    minWidth: "5.625rem"
    padding: "0.375rem"
  service-box-self:
    minWidth: "5.625rem"
    padding: "0.375rem"
---

# Design System: RezusCloud

## 1. Overview

**Creative North Star: "The Machine Room"**

Two operating systems, one machine. Light mode is a 1984 Macintosh 128K: bitmap fonts, 1px borders, stark contrast, amber highlights on warm white. Dark mode is a 1988 NeXTcube: 3D beveled edges, CRT scanlines, phosphor afterglow, muted teal on deep black. The site is not decorated to look retro. It IS retro, running natively in the browser.

The system is defiant, precise, and nostalgic. Every element earns its place. No rounded corners, no gradient text, no glassmorphism, no SaaS hero-metric template, no identical card grids. The layout principle is flat: text flows without container boxes, spacing varies for rhythm, and section backgrounds alternate between two surface levels. Icon squares and accent dots punctuate the hierarchy. Terminals use VT323 monospace with typewriter line reveals.

Five sections tell the story: an asymmetric hero with a boot-sequence terminal, three clean architecture columns, a self-observing live platform grid, four core promises with a Machine Room closer, and an actionable Get Started with GitHub CTA. The page rhythm is left-anchored hero, then centered symmetry through the remaining sections.

For the brand strategy, storytelling context, and anti-references behind this visual system, see PRODUCT.md. For domain terminology (Personal Cloud, Machine Room, Golden Path, Builder), see CONTEXT.md.

**Key Characteristics:**
- Zero border-radius everywhere. Every corner is a right angle.
- Dual-era identity: Mac System 1 in light, NeXTSTEP in dark. Fonts stay the same, only colors change.
- Committed color strategy: amber gold (light) / teal (dark) carry the accent load across 30-40% of visible surface.
- Flat text layout: no boxes wrapping content, no bordered cards. Icon squares, dots, and accent bars punctuate.
- 3D bevel system for NeXT mode: 2px raised/sunken borders on interactive elements and signature components.
- CRT overlay: scanline pattern at 2px intervals with slow vertical scroll animation.
- Bitmap typography: Silkscreen for headings, labels, and nav. system-ui for body. VT323 for terminals.
- All animations respect `prefers-reduced-motion`. FOUC prevention via inline script.
- 5 sections: Hero, Architecture, Live Platform, Features, Get Started.
- Live Platform: collapsible category-first grid with SSE real-time metrics. Self-namespace category auto-expands.
- Progressive disclosure: Live section categories collapsed by default, expand on click.

## 2. Colors

Two complete palettes, each with warm-tinted neutrals and one saturated accent. The light mode uses amber-direction hue (85 degrees) for neutrals. The dark mode uses cool-direction hue (270 degrees) for neutrals and teal (hue 170) for the accent.

### Mac Mode (Light)

- **Paper** (oklch(99.5% 0.004 85)): Primary background. Warm white, not pure white. Used for body, Features, GetStarted sections, active nav pills.
- **Surface** (oklch(95.5% 0.004 85)): Alternating section background. Architecture section.
- **Surface-strong** (oklch(88.5% 0.006 85)): Chrome surfaces. Nav bar, footer, Live Platform section.
- **Ink** (oklch(14% 0.008 65)): Primary text. Near-black with warm undertone. Headings, body copy, active states.
- **Ink-muted** (oklch(30% 0.008 65)): Secondary text. Warm brown for descriptions, labels, inactive nav. Scoped to decorative/secondary use only.
- **Rule** (oklch(72% 0.005 85)): Borders and dividers. Warm gray for 1px borders, section separators.
- **Amber Gold** (oklch(78% 0.16 75)): The one accent. Icon square backgrounds, accent bars, terminal prompts ($ and *), logo "Cloud" highlight, hero "YOUR", hero subtitle emphasis, category labels, CTA buttons (hover), service box borders (hover). Carries the brand identity.

### NeXT Mode (Dark)

- **Next-black** (oklch(6% 0.005 270)): Deepest background. Body, Features, GetStarted, Live Platform sections.
- **Next-dark** (oklch(20% 0.006 270)): Secondary background. Architecture section, nav bar, footer, host header cells.
- **Next-mid** (oklch(34% 0.006 270)): Hover state background. Nav link hover, button hover, terminal title bar (dark).
- **Next-light** (oklch(58% 0.006 270)): Bright neutral for active nav pill backgrounds. NOT used as accent.
- **Next-white** (oklch(88% 0.004 270)): Primary text. Cool white for headings, body copy, service metrics.
- **Next-subtle** (oklch(72% 0.012 85)): Secondary text. Warm-tinted beige (hue 85) for descriptions, labels, legend text.
- **Next-teal** (oklch(60% 0.08 170)): Accent for text and borders. Logo "Cloud", terminal prompts, hero "YOUR", category labels, service detail panel headers, status indicators. The NeXT Inc. brand color.
- **Next-teal-hi** (oklch(70% 0.07 170)): Accent for backgrounds. Icon squares, dots, accent bars, CTA button fill, self-referencing service box borders.

### Bevel System

- **Bevel-hi** (oklch(55% 0.006 270)): Top-left border color for raised elements, bottom-right for sunken.
- **Bevel-lo** (oklch(2% 0.004 270)): Bottom-right border color for raised, top-left for sunken.

### Semantic Colors

- **Positive** (oklch(55% 0.12 150) / oklch(75% 0.14 150)): Success states, checkmarks in terminal boot sequence.
- **Negative** (oklch(50% 0.12 25) / oklch(65% 0.14 25)): Error states, connection lost banners.

### Named Rules

**The Dual-Accent Rule.** Light mode uses amber gold. Dark mode uses teal. Never mix: amber never appears in dark mode, teal never appears in light mode. Each accent occupies the same structural positions in both modes.

**The Alternating Rhythm Rule.** Sections alternate between two background levels. Light: paper (hero, features, getstarted) and surface/surface-strong (architecture, live). Dark: next-black (hero, features, getstarted, live) and next-dark (architecture, nav, footer). This rhythm never breaks.

**The No-Gray-Accent Rule.** In NeXT mode, next-light is a bright neutral for UI chrome (active nav pills only). It is never used for accent purposes like icon squares, dots, or highlights. That role belongs exclusively to next-teal.

## 3. Typography

**Display Font:** Silkscreen (bitmap pixel font, system-ui fallback)
**Body Font:** system-ui (-apple-system, Segoe UI fallback)
**Mono Font:** VT323 (Courier New fallback, ASCII subset: U+0020-007F, U+2713)

**Character:** Three fonts, three voices. Silkscreen is the machine speaking in labels and headings. system-ui is the human reading documentation. VT323 is the terminal: raw, monospaced, alive with blinking cursors.

### Hierarchy

- **Display** (800 weight, text-5xl to text-8xl, leading-[0.88], tracking-tighter): Hero headline only. "YOUR PERSONAL CLOUD" as the single most prominent element on the page.
- **Headline** (700 weight, text-2xl to text-4xl, tracking-tight): Section headings ("How It Works", "Live Platform", "What You Get", "Start Your Cloud", "Your Machine Room"). Centered for all sections except hero.
- **Title** (700 weight, text-lg, leading-snug): Feature names ("No Platform Tax", "Private by Design"), sub-section headings.
- **Body** (400 weight, text-sm to text-lg, leading-relaxed): Paragraph copy, descriptions, feature details. Max line length 65-75ch.
- **Label** (700 weight, text-xs, tracking-widest, uppercase): Nav links, category counts, host detail field labels.
- **Mono** (400 weight, text-[10px] to text-sm): Terminal output, service metrics (CPU/RAM values), host detail values, legend text, category headers. VT323 subset to ASCII range for file size.

### Named Rules

**The No-Serif Rule.** Serif fonts never appear. The system is entirely sans-serif + bitmap + mono.

**The Font-Consistency Rule.** Font families are identical between light and dark modes. Only colors change. Silkscreen always for display/labels, system-ui always for body, VT323 always for terminals.

## 4. Elevation

Mac mode is flat. No shadows, no bevels. Depth comes from background color alternation (paper vs surface vs surface-strong) and border rules. The single exception is the `mac-window-shadow` utility (1px 1px 0 ink) applied to terminal containers.

NeXT mode uses a 2px 3D bevel system. The bevels simulate the chunky, period-authentic NeXTSTEP interface: a top-left highlight (bevel-hi at oklch 55%) and bottom-right shadow (bevel-lo at oklch 2%) create raised elements. Reversing the colors creates sunken elements. Applied to: icon squares, CTA buttons, terminal container borders, mobile nav wrapper.

### Bevel Vocabulary

- **Raised** (`border: 2px solid; border-color: bevel-hi bevel-lo bevel-lo bevel-hi`): Default for interactive containers. Icon squares, CTA buttons, terminal windows.
- **Sunken** (`border: 2px solid; border-color: bevel-lo bevel-hi bevel-hi bevel-lo`): Inset fields. Terminal container in dark mode.
- **Hover raised** (`next-bevel-hover`): Brightens bevel-hi on hover for interactive feedback.

### Overlay Effects

- **CRT Scanlines** (`next-scanlines`): 2px horizontal lines at 8% white opacity, 4px intervals. Slow vertical scroll animation (8s linear infinite). Fixed overlay on entire viewport in dark mode only.
- **CRT Flicker**: White overlay flashes on theme toggle (step-end timing, 200ms total). Three opacity pulses: 0.8, 1, 0, creating a power-cycle illusion.
- **Mac Dot Grid** (`mac-dots`): Radial gradient dots at 12px intervals. Parallax-shifted on mouse move (throttled 150ms). 4% opacity in light mode, hidden in dark mode.
- **Phosphor Fade** (`phosphor-fade`): Text-shadow glow on typed terminal lines (6px white shadow fades to none over 2.5s). CRT authenticity detail.

### Named Rules

**The Flat-By-Mode Rule.** Mac mode is always flat: borders are always 1px solid ink. NeXT mode uses bevels: borders are always 2px with directional shading. Never mix bevel styles within a mode.

## 5. Components

### Navigation

Fixed top bar, z-50. `surface-strong` (light) / `next-dark` (dark) background. 1px bottom border (`rule` / `next-mid`). Logo: icon square (32x32) + "Rezus" in ink/next-white + "Cloud" in accent-gold/next-teal. Desktop links: Silkscreen label weight, uppercase, 4 items (Home, Architecture, Features, Get Started). Active state: pill background (paper/next-light). Hover: surface/next-mid background. Mobile: hamburger menu with slide-down panel (200ms ease-out). IntersectionObserver tracks active section via scroll position.

### CTA Button

Inline-flex. Background: ink (light) / next-teal-hi (dark). Text: paper (light) / next-black (dark). Font: Silkscreen 700. 1px border ink (light) / 2px bevel raised (dark). Hover: accent-gold background, ink text (light) / next-mid background, next-black text (dark). Arrow icon (right-pointing chevron) with 16px size. 150ms color transition. Used for hero "Get Started" and getstarted "Star on GitHub" CTAs.

### Secondary Button

Inline-flex. Transparent background. 1px border: ink (light) / next-mid (dark). Text: ink (light) / next-white (dark). Font: Silkscreen 700. Hover: surface/next-mid background. Used for getstarted "Watch It Live" link.

### Icon Squares

Fixed-size squares. Two sizes: w-12 h-12 (48px, architecture columns, hero counter) and w-10 h-10 (40px, feature items, footer logo). Background: accent-gold (light) / next-teal-hi (dark). 1px border (light) / 2px bevel raised (dark). Contain SVG icons in ink/next-black color. Never contain text paragraphs.

### Accent Dots

Small circles (w-1.5 h-1.5 = 6px). Background: accent-gold (light) / next-teal-hi (dark). Used as list bullet alternatives in Machine Room capabilities grid and feature dot-lists.

### Accent Bars

Horizontal bars (w-12 h-1 = 48x4px). Background: accent-gold (light) / next-teal-hi (dark). Centered as section heading underlines (Features, GetStarted, Machine Room) or left-aligned (hero sub-headline).

### Terminal Window

Signature component. Dark background container (bg-next-black in both modes). Title bar with static diagonal barber pole stripes (repeating-linear-gradient at -45deg, alternating 2px bands). Three colored squares in light mode (mac-window title bar). VT323 monospace font. Blinking cursor via CSS step-end animation. Line-by-line typewriter reveal with configurable delays (base-delay + line-delay per line). Phosphor glow fade on typed lines. `prefers-reduced-motion` shows all lines immediately. Used in hero (rezusctl boot sequence) and getstarted (git clone sequence).

### Live Platform Grid

Category-first CSS grid (`gap-x-6 gap-y-0`). Hosts as columns, categories as rows. Each host has a clickable header button with toggle arrow (expand/collapse node details). Category rows have clickable headers with service count and expand/collapse arrow. Service boxes (min 90px) show status dot + CPU cores (3 decimals, "c" suffix) + RAM (megabytes, "M" suffix). Self-referencing service ("THIS SITE") gets accent-colored border and label. Categories collapsed by default (progressive disclosure); self-namespace category auto-expands. SSE stream updates metrics every 5 seconds. Detail panels show 4 sparkline charts (CPU, RAM, Network, Disk).

### Feature Items

4 core promises in 2-column grid (md:grid-cols-2). Each has: icon square (40px) + title (text-lg, Silkscreen bold) + description (text-sm, system-ui). No borders, no backgrounds. Icon square is the only colored element. Below: Machine Room conclusion with accent bar divider, 8 capability items in 2-column grid (period-terminated labels + muted descriptions), and a closing manifesto quote in accent color.

### Footer

`surface-strong` (light) / `next-dark` (dark) background. Top border. Logo with icon square. "Platform" and "Explore" link groups with internal anchors (Features, Architecture, Get Started, Live Platform). Terminal-style copyright line: "MODEL RC-001 // SERIAL: OPEN-SOURCE // Built with Go, templ, HTMX, Alpine.js, Tailwind CSS" with accent-colored tech stack mention.

### Skip Link

`sr-only` by default. On focus: absolute positioned, accent-gold background, ink text, z-60, Silkscreen font. Keyboard accessible skip to main content.

### Theme Toggle

40x40px button. Border + bevel raised (dark). Shows moon icon (light) / sun icon (dark). Alpine.js x-show toggles visibility with x-cloak for no-flash. CRT flicker overlay on toggle (200ms white flash with 3-step opacity animation).

## 6. Do's and Don'ts

### Do:
- **Do** use zero border-radius on every element. Every corner is a right angle. This is non-negotiable.
- **Do** use Silkscreen for all headings, labels, and nav. system-ui for body. VT323 for terminals only.
- **Do** alternate section backgrounds. Light: paper/surface/surface-strong. Dark: next-black/next-dark. The pattern is: hero (paper/next-black), architecture (surface/next-dark), live (surface-strong/next-black), features (paper/next-black), getstarted (paper/next-black).
- **Do** use amber gold (light) / teal (dark) as the exclusive accent. One accent per mode, same structural positions.
- **Do** use 2px bevels in NeXT mode for interactive elements. They are the elevation system.
- **Do** respect `prefers-reduced-motion`. All animations must have a static fallback.
- **Do** use warm-tinted neutrals. Paper and ink have hue 85 (amber direction). Next-black and next-white have hue 270 (cool direction).
- **Do** use next-subtle (warm-tinted at hue 85) for dark mode secondary text. The two-color text system (cool white + warm beige) mirrors light mode (black + brown).
- **Do** vary section padding for rhythm. pt-20/py-16 (hero), py-16 (architecture, live), py-20 (features), py-12 (getstarted).
- **Do** collapse live section categories by default. Progressive disclosure reduces cognitive load from 65 service boxes to ~10 category rows.
- **Do** add "THIS SITE" label to the self-referencing service box. It proves the dogfooding story.
- **Do** provide actionable CTAs. GitHub star button and "Watch It Live" link, not "Coming soon" placeholders.
- **Do** reference CONTEXT.md for domain terminology when writing copy.

### Don't:
- **Don't** use rounded corners anywhere. No `rounded-*` classes. No `border-radius` in CSS.
- **Don't** use gradient text (`background-clip: text`). Solid colors only.
- **Don't** use glassmorphism. No blur overlays, no frosted glass cards.
- **Don't** wrap text in bordered container boxes. Layout is flat: text flows with icon squares and accent dots for punctuation.
- **Don't** mix accent colors between modes. Amber never in dark, teal never in light.
- **Don't** use `#000` or `#fff`. All neutrals are tinted with brand hue. Minimum chroma 0.004.
- **Don't** use next-light as an accent color. It is a bright neutral for active nav pills only. Teal is the accent.
- **Don't** animate CSS layout properties. Animate opacity and transform only.
- **Don't** use bounce, elastic, or spring easings. Use ease-out-quart or exponential curves.
- **Don't** use em dashes in copy. Use commas, colons, semicolons, periods, or parentheses.
- **Don't** show all 65 service boxes at once. Use progressive disclosure with collapsible categories.
- **Don't** use "Coming soon" as the only CTA. Always provide at least one actionable next step.
- **Don't** create platforms designed for procurement committees and sold via sales calls. This platform gives engineers enterprise capabilities directly.
- **Don't** use trust logos, monthly pricing tables, or "Get Started Free" CTAs. This is not a subscription product.

For brand-level anti-references (what this product is NOT), see PRODUCT.md.
