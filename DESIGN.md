---
name: RezusCloud
description: "The Personal Cloud. Mac System 1 by day, NeXTSTEP by night."
colors:
  paper: "#faf8f3"
  surface-warm: "#f5f0e8"
  ink: "#2a1a0e"
  ink-muted: "#5e503b"
  rule: "#d4c9b0"
  accent-gold: "#ffb000"
  accent-gold-dark: "#cc8d00"
  next-black: "#1a1a1a"
  next-dark: "#2d2d2d"
  next-mid: "#4a4a4a"
  next-light: "#6e6e6e"
  next-white: "#d4d4d4"
  next-subtle: "#999999"
  next-bevel-hi: "#666666"
  next-bevel-lo: "#111111"
typography:
  mac-display:
    fontFamily: "'Chicago', 'Charcoal', 'Geneva', system-ui, sans-serif"
    fontSize: "clamp(2rem, 5vw, 3.5rem)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "normal"
  mac-body:
    fontFamily: "'Geneva', 'Chicago', system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  mac-label:
    fontFamily: "'Geneva', 'Chicago', system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.05em"
  mac-mono:
    fontFamily: "'Monaco', 'VT323', 'Courier New', monospace"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
  next-display:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif"
    fontSize: "clamp(2rem, 5vw, 3.5rem)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "-0.02em"
  next-body:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  next-label:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, Arial, sans-serif"
    fontSize: "11px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.04em"
  next-mono:
    fontFamily: "'Courier New', 'Menlo', 'Consolas', monospace"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: "normal"
rounded:
  none: "0px"
spacing:
  tight: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "48px"
  section: "80px"
components:
  mac-button-primary:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.paper}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  mac-button-primary-hover:
    backgroundColor: "{colors.accent-gold}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  mac-button-ghost:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  mac-button-ghost-hover:
    backgroundColor: "{colors.surface-warm}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  next-button-primary:
    backgroundColor: "{colors.next-mid}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  next-button-primary-hover:
    backgroundColor: "{colors.next-light}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "8px 24px"
  mac-card:
    backgroundColor: "{colors.surface-warm}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "16px"
  next-card:
    backgroundColor: "{colors.next-dark}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "16px"
  mac-nav:
    backgroundColor: "{colors.surface-warm}"
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
    textColor: "{colors.ink-muted}"
    rounded: "{rounded.none}"
    padding: "6px 16px"
  mac-table-row:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 16px"
  next-table-header:
    backgroundColor: "{colors.next-mid}"
    textColor: "{colors.next-subtle}"
    rounded: "{rounded.none}"
    padding: "6px 16px"
  next-table-row:
    backgroundColor: "{colors.next-black}"
    textColor: "{colors.next-white}"
    rounded: "{rounded.none}"
    padding: "8px 16px"
---

# Design System: RezusCloud

## 1. Overview

**Creative North Star: "The First Macintosh"**

This design system inhabits two eras from the same lineage. Day mode speaks the visual language of the 1984 Macintosh: paper-white surfaces, 1px black borders, Chicago bitmap type, offset drop shadows, and zero border-radius. Night mode transforms into NeXTSTEP (1988): dark gray surfaces, 3D beveled edges, smooth anti-aliased Helvetica, pure grayscale palette, and the same strict rectangular geometry. The theme toggle is not dark/light. It is a shift between two computing revolutions from the same visionary.

The two modes share the same DNA (Steve Jobs created both) but express it differently. Mac mode is warm, approachable, bitmap-textured: the computer that changed everything by being friendly. NeXT mode is cool, precise, beveled: the computer that changed everything by being sophisticated. Both reject modern web conventions. No rounded corners, no gradient text, no floating card shadows, no hero metrics.

Shape language is constant across modes because the constraint is constant: the pixel grid. Rectangles only. 1px borders in Mac mode, 3D bevels in NeXT mode. No curves, no softness. The difference is chromatic and textural: Mac extends its original black-and-white with warm amber on cream; NeXT commits to pure grayscale on dark gray. The accent color shifts from gold (warm, personal) to nothing (the bevel is the emphasis).

