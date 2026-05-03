---
name: RezusCloud Platform Website
description: Dual-era retro computing visual system. Mac System 1 (1984) for light mode, NeXTSTEP (1988) with teal accent for dark mode.
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
  next-teal: "oklch(52% 0.08 170)"
  next-teal-hi: "oklch(62% 0.07 170)"
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
    fontWeight: 700
  headline:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    letterSpacing: "normal"
  title:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    letterSpacing: "normal"
  body:
    fontFamily: "system-ui, -apple-system, Segoe UI, sans-serif"
    fontWeight: 400
  label:
    fontFamily: "Silkscreen, system-ui, sans-serif"
    fontWeight: 700
    letterSpacing: "0.05em"
  mono:
    fontFamily: "VT323, Courier New, monospace"
    fontWeight: 400
rounded:
  none: "0px"
spacing:
  section-sm: "3.5rem 4rem"
  section-md: "5rem 7rem"
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
  icon-square:
    size: "3rem"
    backgroundColor: "{colors.accent-gold}"
  icon-square-dark:
    size: "3rem"
    backgroundColor: "{colors.next-teal-hi}"
  accent-dot:
    size: "0.5rem"
    backgroundColor: "{colors.accent-gold}"
  accent-dot-dark:
    size: "0.5rem"
    backgroundColor: "{colors.next-teal-hi}"
---

# Design System: RezusCloud

## 1. Overview

**Creative North Star: "The Machine Room"**

Two operating systems, one machine. Light mode is a 1984 Macintosh 128K: bitmap fonts, 1px borders, stark contrast, amber highlights on warm white. Dark mode is a 1988 NeXTcube: 3D beveled edges, CRT scanlines, phosphor afterglow, muted teal on deep black. The site is not decorated to look retro. It IS retro, running natively in the browser.

The system is defiant, precise, and nostalgic. Every element earns its place. No rounded corners, no gradient text, no glassmorphism, no SaaS hero-metric template, no identical card grids. The layout principle is flat: text flows without container boxes, spacing varies for rhythm, and section backgrounds alternate between two surface levels. Icon squares and accent dots punctuate the hierarchy. Terminals use VT323 monospace with typewriter line reveals.

For the brand strategy, storytelling context, and anti-references behind this visual system, see PRODUCT.md.

**Key Characteristics:**
- Zero border-radius everywhere. Every corner is a right angle.
- Dual-era identity: Mac System 1 in light, NeXTSTEP in dark. Fonts stay the same, only colors change.
- Committed color strategy: amber gold (light) / teal (dark) carry the accent load across 30-40% of visible surface.
- Flat text layout: no boxes wrapping content, no bordered cards. Icon squares, dots, and accent bars punctuate.
- 3D bevel system for NeXT mode: 2px raised/sunken borders on interactive elements and signature components.
- CRT overlay: scanline pattern at 4px intervals with slow vertical scroll animation.
- Bitmap typography: Silkscreen for headings, labels, and nav. system-ui for body. VT323 for terminals.
- All animations respect `prefers-reduced-motion`. FOUC prevention via inline script.

## 2. Colors

Two complete palettes, each with warm-tinted neutrals and one saturated accent. The light mode uses amber-direction hue (85 degrees) for neutrals. The dark mode uses cool-direction hue (270 degrees) for neutrals and teal (hue 170) for the accent.

### Mac Mode (Light)

- **Paper** (oklch(99.5% 0.004 85)): Primary background. Warm white, not pure white. Used for body, most sections, active nav pills.
- **Surface** (oklch(95.5% 0.004 85)): Alternating section background. Softly differentiated from paper. Used for Architecture, Networking, UseCases.
- **Surface-strong** (oklch(88.5% 0.006 85)): Chrome surfaces. Nav bar, footer. Higher contrast than surface for fixed UI elements.
- **Ink** (oklch(14% 0.008 65)): Primary text. Near-black with warm undertone. Headings, body copy, active states.
- **Ink-muted** (oklch(30% 0.008 65)): Secondary text. Warm brown for descriptions, labels, inactive nav. Scoped to decorative/secondary use only.
- **Rule** (oklch(72% 0.005 85)): Borders and dividers. Warm gray for 1px borders, table rules, section separators.
- **Amber Gold** (oklch(78% 0.16 75)): The one accent. Icon square backgrounds, accent bars, terminal prompts ($ and *), logo "Cloud" highlight, hero "YOUR", hero subtitle, table headers. Carries the brand identity.

