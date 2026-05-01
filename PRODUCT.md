# Product

## Register

brand

## Users

Individual developers and small-to-medium businesses who limit their software design capabilities to avoid cloud complexity or heavy provider bills. They recognize that 99% of cloud workloads use 1% of the available services, and they want a simpler path. They are comfortable with open source, skeptical of vendor lock-in, and attracted to the idea of owning their infrastructure the way they own their computer. They are not Kubernetes experts, and they should not need to become one.

## Product Purpose

RezusCloud is the Personal Cloud: a fully open source, free platform that provides an architectural golden path covering 90% of real-world use cases with full scalability out of the box. It lets anyone deploy like a full-scale enterprise, on hardware they already own, in an afternoon. The website exists to communicate this vision, build conviction, and capture early interest before public launch.

## Brand Personality

**Defiant, precise, nostalgic.**

Defiant: challenges the status quo that cloud must be rented, complex, and expensive. Positions cloud providers as the new mainframes. The tone is confident and direct, never pleading or salesy.

Precise: every claim is a fact, not hype. Numbers are specific (200MB, 30s boot, 2-4h setup). The platform works because the engineering is sound, not because the marketing says so.

Nostalgic: visual language drawn from the Mac System 1 era (1984), the moment a device challenged the IT industry by making computing personal. The retro aesthetic is not decoration; it is the argument made visual. Shapes, borders, fonts, and color treatment all reference that era. The Mac was a revolution presented as friendly and approachable. So is this.

## Anti-references

- Corporate Kubernetes vendor pages (Rancher, VMware Tanzu, Red Hat OpenShift): too enterprise, too many badges and certifications, no personality.
- Generic SaaS landing pages: gradient heroes, trust logos, monthly pricing tables, "Get Started Free" CTAs. This is not a subscription product.
- Cloud provider console aesthetics (AWS, GCP, Azure): infinite service catalogs, complex navigation, feature-matrix overload. This product is the opposite of that.
- Dribbble/Behance "retro" landing pages that use retro as surface decoration without the shape language to back it up. If it has rounded corners and a gradient, it is not retro.

## Design Principles

1. **Personal, not enterprise.** Every design choice asks: would the 1984 Mac have done it this way? If it feels like it belongs in a boardroom, it is wrong. If it feels like it belongs on a desk, it is right.

2. **Shapes carry the era.** 1px solid borders, no border-radius, drop shadows offset 1px bottom-right, bitmap texture patterns. The retro feel comes from geometry, not from slapping a pixel font on a modern layout.

3. **Facts over features.** No feature lists padded with buzzwords. Every section makes one clear point with evidence. The comparison tables are the canonical format: here is what renting looks like, here is what owning looks like, you decide.

4. **Revolution, not subscription.** The page is a manifesto, not a sales funnel. No pricing tiers, no monthly plans, no "contact sales." The CTA is conviction, not checkout.

5. **The 1% principle.** 99% of workloads use 1% of services. The design reflects this restraint. No infinite grids, no overwhelming information density. Show exactly what matters, stop.

## Accessibility

WCAG AA compliance following web framework best practices. No specific user needs beyond that. The bitmap aesthetic naturally limits some contrast scenarios; meet AA minimums without compromising the visual identity. Ensure keyboard navigation works for all interactive elements. Provide sufficient motion reduction support for the CRT and animation effects.

## Visual Direction

Mac System 1 (1984) as the primary visual reference. The original Macintosh was black and white; color is extended sparingly from that foundation using the warm amber already established in the project palette. Shape language is strict: rectangular, 1px borders, offset drop shadows, bitmap textures. The current cream/amber/phosphor palette maps well to an extended System 1 aesthetic where cream is the paper-white background, amber is the accent, and phosphor green marks terminal/code content. Typography should use bitmap-style fonts (VT323 is correct; IBM Plex Mono is too modern for this direction and should be reconsidered or supplemented with a more era-appropriate choice like Chicago or a Geneva-style bitmap face).