**Key Characteristics:**
- Dual personality: Mac System 1 (day) and NeXTSTEP (night). Not "dark mode" but a shift between two distinct computing eras from the same lineage.
- Zero border-radius everywhere. Neither the 1984 Mac nor the 1988 NeXT had curves.
- Flat 1px borders in Mac mode. 3D beveled edges (light top-left, dark bottom-right) in NeXT mode. The border strategy IS the mode signal.
- Bitmap fonts (Chicago, Geneva) in Mac mode. Anti-aliased Helvetica in NeXT mode. The font switch is the other strong mode signal.
- One accent color in Mac mode: amber gold, under 10% of viewport. No accent in NeXT mode: pure grayscale, the bevel carries emphasis.
- Comparison tables as the canonical content format, not card grids. Facts in rows, not icons in boxes.
- Offset drop shadows (1px 1px 0) for Mac window-level elevation only. NeXT mode uses beveled surfaces instead of shadows.

## 2. Colors

Two chromatic worlds. Mac mode is warm monochrome with gold accent. NeXT mode is pure grayscale, no accent.

### Day Mode: The Macintosh Palette

Warm, papery, near-monochrome with a single gold accent. Extends the original Mac's strict black-and-white with the warm amber already established as the brand color.

- **Paper** (#faf8f3): Primary background. Warm off-white, not pure white. The color of aged printer paper under desk lamp light. Every surface starts here.
- **Warm Surface** (#f5f0e8): Secondary surface for cards, panels, nav bar. One shade darker than paper, providing depth without shadow.
- **Ink** (#2a1a0e): Primary text and borders. Near-black with a warm brown tint. The color of typewriter ribbon on cream stock. Used for all border lines, body text, and button fills.
- **Ink Muted** (#5e503b): Secondary text, labels, table headers. Readable but recessive. The color of pencil notes in margins.
- **Rule** (#d4c9b0): Dividers, table borders, subtle separators. Warm gray, visible but not assertive. System 1 used thin black lines for structure; this is the softened equivalent.
- **Accent Gold** (#ffb000): The one color. Used for active states, highlighted values, brand emphasis, hover effects. Never fills more than 10% of any viewport. Its rarity is its impact.
- **Accent Gold Dark** (#cc8d00): Hover and pressed states for gold elements. Slightly deeper, never brighter.

### Night Mode: The NeXTSTEP Palette

Pure grayscale, no chromatic color. The NeXTcube was matte black, the UI was grayscale, and the beveled 3D edges were the visual identity. Depth comes from lightness variation, not hue.

- **NeXT Black** (#1a1a1a): Primary background. The matte black of the NeXTcube itself. Not pure black, slightly warm.
- **NeXT Dark** (#2d2d2d): Secondary surface for cards, panels, nav. One step up from black, like a gray bezel.
- **NeXT Mid** (#4a4a4a): Elevated surfaces, table headers, button fills. The medium gray of NeXTSTEP window backgrounds.
- **NeXT Light** (#6e6e6e): Bevel highlights, subtle borders. The lighter edge that catches light.
- **NeXT White** (#d4d4d4): Primary text. Not pure white. Slightly softened, like text on a CRT with a slight warm cast.
- **NeXT Subtle** (#999999): Secondary text, labels, disabled states. Medium gray, readable but recessive.
- **NeXT Bevel Hi** (#666666): The light edge of 3D bevels. Applied to top and left borders.
- **NeXT Bevel Lo** (#111111): The dark edge of 3D bevels. Applied to bottom and right borders.

### Named Rules

**The Dual Era Rule.** Day mode and night mode are not light/dark variants of one palette. They are two distinct computing eras with different fonts, different border strategies, and different color logic. Mac mode uses warm browns and gold on paper. NeXT mode uses pure grayscale on dark surfaces. Do not blend them or create a "neutral" middle.

**The One Accent Rule.** Mac mode gets exactly one saturated color: amber gold, on no more than 10% of any viewport. NeXT mode gets zero accent colors. The bevel is the emphasis. If NeXT mode introduces chromatic color, it breaks the grayscale purity that makes it read as NeXTSTEP.

## 3. Typography

Two complete type systems, one per mode. The Mac used bitmap proportional faces. NeXTSTEP used anti-aliased Helvetica via Display PostScript, one of the first UIs with smooth text. This distinction is load-bearing: the font rendering is the strongest signal of which era is active.

**Mac Mode Fonts:** Chicago-style bitmap sans for display, Geneva-style bitmap sans for body. Both require bitmap-rendered web fonts (.woff2), not anti-aliased recreations. Source a Chicago web font (various recreations exist as "Chicago" or "Charcoal") and a Geneva web font. Monaco for monospace content (code, data).

**NeXT Mode Fonts:** System font stack (-apple-system, Helvetica Neue, Helvetica, sans-serif). NeXTSTEP was designed around Helvetica and was one of the first interfaces to use anti-aliased text. The smooth rendering is authentic to the era, not a modern concession. Courier New for monospace content (code, terminal).

**Character:** The Mac type is friendly and approachable, with the rounded bitmap shapes Susan Kare designed to make the computer feel human. The NeXT type is precise and sophisticated, with Helvetica's clean geometry conveying technical authority. Together they embody the brand personality: defiant (NeXT's raw ambition), precise (Helvetica's geometry), nostalgic (both).

### Hierarchy

**Mac Mode:**
- **Display** (bold, clamp(2rem, 5vw, 3.5rem), 1.1 line-height): Hero headlines only. Chicago-style. Sets the tone for the entire page.
- **Headline** (bold, 24px, 1.2 line-height): Section titles. Chicago-style. Bold and clear.
- **Title** (semibold, 18px, 1.3 line-height): Subsection headings, card titles. Geneva-style.
- **Body** (regular, 14px, 1.5 line-height): Paragraphs, descriptions. Geneva-style. Max line length 65ch.
- **Label** (semibold, 12px, 1.2 line-height, uppercase): Navigation items, table headers, tags. Geneva-style with tracking.
- **Mono** (regular, 13px, 1.4 line-height): Code, terminal output, data values. Monaco/VT323.

**NeXT Mode:**
- **Display** (bold, clamp(2rem, 5vw, 3.5rem), 1.1 line-height, -0.02em tracking): Hero headlines. Helvetica at large size with tight tracking. Weight contrast is the hierarchy tool.
- **Headline** (bold, 22px, 1.2 line-height): Section titles.
- **Title** (semibold, 16px, 1.3 line-height): Subsection headings.
- **Body** (regular, 14px, 1.5 line-height): Paragraphs, descriptions. Same size as Mac body but smooth rendering. Max line length 65ch.
- **Label** (semibold, 11px, 1.2 line-height, uppercase): Navigation, headers. Tighter and smaller than Mac labels, matching NeXTSTEP's compact UI density.
- **Mono** (regular, 13px, 1.4 line-height): Code, terminal output. Courier New.

### Named Rules

**The Era Fidelity Rule.** Mac mode never uses the system font stack for body text. NeXT mode never uses Chicago/Geneva. The font is the strongest signal of which era is active. Terminal widgets and code blocks may use monospace in both modes, but UI chrome and body copy stay in their respective families.

## 4. Elevation

The two modes use fundamentally different elevation strategies. Mac mode is flat with occasional offset shadows. NeXT mode uses 3D beveled edges as its primary visual language. Both reject the modern web pattern of diffuse drop shadows on every card.

### Mac Mode: Flat with Offset Shadow

The original Mac used a 1px black drop shadow offset down-right to indicate window layering. This system recreates that treatment sparingly:

- **Window Shadow** (`box-shadow: 1px 1px 0 #2a1a0e`): Applied to window-level containers only (the terminal widget, hypothetical dialog boxes). Never on cards, buttons, chips, or badges.
- **Flat at rest** (`box-shadow: none`): Cards, tables, navigation, badges. All flat. The 1px border is the shape.
- **Tonal layering**: Sections alternate paper → surface-warm → paper for rhythm.

### NeXT Mode: 3D Beveled Edges

NeXTSTEP's entire visual identity was its 3D beveled edges. Every surface had a light edge (top-left) and dark edge (bottom-right), creating the illusion of depth without shadows:

- **Raised Bevel**: `border-style: solid; border-width: 1px; border-color: #666666 #111111 #111111 #666666;` Used for buttons, elevated panels, active elements. The surface appears to protrude.
- **Sunken Bevel**: `border-style: solid; border-width: 1px; border-color: #111111 #666666 #666666 #111111;` Used for input fields, inset areas, the terminal body. The surface appears recessed.
- **No box-shadow**: NeXTSTEP did not use drop shadows. The bevel is the elevation.
- **Tonal layering**: Sections alternate next-black → next-dark → next-black for rhythm.

### Named Rules

**The Flat-By-Default Rule (Mac).** Mac surfaces are flat at rest. Shadows appear only on window-level containers, never on content cards or interactive elements.

**The Bevel-Is-Elevation Rule (NeXT).** Every NeXT surface has a bevel. Raised for protruding elements (buttons, panels), sunken for recessed elements (inputs, terminal body). If a NeXT element has no bevel, it is wrong.

## 5. Components

Every component is rectangular across both modes. No border-radius exists in this system. The difference between modes is in border treatment and color, not shape.

### Buttons

Rectangular, no curves. Mac buttons are flat with 1px borders. NeXT buttons are 3D beveled.

- **Mac Primary:** Ink background, paper text, no border (the fill is the shape). Padding 8px 24px. Font: Geneva bold 12px, uppercase with tracking.
- **Mac Primary Hover:** Accent gold background, ink text.
- **Mac Ghost:** Paper background, ink text, 1px ink border. Hover fills surface-warm.
- **NeXT Primary:** NeXT Mid background (#4a4a4a), NeXT White text, raised bevel (light top-left, dark bottom-right). Padding 8px 24px. Font: Helvetica bold 11px, uppercase with tracking.
- **NeXT Primary Hover:** NeXT Light background (#6e6e6e), same bevel.
- **NeXT Ghost:** NeXT Dark background (#2d2d2d), NeXT White text, 1px NeXT Light border. Hover fills NeXT Mid.
- **Disabled (both):** Same shape, reduced opacity (0.4). Cursor not-allowed.

### Cards / Containers

- **Corner Style:** Sharp (0px radius). Uniform across both modes.
- **Mac Background:** surface-warm (#f5f0e8) with 1px rule (#d4c9b0) flat border. Hover deepens border to accent-gold.
- **NeXT Background:** next-dark (#2d2d2d) with raised bevel (light top-left, dark bottom-right). Hover brightens the bevel highlight.
- **Shadow Strategy:** None in either mode. Mac uses flat borders; NeXT uses bevels.
- **Internal Padding:** 16px standard. Content-dense cards may use 24px.
- **Nesting:** Prohibited. If a card contains another bordered container, remove the inner border or restructure.

### Tables

The signature component. PRODUCT.md's "facts over features" principle means comparison tables carry the argument.

- **Mac Header:** Rule background (#d4c9b0), ink-muted text, uppercase 12px with tracking. 1px rule border bottom.
- **Mac Row:** Paper background, ink text. 1px rule border bottom. Hover shifts to surface-warm.
- **NeXT Header:** NeXT Mid background (#4a4a4a), NeXT Subtle text (#999999), uppercase 11px with tracking. Raised bevel on the header row.
- **NeXT Row:** NeXT Black background (#1a1a1a), NeXT White text (#d4d4d4). 1px NeXT Dark border bottom. Hover shifts to NeXT Dark.
- **Accent columns:** Positive values in green (#339933 for Mac, #55aa55 for NeXT), negative in red (#cc3333 for Mac, #aa4444 for NeXT). Used for data contrast only.

### Navigation

- **Mac:** Surface-warm background, 1px rule border bottom (flat). Geneva bold 12px links, uppercase. Active section in ink weight. Fixed top.
- **NeXT:** NeXT Dark background, raised bevel bottom edge. Helvetica bold 11px links, uppercase. Active section in NeXT White, inactive in NeXT Subtle. Fixed top.
- **Mobile:** Hamburger icon (rectangular, no curves). Menu drops as a full-width panel with vertical link stack. No overlay, no backdrop blur.
- **Logo:** The existing monitor icon stays. Mac: ink fill with accent-gold screen. NeXT: NeXT Mid fill with NeXT White screen. The icon should feel like a 32x32 bitmap.

### Chips / Badges

- **Shape:** Sharp rectangle, 0px radius. Inline-flex with tight padding (4px 12px).
- **Mac:** Surface-warm background, 1px rule border, ink-muted text. Dot indicator in accent-gold.
- **NeXT:** NeXT Dark background, raised bevel, NeXT Subtle text. No dot indicator (the bevel is the accent).

### Terminal Widget (signature component)

In both modes, the terminal should feel like an actual screen within the page.

- **Mac mode:** Paper background with 1px ink border. Window shadow (1px 1px 0 ink). Title bar with repeating horizontal lines pattern in rule color, three circular dots (red, yellow, green) at left. Body in Monaco 13px, ink-muted text. Prompt in accent-gold. Blinking cursor in ink.
- **NeXT mode:** NeXT Black background with sunken bevel (recessed into the page, like a screen inset into the bezel). No title bar with dots. Instead, a NeXT-style title bar: medium gray strip with the terminal title in NeXT Subtle. Body in Courier New 13px, NeXT White text. Prompt in NeXT Light. Blinking cursor in NeXT White.

### Theme Toggle

The bridge between two eras. Should feel like switching computers.

- **Mac mode visible:** Moon icon in ink on a ghost button (1px ink border, paper fill).
- **NeXT mode visible:** Sun icon in NeXT White on a beveled button (raised, NeXT Dark fill).
- **Button shape:** Sharp rectangle, 0px radius, 32x32px.
- **Transition:** Quick (150ms ease-out). Switches background, text color, border strategy (flat ↔ bevel), and font family. The font change is the strongest signal.

## 6. Do's and Don'ts

### Do:

- **Do** use 1px flat borders in Mac mode. Every container, card, and interactive element has a 1px border in rule color. The border is the shape.
- **Do** use 3D beveled edges in NeXT mode. Raised bevels on protruding elements, sunken bevels on recessed elements. The bevel IS the visual identity.
- **Do** use zero border-radius on every element without exception. If an element has border-radius, it is not part of this system.
- **Do** alternate section backgrounds (paper/surface-warm for Mac, next-black/next-dark for NeXT) for visual rhythm.
- **Do** use bitmap fonts in Mac mode (Chicago, Geneva, Monaco). Anti-aliased smooth fonts break the 1984 Mac illusion.
- **Do** use anti-aliased Helvetica in NeXT mode. Bitmap fonts in NeXT mode break the 1988 NeXTSTEP illusion.
- **Do** cap body line length at 65ch in both modes.
- **Do** use comparison tables as the primary content format for arguments.
- **Do** keep the accent gold under 10% of any Mac mode viewport. Gold is a spice, not the meal.
- **Do** keep NeXT mode pure grayscale. No chromatic color. The bevel is the emphasis.
- **Do** make the theme toggle feel like switching between a Macintosh and a NeXTcube, not adjusting brightness.

### Don't:

- **Don't** use border-radius anywhere. Rounded corners instantly break both the System 1 and NeXTSTEP illusions.
- **Don't** use gradient text (`background-clip: text` with gradient). Neither the Mac nor NeXT could render gradients on text. A single solid color always.
- **Don't** use glassmorphism, backdrop blur, or frosted-glass effects. Neither platform had transparency.
- **Don't** create identical card grids with icon + heading + text repeated endlessly. PRODUCT.md's anti-references reject this explicitly. Vary sizes, merge into tables, or eliminate.
- **Don't** use the hero-metric template (big number, small label, stats). SaaS cliché, explicitly rejected in PRODUCT.md.
- **Don't** use side-stripe borders (`border-left` or `border-right` greater than 1px as a colored accent). Use full borders, background tints, or leading icons.
- **Don't** use 2px or 3px borders. Mac used 1px; NeXT used 1px bevels. Thicker borders look like modern "retro" decoration.
- **Don't** apply box-shadow to cards, buttons, chips, or badges. Mac mode shadows are for window-level containers only. NeXT mode never uses shadows.
- **Don't** use IBM Plex Mono or any modern tech sans-serif as a primary font. Chicago/Geneva for Mac, Helvetica for NeXT. Period.
- **Don't** blend the two modes. Mac is warm amber on cream with flat borders. NeXT is pure grayscale on dark gray with bevels. No hybrid elements.
- **Don't** use em dashes. Use commas, colons, semicolons, or periods.
- **Don't** use nested cards. Flatten the structure or remove the inner container.
- **Don't** use rounded icons with gradient backgrounds above every heading. Screams template.
- **Don't** treat night mode as "dark mode." It is NeXTSTEP. Different fonts, different border strategy, different color logic. Night mode is not Mac mode with inverted colors.
- **Don't** introduce chromatic color in NeXT mode. No blue links, no green accents, no gold carry-over. Pure grayscale. The bevel carries all emphasis.
