(() => {
  const root = document.querySelector("[data-scene-root]")
  if (!root) return

  const track = root.querySelector("[data-scene-track]")
  const camera = root.querySelector("[data-scene-camera]")
  const world = root.querySelector("[data-scene-world]")
  if (!track || !camera || !world) return

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)")
  const targetNames = ["terminal", "mac", "linux"]

  let viewport = { width: 1, height: 1 }
  let metrics = { start: 0, span: 1 }
  let targets = new Map()
  let ticking = false

  function clamp(value, min = 0, max = 1) {
    return Math.min(max, Math.max(min, value))
  }

  function lerp(a, b, t) {
    return a + (b - a) * t
  }

  function easeOutExpo(value) {
    if (value <= 0) return 0
    if (value >= 1) return 1
    return 1 - Math.pow(2, -10 * value)
  }

  function blendRect(a, b, t) {
    return {
      left: lerp(a.left, b.left, t),
      top: lerp(a.top, b.top, t),
      width: lerp(a.width, b.width, t),
      height: lerp(a.height, b.height, t),
    }
  }

  function fitRect(rect, overscan) {
    const scale = Math.max(viewport.width / rect.width, viewport.height / rect.height)
    const fittedScale = scale * overscan
    const translateX = viewport.width / 2 - (rect.left + rect.width / 2) * fittedScale
    const translateY = viewport.height / 2 - (rect.top + rect.height / 2) * fittedScale

    return {
      scale: fittedScale,
      translateX,
      translateY,
    }
  }

  function worldRectFor(element) {
    const worldBox = world.getBoundingClientRect()
    const box = element.getBoundingClientRect()
    const scaleX = world.offsetWidth / worldBox.width
    const scaleY = world.offsetHeight / worldBox.height

    return {
      left: (box.left - worldBox.left) * scaleX,
      top: (box.top - worldBox.top) * scaleY,
      width: box.width * scaleX,
      height: box.height * scaleY,
    }
  }

  function captureTargets() {
    targets = new Map()
    for (const name of targetNames) {
      const element = root.querySelector(`[data-scene-target="${name}"]`)
      if (!element) continue
      targets.set(name, worldRectFor(element))
    }
  }

  function computeProgress(scrollY) {
    const raw = clamp((scrollY - metrics.start) / metrics.span)

    if (raw < 0.14) return { phase: 0, mix: 0, overall: 0 }
    if (raw < 0.44) {
      const local = easeOutExpo((raw - 0.14) / 0.3)
      return { phase: 0, mix: local, overall: raw }
    }
    if (raw < 0.58) return { phase: 1, mix: 0, overall: raw }
    if (raw < 0.9) {
      const local = easeOutExpo((raw - 0.58) / 0.32)
      return { phase: 1, mix: local, overall: raw }
    }

    return { phase: 2, mix: 0, overall: raw }
  }

  function activeRectFor(state) {
    const terminal = targets.get("terminal")
    const mac = targets.get("mac")
    const linux = targets.get("linux")
    if (!terminal || !mac || !linux) return null

    if (state.phase === 0 && state.mix > 0) return blendRect(terminal, mac, state.mix)
    if (state.phase === 1 && state.mix === 0) return mac
    if (state.phase === 1 && state.mix > 0) return blendRect(mac, linux, state.mix)
    if (state.phase === 2) return linux
    return terminal
  }

  function overscanFor(state) {
    if (state.phase === 0 && state.mix === 0) return 1.08
    if (state.phase === 0) return 1.04
    if (state.phase === 1 && state.mix === 0) return 1.02
    return 1
  }

  function applyScene(scrollY) {
    if (prefersReducedMotion.matches || window.innerWidth <= 1100) {
      root.dataset.sceneMotion = "reduced"
      root.style.setProperty("--scene-progress", "1")
      root.style.setProperty("--scene-scale", "1")
      root.style.setProperty("--scene-translate-x", "0px")
      root.style.setProperty("--scene-translate-y", "0px")
      return
    }

    delete root.dataset.sceneMotion

    const state = computeProgress(scrollY)
    const rect = activeRectFor(state)
    if (!rect) return

    const fitted = fitRect(rect, overscanFor(state))
    root.style.setProperty("--scene-progress", state.overall.toFixed(4))
    root.style.setProperty("--scene-scale", fitted.scale.toFixed(5))
    root.style.setProperty("--scene-translate-x", `${Math.round(fitted.translateX)}px`)
    root.style.setProperty("--scene-translate-y", `${Math.round(fitted.translateY)}px`)
  }

  function measure() {
    viewport = {
      width: window.innerWidth,
      height: window.innerHeight,
    }

    const rect = track.getBoundingClientRect()
    const top = window.scrollY + rect.top
    metrics = {
      start: top,
      span: Math.max(1, track.offsetHeight - window.innerHeight),
    }

    captureTargets()
    applyScene(window.scrollY)
  }

  function update() {
    ticking = false
    applyScene(window.scrollY)
  }

  function schedule() {
    if (ticking) return
    ticking = true
    requestAnimationFrame(update)
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
    observer.observe(world)
    for (const name of targetNames) {
      const element = root.querySelector(`[data-scene-target="${name}"]`)
      if (element) observer.observe(element)
    }
  }

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", refresh)
  }

  document.body.addEventListener("htmx:afterSwap", () => {
    requestAnimationFrame(refresh)
  })

  refresh()
})()
