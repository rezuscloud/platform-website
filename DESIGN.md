---
name: RezusCloud
description: "The Personal Cloud. Mac System 1 by day, NeXTSTEP by night."
colors:
  paper: "oklch(99.5% 0.004 85)"
  surface: "oklch(88.5% 0.006 85)"
  ink: "oklch(14% 0.008 65)"
  ink-muted: "oklch(30% 0.008 65)"
  rule: "oklch(72% 0.005 85)"
  accent-gold: "oklch(78% 0.16 75)"
  accent-gold-dark: "oklch(65% 0.14 75)"
  positive: "oklch(55% 0.12 150)"
  negative: "oklch(50% 0.12 25)"
  next-black: "oklch(6% 0.005 270)"
  next-dark: "oklch(20% 0.006 270)"
  next-mid: "oklch(34% 0.006 270)"
  next-light: "oklch(58% 0.006 270)"
  next-white: "oklch(88% 0.004 270)"
  next-subtle: "oklch(74% 0.006 270)"
  next-bevel-hi: "oklch(55% 0.006 270)"
  next-bevel-lo: "oklch(2% 0.004 270)"
  positive-next: "oklch(75% 0.14 150)"
  negative-next: "oklch(65% 0.14 25)"
typography:
  display:
    fontFamily: "'Silkscreen', system-ui, sans-serif"
    fontSize: "clamp(3rem, 7vw, 6rem)"
    fontWeight: 800
    lineHeight: 0.88
    letterSpacing: "-0.02em"
  headline:
    fontFamily: "'Silkscreen', system-ui, sans-serif"
    fontSize: "clamp(1.875rem, 4vw, 3rem)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "normal"
  title:
    fontFamily: "'Silkscreen', system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 700
    lineHeight: 1.3
    letterSpacing: "normal"
  body:
    fontFamily: "system-ui, -apple-system, 'Segoe UI', sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.625
    letterSpacing: "normal"
  label:
    fontFamily: "'Silkscreen', system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "0.05em"
  mono:
    fontFamily: "'VT323', 'Courier New', monospace"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.625
    letterSpacing: "normal"
rounded:
  none: "0px"
spacing:
  tight: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
  2xl: "48px"
  section-sm: "40px"
  section-md: "64px"
  section-lg: "80px"
  section-xl: "128px"
components:
  mac-button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.paper}"
    rounded: "{rounded.none}"
    padding: "12px 24px"
  mac-button-primary-hover:
    backgroundColor: "{colors.accent-gold}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "12px 24px"
  next-button-primary:
    backgroundColor: "{colors.next-light}"
    textColor: "{colors.next-black}"
    rounded: "{rounded.none}"
    padding: "12px 24px"
  next-button-primary-hover:
    backgroundColor: "{colors.next-mid}"
    textColor: "{colors.next-black}"
    rounded: "{rounded.none}"
    padding: "12px 24px"
  mac-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "16px"
  next-card:
    backgroundColor: "{colors.next-dark}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "16px"
  mac-nav:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "0 16px"
  next-nav:
    backgroundColor: "{colors.next-dark}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "0 16px"
  mac-table-header:
    backgroundColor: "{colors.rule}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "12px 16px"
  next-table-header:
    backgroundColor: "{colors.next-mid}"
    textColor: "{colors.next-black}"
    rounded: "{rounded.none}"
    padding: "12px 16px"
  mac-chip:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 12px"
  next-chip:
    backgroundColor: "{colors.next-dark}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "8px 12px"
---

# Design System: RezusCloud

## 1. Overview

**Creative North Star: "The Dual Desktop"**

This design system inhabits two eras of personal computing. Day mode is Mac System 1 (1984): paper-white surfaces, 1px flat borders, bitmap type, offset drop shadows, zero border-radius. Night mode is NeXTSTEP (1988): deep black surfaces, 3D beveled edges, scrolling CRT scanlines, phosphor afterglow, zero border-radius. The theme toggle is not dark/light. It is a shift between two computing revolutions from the same lineage.

Both modes share identical font families: Silkscreen for headings and labels, system-ui for body, VT323 for terminal and code. Only the colors and border strategies change between modes. Mac mode uses warm-tinted oklch neutrals (hue 85 degrees) with one amber gold accent. NeXT mode uses cool-tinted oklch neutrals (hue 270 degrees) in pure grayscale with no chromatic accent. The bevel is the emphasis in NeXT. The gold is the emphasis in Mac.

