(() => {
  const root = document.querySelector("[data-scene-root]")
  if (!root) {
    return
  }

  const chapters = Array.from(root.querySelectorAll("[data-scene-chapter]"))
  const navLinks = Array.from(document.querySelectorAll("[data-story-link]"))
  if (chapters.length < 3) {
    return
  }

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)")

  const presets = {
    desktop: {
      terminal: { scale: 2.7, x: 15, y: 9 },
      mac: { scale: 1.62, x: 4, y: 3 },
      linux: { scale: 1.0, x: 0, y: 0 },
    },
    mobile: {
      terminal: { scale: 2.18, x: 10, y: 7 },
      mac: { scale: 1.34, x: 3, y: 2 },
      linux: { scale: 0.94, x: 0, y: 0 },
    },
  }

  let metrics = []
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

  function lerpPreset(start, end, amount) {
    return {
      scale: lerp(start.scale, end.scale, amount),
      x: lerp(start.x, end.x, amount),
      y: lerp(start.y, end.y, amount),
    }
  }

  function activePresets() {
    return window.innerWidth < 768 ? presets.mobile : presets.desktop
  }

  function setState(state) {
    root.style.setProperty("--scene-scale", state.scale.toFixed(3))
    root.style.setProperty("--scene-x", `${state.x.toFixed(3)}%`)
    root.style.setProperty("--scene-y", `${state.y.toFixed(3)}%`)
  }

  function setProgress(name, value) {
    root.style.setProperty(`--${name}-progress`, clamp(value).toFixed(3))
  }

  function progressBetween(start, end, position) {
    const span = Math.max(1, end - start)
    return clamp((position - start) / span)
  }

  function markActiveChapter(position) {
    chapters.forEach((chapter, index) => {
      const next = metrics[index + 1]
      const isActive = position >= metrics[index].start && (!next || position < next.start)
      chapter.toggleAttribute("data-scene-active", isActive)
    })

    const active = chapters.find((chapter) => chapter.hasAttribute("data-scene-active"))
    const activeName = active?.dataset.sceneChapter
    navLinks.forEach((link) => {
      link.toggleAttribute("data-story-active", link.dataset.storyLink === activeName)
    })
  }

  function measure() {
    metrics = chapters.map((chapter) => {
      const rect = chapter.getBoundingClientRect()
      const top = window.scrollY + rect.top
      return {
        element: chapter,
        start: top,
        end: top + chapter.offsetHeight,
      }
    })
  }

  function applyReducedMotionState() {
    root.dataset.sceneMotion = "reduced"
    setState(activePresets().linux)
    setProgress("terminal", 1)
    setProgress("mac", 1)
    setProgress("linux", 1)
  }

  function update() {
    ticking = false

    if (prefersReducedMotion.matches) {
      applyReducedMotionState()
      return
    }

    delete root.dataset.sceneMotion

    const currentPresets = activePresets()
    const focusPoint = window.scrollY + window.innerHeight * 0.5
    const terminalToMac = ease(progressBetween(metrics[0].start, metrics[1].start, focusPoint))
    const macToLinux = ease(progressBetween(metrics[1].start, metrics[2].start, focusPoint))

    let state = currentPresets.terminal

    if (macToLinux > 0) {
      state = lerpPreset(currentPresets.mac, currentPresets.linux, macToLinux)
    } else {
      state = lerpPreset(currentPresets.terminal, currentPresets.mac, terminalToMac)
    }

    setProgress("terminal", 1 - terminalToMac * 0.18)
    setProgress("mac", terminalToMac)
    setProgress("linux", macToLinux)

    markActiveChapter(focusPoint)
    setState(state)
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
    chapters.forEach((chapter) => observer.observe(chapter))
  }

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", refresh)
  }

  refresh()
})()
