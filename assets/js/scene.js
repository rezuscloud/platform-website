(() => {
  const root = document.querySelector("[data-scene-root]")
  if (!root) return

  const track = root.querySelector("[data-scene-track]")
  if (!track) return

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)")
  const navHeight = 56

  const presets = {
    desktop: { startScale: 1.75, endScale: 1 },
    mobile: { startScale: 1.45, endScale: 1 },
  }

  let metrics = { start: 0, end: 1 }
  let ticking = false

  function clamp(v) { return Math.max(0, Math.min(1, v)) }
  function ease(v) { return v * v * (3 - 2 * v) }
  function lerp(a, b, t) { return a + (b - a) * t }

  function activePreset() {
    return window.innerWidth < 768 ? presets.mobile : presets.desktop
  }

  function apply(progress) {
    const p = activePreset()
    const eased = ease(progress)
    const scale = lerp(p.startScale, p.endScale, eased)
    root.style.setProperty("--scene-progress", eased.toFixed(3))
    root.style.setProperty("--scene-scale", scale.toFixed(3))
  }

  function measure() {
    const rect = track.getBoundingClientRect()
    const top = window.scrollY + rect.top
    const stickyH = Math.max(1, window.innerHeight - navHeight)
    metrics = {
      start: top,
      end: top + Math.max(1, track.offsetHeight - stickyH),
    }
  }

  function update() {
    ticking = false
    if (prefersReducedMotion.matches) {
      root.dataset.sceneMotion = "reduced"
      root.style.setProperty("--scene-progress", "1")
      root.style.setProperty("--scene-scale", "1")
      return
    }
    delete root.dataset.sceneMotion
    const span = Math.max(1, metrics.end - metrics.start)
    apply(clamp((window.scrollY - metrics.start) / span))
  }

  function schedule() {
    if (ticking) return
    ticking = true
    requestAnimationFrame(update)
  }

  function refresh() { measure(); schedule() }

  window.addEventListener("scroll", schedule, { passive: true })
  window.addEventListener("resize", refresh, { passive: true })
  window.addEventListener("load", refresh)

  if (typeof ResizeObserver !== "undefined") {
    new ResizeObserver(refresh).observe(track)
  }

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", refresh)
  }

  refresh()
})()
