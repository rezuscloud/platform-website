(() => {
  const root = document.querySelector("[data-scene-root]")
  if (!root) {
    return
  }

  const track = root.querySelector("[data-scene-track]")
  if (!track) {
    return
  }

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)")
  const navHeight = 56

  const presets = {
    desktop: {
      startScale: 1.72,
      endScale: 1,
    },
    mobile: {
      startScale: 1.42,
      endScale: 1,
    },
  }

  let metrics = { start: 0, end: 1 }
  let ticking = false

  function clamp(value) {
    return Math.max(0, Math.min(1, value))
  }

  function ease(value) {
    return value * value * (3 - 2 * value)
  }

  function lerp(start, end, amount) {
    return start + (end - start) * amount
  }

  function activePresets() {
    return window.innerWidth < 768 ? presets.mobile : presets.desktop
  }

  function setState(progress) {
    const currentPresets = activePresets()
    const eased = ease(progress)
    const scale = lerp(currentPresets.startScale, currentPresets.endScale, eased)

    root.style.setProperty("--scene-progress", eased.toFixed(3))
    root.style.setProperty("--scene-scale", scale.toFixed(3))
  }

  function measure() {
    const rect = track.getBoundingClientRect()
    const top = window.scrollY + rect.top
    const stickyHeight = Math.max(1, window.innerHeight - navHeight)

    metrics = {
      start: top,
      end: top + Math.max(1, track.offsetHeight - stickyHeight),
    }
  }

  function applyReducedMotionState() {
    root.dataset.sceneMotion = "reduced"
    root.style.setProperty("--scene-progress", "1")
    root.style.setProperty("--scene-scale", "1")
  }

  function update() {
    ticking = false

    if (prefersReducedMotion.matches) {
      applyReducedMotionState()
      return
    }

    delete root.dataset.sceneMotion

    const span = Math.max(1, metrics.end - metrics.start)
    const progress = clamp((window.scrollY - metrics.start) / span)

    setState(progress)
  }

  function schedule() {
    if (ticking) {
      return
    }

    ticking = true
    window.requestAnimationFrame(update)
  }

  function refresh() {
    measure()
    schedule()
  }

  window.addEventListener("scroll", schedule, { passive: true })
  window.addEventListener("resize", refresh, { passive: true })
  window.addEventListener("load", refresh)

  if (typeof ResizeObserver !== "undefined") {
    const observer = new ResizeObserver(refresh)
    observer.observe(track)
  }

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", refresh)
  }

  refresh()
})()