### NeXT Mode (Dark)

- **Next-black** (oklch(6% 0.005 270)): Deepest background. Near-black with cool undertone. Body, alternating sections (Features, Comparison, GetStarted).
- **Next-dark** (oklch(20% 0.006 270)): Secondary background. Charcoal. Architecture, Networking, UseCases, Challenge, nav bar, footer.
- **Next-mid** (oklch(34% 0.006 270)): Hover state background. Used for nav link hover, button hover.
- **Next-light** (oklch(58% 0.006 270)): Bright neutral for active nav pill backgrounds. NOT used as accent (that role belongs to next-teal).
- **Next-white** (oklch(88% 0.004 270)): Primary text. Cool white for headings, body copy.
- **Next-subtle** (oklch(72% 0.012 85)): Secondary text. Warm-tinted beige (hue 85) for descriptions, labels. Two-color text system: cool white + warm beige.
- **Next-teal** (oklch(52% 0.08 170)): Accent for text and borders. Logo "Cloud", terminal prompts, hero "YOUR", hero subtitle, table headers, badge borders. The NeXT Inc. brand color.
- **Next-teal-hi** (oklch(62% 0.07 170)): Accent for backgrounds. Icon squares, dots, accent bars, CTA button. Lighter variant for contrast on dark surfaces.

### Bevel System

- **Bevel-hi** (oklch(55% 0.006 270)): Top-left border color for raised elements, bottom-right for sunken.
- **Bevel-lo** (oklch(2% 0.004 270)): Bottom-right border color for raised, top-left for sunken.

### Named Rules

**The Dual-Accent Rule.** Light mode uses amber gold. Dark mode uses teal. Never mix: amber never appears in dark mode, teal never appears in light mode. Each accent occupies the same structural positions in both modes.

**The Alternating Rhythm Rule.** Sections alternate between two background levels. Light: paper and surface. Dark: next-black and next-dark. The challenge section inverts: ink (light) and next-dark (dark). This rhythm never breaks.

**The No-Gray-Accent Rule.** In NeXT mode, next-light is a bright neutral for UI chrome (active nav pills only). It is never used for accent purposes like icon squares, dots, or highlights. That role belongs exclusively to next-teal.

## 3. Typography

**Display Font:** Silkscreen (bitmap pixel font, system-ui fallback)
**Body Font:** system-ui (-apple-system, Segoe UI fallback)
**Mono Font:** VT323 (Courier New fallback, ASCII subset only: U+0020-007F, U+2713)

**Character:** Three fonts, three voices. Silkscreen is the machine speaking in labels and headings. system-ui is the human reading documentation. VT323 is the terminal: raw, monospaced, alive with blinking cursors.

### Hierarchy

- **Display** (700 weight, text-8xl, leading-none): Hero headline "YOUR" only. Single most prominent element on the page.
- **Headline** (700 weight, text-2xl to text-3xl, leading-snug to leading-tight): Section headings. Left-aligned for Architecture and Networking, centered for others.
- **Title** (700 weight, text-lg to text-xl, leading-snug): Sub-section headings, card titles, feature names.
- **Body** (400 weight, text-sm to text-lg, leading-relaxed): Paragraph copy, descriptions, feature details. Max line length 65-75ch.
- **Label** (700 weight, text-xs, tracking-widest, uppercase): Nav links, section badges ("1984 // Then vs Now"), feature categories, tech tags.
- **Mono** (400 weight, text-sm to text-base): Terminal output, code prompts, boot sequence lines, cursor blink.

### Named Rules

**The No-Serif Rule.** Serif fonts never appear. The system is entirely sans-serif + bitmap + mono.

**The Font-Consistency Rule.** Font families are identical between light and dark modes. Only colors change. Silkscreen always for display/labels, system-ui always for body, VT323 always for terminals.

## 4. Elevation

Mac mode is flat. No shadows, no bevels. Depth comes from background color alternation (paper vs surface) and border rules. This is the Mac System 1 aesthetic: 1px black borders on stark white.

NeXT mode uses a 2px 3D bevel system. The bevels simulate the chunky, period-authentic NeXTSTEP interface: a top-left highlight (bevel-hi at oklch 55%) and bottom-right shadow (bevel-lo at oklch 2%) create raised elements. Reversing the colors creates sunken elements. Applied to: icon squares, CTA button, tech tags, mobile nav wrapper.

### Bevel Vocabulary