Shape language is constant: rectangles only, zero border-radius, 1px borders in Mac mode, 2px beveled edges in NeXT mode. The Mac desktop has a custom SVG arrow cursor, dot grid parallax, and animated title bar stripes. The NeXT desktop has scrolling CRT scanlines, phosphor text glow, and bevel hover sweeps. All animations respect prefers-reduced-motion.

**Key Characteristics:**
- Dual personality: Mac System 1 (day) and NeXTSTEP (night). Same fonts, different colors and borders.
- Zero border-radius everywhere. Neither the 1984 Mac nor the 1988 NeXT had curves.
- Flat 1px borders in Mac mode. 3D beveled edges (light top-left, dark bottom-right) in NeXT mode.
- One font set across both modes: Silkscreen bitmap (display/labels), system-ui (body), VT323 pixel (terminal).
- One accent color in Mac mode: amber gold, used sparingly. No accent in NeXT mode: pure grayscale.
- oklch color tokens throughout. All neutrals tinted: warm (hue 85 degrees) for Mac, cool (hue 270 degrees) for NeXT.
- Comparison tables as the canonical content format, not card grids.
- Mac offset drop shadows (1px 1px 0) for window-level elevation only. NeXT uses bevels exclusively.

## 2. Colors

Two chromatic worlds sharing a single oklch token system. Mac mode is warm monochrome with gold accent. NeXT mode is pure grayscale, no accent. All values are oklch with perceptual uniformity.

### Day Mode: The Macintosh Palette

Warm, papery, near-monochrome with a single gold accent. Extends the original Mac's strict black-and-white with warm amber tinting.

- **Paper** (oklch 99.5% 0.004 85): Primary background. Warm off-white, not pure white. Every surface starts here.
- **Surface** (oklch 88.5% 0.006 85): Secondary surface for cards, panels, nav bar. Provides depth through tonal layering.
- **Ink** (oklch 14% 0.008 65): Primary text and borders. Near-black with warm brown tint. Used for body text, border lines, and button fills.
- **Ink Muted** (oklch 30% 0.008 65): Decorative text: nav inactive state, footer labels, terminal output, badges. Not for functional/meaningful text (which uses ink at full contrast).
- **Rule** (oklch 72% 0.005 85): Dividers, table borders, subtle separators. Warm gray, visible but not assertive.
- **Accent Gold** (oklch 78% 0.16 75): The one saturated color. Used for hero emphasis, brand marks, icon dots, active states, divider bars. Never fills more than 10% of any viewport.
- **Accent Gold Dark** (oklch 65% 0.14 75): Hover and pressed states for gold elements.
- **Positive** (oklch 55% 0.12 150): Success indicators in terminal output.
- **Negative** (oklch 50% 0.12 25): Error indicators.

### Night Mode: The NeXTSTEP Palette

Pure grayscale, no chromatic color. Cool-tinted neutrals (hue 270 degrees) create a CRT phosphor feel. Depth comes from lightness variation, not hue. The bevel carries all emphasis.

- **NeXT Black** (oklch 6% 0.005 270): Primary background. The matte black of the NeXTcube.
- **NeXT Dark** (oklch 20% 0.006 270): Secondary surface for cards, panels, nav. Clearly distinct from black (14% gap) for section rhythm.
- **NeXT Mid** (oklch 34% 0.006 270): Elevated surfaces, table headers, button fills. The medium gray of NeXTSTEP window backgrounds.
- **NeXT Light** (oklch 58% 0.006 270): Bevel highlights, active nav pill background, CTA button fill. The lightest functional gray.
- **NeXT White** (oklch 88% 0.004 270): Primary text. Softened white like text on a CRT.
- **NeXT Subtle** (oklch 74% 0.006 270): Decorative secondary text. 7.2:1 on black (passes AA), 3.2:1 on dark surfaces (large text AA).
- **NeXT Bevel Hi** (oklch 55% 0.006 270): The light edge of 3D bevels. Applied to top and left borders.
- **NeXT Bevel Lo** (oklch 2% 0.004 270): The dark edge of 3D bevels. Applied to bottom and right borders.
- **Positive Next** (oklch 75% 0.14 150): Success indicators (bright for dark backgrounds).
- **Negative Next** (oklch 65% 0.14 25): Error indicators (bright for dark backgrounds).

### Named Rules

**The Dual Era Rule.** Day mode and night mode are not light/dark variants. They are two distinct computing eras with different border strategies and color logic, sharing only font families. Mac mode uses warm oklch on paper. NeXT mode uses cool oklch on black. Do not blend them.

**The One Accent Rule.** Mac mode gets exactly one saturated color: amber gold, under 10% of any viewport. NeXT mode gets zero chromatic colors. The bevel is the emphasis. Chromatic color in NeXT mode breaks the grayscale purity.

