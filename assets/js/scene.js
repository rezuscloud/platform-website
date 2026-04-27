(() => {
  const root = document.querySelector("[data-scene-root]")
  if (!root) {
    return
  }

  const chapters = Array.from(root.querySelectorAll("[data-scene-chapter]"))
  if (chapters.length < 3) {
    return
  }

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)")

  const presets = {
    desktop: {
      terminal: { scale: 2.45, x: 14, y: 8 },
      mac: { scale: 1.52, x: 5, y: 3 },
      linux: { scale: 1.0, x: 0, y: 0 },
    },
    mobile: {
      terminal: { scale: 2.0, x: 9, y: 7 },
      mac: { scale: 1.28, x: 3, y: 2 },
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
  }

  function measure() {
    metrics = chapters.map((chapter) => {
      const rect = chapter.getBoundingClientRect()
      const top = window.scrollY + rect.top
      return {
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
    const terminalExit = metrics[1].start
    const linuxEntry = metrics[2].start

    let state = currentPresets.terminal

    if (focusPoint < terminalExit) {
      const amount = ease(progressBetween(metrics[0].start, terminalExit, focusPoint))
      state = lerpPreset(currentPresets.terminal, currentPresets.mac, amount)
      setProgress("terminal", 1)
      setProgress("mac", amount)
      setProgress("linux", 0)
    } else {
      const amount = ease(progressBetween(terminalExit, linuxEntry, focusPoint))
      state = lerpPreset(currentPresets.mac, currentPresets.linux, amount)
      setProgress("terminal", 1)
      setProgress("mac", 1)
      setProgress("linux", amount)
    }

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