- **Raised** (`border: 2px solid; border-color: bevel-hi bevel-lo bevel-lo bevel-hi`): Default for interactive containers. Icon squares, buttons, tech tags.
- **Sunken** (`border: 2px solid; border-color: bevel-lo bevel-hi bevel-hi bevel-lo`): Inset fields. Available but not currently used.
- **Hover raised** (`next-bevel-hover`): Brightens bevel-hi on hover for interactive feedback.

### Overlay Effects

- **CRT Scanlines** (`next-scanlines`): 2px horizontal lines at 8% white opacity, 4px intervals. Slow vertical scroll animation (8s linear infinite). Fixed overlay on entire viewport in dark mode.
- **CRT Flicker**: White overlay flashes on theme toggle (step-end timing, 200ms total).

### Named Rules

**The Flat-By-Mode Rule.** Mac mode is always flat: borders are always 1px solid ink. NeXT mode uses bevels: borders are always 2px with directional shading. Never mix bevel styles within a mode.

## 5. Components

### Navigation

Fixed top bar. `surface-strong` (light) / `next-dark` (dark) background. 1px bottom border. Logo: "Rezus" in ink/next-white + "Cloud" in accent-gold/next-teal. Desktop links: Silkscreen label weight, uppercase. Active state: pill background (paper/next-light). Hover: surface/next-mid background. Mobile: hamburger menu with slide-down panel.

### CTA Button

Inline-flex. Background: ink (light) / next-teal-hi (dark). Text: paper (light) / next-black (dark). Font: Silkscreen 700. 1px border ink (light) / 2px bevel raised (dark). Hover: accent-gold (light) / next-mid (dark). 150ms color transition.

### Icon Squares

Fixed-size squares (w-12 h-12 for primary, w-10 h-10 for features). Background: accent-gold (light) / next-teal-hi (dark). 1px border (light) / 2px bevel raised (dark). Contain SVG icons or bold single characters. Never contain text paragraphs.

### Accent Dots

Small circles (w-2 h-2 or w-1.5 h-1.5). Background: accent-gold (light) / next-teal-hi (dark). Used as list bullet alternatives, connection indicators, status markers.

### Accent Bars

Horizontal bars (w-12 h-1). Background: accent-gold (light) / next-teal-hi (dark). Centered or left-aligned as section heading underlines.

### Terminal

VT323 monospace. Dark background container with title bar (static diagonal barber pole stripes). Blinking cursor via CSS animation. Line-by-line typewriter reveal. Green/amber text for prompts, white for output. `prefers-reduced-motion` shows all lines immediately.

### Comparison Table

Full-width bordered table. 1px borders in rule (light) / next-mid (dark). Header row with accent-colored column labels. Alternating row backgrounds via surface/next-dark. Challenge section uses inverted background (ink/next-dark).

### Footer

`surface-strong` (light) / `next-dark` (dark) background. Top border. Logo with icon square. Navigation links. Terminal-style copyright line with accent-colored tech stack mention.

### Skip Link

`sr-only` by default. On focus: absolute positioned, accent-gold background, ink text, z-60, Silkscreen font. Keyboard accessible skip to main content.

### Theme Toggle

40x40px button. Border + bevel raised (dark). Shows moon icon (light) / sun icon (dark). Alpine.js x-show toggles visibility with x-cloak for no-flash. CRT flicker overlay on toggle (200ms white flash).

## 6. Do's and Don'ts

### Do:
- **Do** use zero border-radius on every element. Every corner is a right angle. This is non-negotiable.
- **Do** use Silkscreen for all headings, labels, and nav. system-ui for body. VT323 for terminals only.
- **Do** alternate section backgrounds every section. Paper/surface (light), next-black/next-dark (dark). Never three consecutive sections with the same background.
- **Do** use amber gold (light) / teal (dark) as the exclusive accent. One accent per mode, same structural positions.
- **Do** use 2px bevels in NeXT mode for interactive elements. They are the elevation system.
- **Do** respect `prefers-reduced-motion`. All animations must have a static fallback.
- **Do** use warm-tinted neutrals. Paper and ink have hue 85 (amber direction). Next-black and next-white have hue 270 (cool direction).
- **Do** use next-subtle (warm-tinted at hue 85) for dark mode secondary text. The two-color text system (cool white + warm beige) mirrors light mode (black + brown).
- **Do** vary section padding for rhythm. py-12 to py-32 depending on section importance.

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

For brand-level anti-references (what this product is NOT), see PRODUCT.md.