**The Tinted Neutral Rule.** No pure black or pure white. All neutrals are oklch with chroma 0.003 to 0.008, tinted warm (hue 85) for Mac or cool (hue 270) for NeXT. Pure gray is forbidden.

## 3. Typography

One font set for both modes. The distinction is chromatic, not typographic.

**Display Font:** Silkscreen (bitmap pixel font, 12KB regular + 8KB bold woff2)
**Body Font:** system-ui (platform native, no download)
**Mono Font:** VT323 (pixel terminal font, 12KB woff2, ASCII subset with unicode-range)

**Character:** Silkscreen is a proper bitmap font that evokes the Mac's original Chicago and Geneva without being a recreation. It has the same pixel-grid clarity and friendly geometry Susan Kare designed. VT323 provides period-authentic terminal output. system-ui keeps body text readable at speed. The pairing is defiant (bitmap display), precise (system body), nostalgic (pixel terminal).

### Hierarchy

- **Display** (extrabold, clamp(3rem, 7vw, 6rem), 0.88 line-height, tight tracking): Hero headline only. Silkscreen. The biggest word on the page.
- **Headline** (bold, clamp(1.875rem, 4vw, 3rem), 1.1 line-height): Section titles. Silkscreen. Bold and unmissable.
- **Title** (bold, 1.125rem, 1.3 line-height): Subsection headings, card titles, layer labels. Silkscreen.
- **Body** (regular, 1rem, 1.625 line-height): Paragraphs, descriptions. system-ui. Max line length 65ch.
- **Label** (bold, 0.75rem, 1.2 line-height, uppercase, 0.05em tracking): Navigation, table headers, badges. Silkscreen.
- **Mono** (regular, 0.75rem to 0.875rem, 1.625 line-height): Terminal output, code, system readout. VT323.

### Named Rules

**The Unified Font Rule.** Both modes use the same font families. Only colors change between Mac and NeXT. Display in Silkscreen, body in system-ui, terminal in VT323. The font switch is not a mode signal. The color switch is.

## 4. Elevation

The two modes use fundamentally different elevation strategies. Mac mode is flat with occasional offset shadows. NeXT mode uses 3D beveled edges. Both reject diffuse drop shadows.

### Mac Mode: Flat with Offset Shadow

The original Mac used a 1px black drop shadow offset down-right for window layering. This system recreates that treatment sparingly:

- **Window Shadow** (box-shadow: 1px 1px 0 var(--color-ink)): Applied to window-level containers only (terminal widgets). Never on cards, buttons, chips, or badges.
- **Flat at rest** (box-shadow: none): Cards, tables, navigation, badges. The 1px border is the shape.
- **Tonal layering**: Sections alternate paper (99.5%) and surface (88.5%) for rhythm.

### NeXT Mode: 3D Beveled Edges

NeXTSTEP's entire visual identity was its 3D beveled edges. Every surface has a light edge (top-left) and dark edge (bottom-right):

- **Raised Bevel** (2px solid, border-color: bevel-hi bevel-lo bevel-lo bevel-hi): Buttons, elevated panels, active elements. The surface appears to protrude.
- **Sunken Bevel** (2px solid, border-color: bevel-lo bevel-hi bevel-hi bevel-lo): Input fields, terminal body, inset areas. The surface appears recessed.
- **No box-shadow**: NeXTSTEP did not use drop shadows. The bevel is the elevation.
- **Tonal layering**: Sections alternate next-black (6%) and next-dark (20%) for rhythm.

### CRT Effects (NeXT only)

- **Scanlines**: 1px bright lines every 2px at 8% opacity, scrolling vertically at 8s cycle.
- **Phosphor afterglow**: Terminal lines glow on appear (6px + 12px text-shadow), fading over 2.5s.
- **CRT flicker**: Theme toggle triggers a double-flash white overlay (40ms on, 40ms off, 120ms switch).

### Named Rules

**The Flat-By-Default Rule (Mac).** Mac surfaces are flat at rest. Shadows appear only on window-level containers.

**The Bevel-Is-Elevation Rule (NeXT).** Every NeXT surface has a bevel. Raised for protruding elements, sunken for recessed. If a NeXT element has no bevel, it is wrong.

## 5. Components

Every component is rectangular. Zero border-radius is non-negotiable. The difference between modes is border treatment and color.

### Buttons

- **Mac Primary**: Ink background, paper text, 1px ink border. Padding 12px 24px. Font: Silkscreen bold. Hover fills accent-gold.
- **NeXT Primary**: NeXT Light background, NeXT Black text, raised bevel. Padding 12px 24px. Font: Silkscreen bold. Hover fills NeXT Mid.
- **Ghost (both)**: Transparent background, text color, 1px border. Hover fills surface color.

### Cards / Containers

- **Corner Style**: Sharp (0px radius).
- **Mac**: Surface background, 1px rule border, flat. Hover deepens border to accent-gold.
- **NeXT**: NeXT Dark background, raised bevel. Hover brightens bevel highlight.
- **Internal Padding**: 16px standard, 24px for content-dense layouts.
- **Nesting**: Prohibited.

### Tables

The signature component. Comparison tables carry the argument.

- **Mac Header**: Rule background, ink text, uppercase Silkscreen labels.
- **Mac Row**: Paper background, ink text. Hover shifts to surface.
- **NeXT Header**: NeXT Mid background, NeXT Black text, raised bevel.
- **NeXT Row**: NeXT Black background, NeXT White text.
- **Accent columns**: Positive/negative semantic colors per mode.

### Navigation

- **Mac**: Surface background, 1px rule border bottom (flat). Silkscreen bold labels. Active section highlighted with paper background.
- **NeXT**: NeXT Dark background, raised bevel bottom edge. Active section highlighted with NeXT Light background and NeXT Black text (5.7:1 contrast).
- **Mobile**: Hamburger icon (rectangular). Menu drops as full-width panel. No overlay blur.
- **Touch targets**: Desktop py-1.5 (~36px), mobile py-2 (~44px).

### Chips / Badges

- **Mac**: Surface background, 1px rule border, ink text. Amber dot indicator.
- **NeXT**: NeXT Dark background, raised bevel, NeXT White text. No dot indicator.

### Terminal Widget (signature component)

- **Mac mode**: Paper background, 1px ink border, window shadow (1px 1px 0). Animated title bar with repeating horizontal stripes. Three colored dots (red, yellow, green) hidden in dark mode. VT323 body text. Typewriter entrance animation.
- **NeXT mode**: NeXT Black background, sunken bevel (recessed screen). Medium gray title strip. VT323 body text with phosphor afterglow on each line. Typewriter entrance animation.

### Theme Toggle

- **Mac state**: Moon icon on ghost button (paper fill, ink border). 40x40px.
- **NeXT state**: Sun icon on beveled button (NeXT Dark fill, raised bevel). 40x40px.
- **Transition**: CRT flicker overlay (double-flash white), then 150ms ease-out color switch. Font families stay constant.

## 6. Do's and Don'ts

### Do:

- **Do** use 1px flat borders in Mac mode. Every container, card, and interactive element has a 1px border in rule color.
- **Do** use 2px beveled edges in NeXT mode. Raised for protruding, sunken for recessed. The bevel IS the visual identity.
- **Do** use zero border-radius on every element without exception.
- **Do** alternate section backgrounds (paper/surface for Mac, next-black/next-dark for NeXT) for visual rhythm.
- **Do** use Silkscreen for all display, heading, label, and nav text. system-ui for body. VT323 for terminal.
- **Do** keep fonts identical between modes. Only colors change.
- **Do** cap body line length at 65ch.
- **Do** use comparison tables as the primary content format for arguments.
- **Do** keep accent gold under 10% of any Mac mode viewport.
- **Do** keep NeXT mode pure grayscale. No chromatic color.
- **Do** use oklch for all color tokens. Tint all neutrals.
- **Do** respect prefers-reduced-motion. Disable all animations, scanlines, and parallax.

### Don't:

- **Don't** use border-radius anywhere. Rounded corners break both eras instantly.
- **Don't** use gradient text (background-clip: text with gradient). Neither platform could render gradients on text.
- **Don't** use glassmorphism, backdrop blur, or frosted-glass effects. Neither platform had transparency.
- **Don't** create identical card grids with icon + heading + text repeated endlessly. Vary layout per section.
- **Don't** use the hero-metric template (big number, small label, stats). SaaS cliche.
- **Don't** use side-stripe borders (border-left or border-right greater than 1px as colored accent).
- **Don't** apply box-shadow to cards, buttons, chips, or badges. Mac shadows are window-level only. NeXT never uses shadows.
- **Don't** use em dashes. Use commas, colons, semicolons, or periods.
- **Don't** use nested cards. Flatten or restructure.
- **Don't** treat night mode as "dark mode." It is NeXTSTEP. Different border strategy, different color logic.
- **Don't** introduce chromatic color in NeXT mode. No blue links, no green accents, no gold carry-over.
- **Don't** use pure black (#000) or pure white (#fff). All neutrals must be tinted via oklch.
- **Don't** use different font families between modes. The font set is shared. Colors differ, fonts do not.
