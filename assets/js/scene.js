(() => {
  const root = document.querySelector("[data-scene-root]");
  if (!root) return;

  const track = root.querySelector("[data-scene-track]");
  const camera = root.querySelector("[data-scene-camera]");
  const world = root.querySelector("[data-scene-world]");
  if (!track || !camera || !world) return;

  const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
  const targetNames = ["linux", "mac", "terminal"];

  let viewport = { width: 1, height: 1 };
  let metrics = { start: 0, span: 1 };
  let targets = new Map();
  let ticking = false;

  function clamp(v, lo, hi) {
    return Math.min(hi !== undefined ? hi : 1, Math.max(lo !== undefined ? lo : 0, v));
  }

  function lerp(a, b, t) {
    return a + (b - a) * t;
  }

  function easeOutExpo(v) {
    return v <= 0 ? 0 : v >= 1 ? 1 : 1 - Math.pow(2, -10 * v);
  }

  function blendRect(a, b, t) {
    return {
      left: lerp(a.left, b.left, t),
      top: lerp(a.top, b.top, t),
      width: lerp(a.width, b.width, t),
      height: lerp(a.height, b.height, t),
    };
  }

  function fitRect(rect, overscan) {
    var w = rect.width || 1;
    var h = rect.height || 1;
    var scale = Math.max(viewport.width / w, viewport.height / h) * (overscan || 1);
    return {
      scale: scale,
      translateX: viewport.width / 2 - (rect.left + w / 2) * scale,
      translateY: viewport.height / 2 - (rect.top + h / 2) * scale,
    };
  }

  function worldRectFor(el) {
    var left = 0;
    var top = 0;
    var node = el;
    while (node && node !== world) {
      left += node.offsetLeft || 0;
      top += node.offsetTop || 0;
      node = node.offsetParent;
    }
    return {
      left: left,
      top: top,
      width: el.offsetWidth || 1,
      height: el.offsetHeight || 1,
    };
  }

  function captureTargets() {
    targets = new Map();
    for (var i = 0; i < targetNames.length; i++) {
      var name = targetNames[i];
      var el = root.querySelector('[data-scene-target="' + name + '"]');
      if (!el) continue;
      var rect = worldRectFor(el);
      if (rect.width < 2 || rect.height < 2) continue;
      targets.set(name, rect);
    }
  }

  function computeProgress(scrollY) {
    var raw = clamp((scrollY - metrics.start) / metrics.span);

    if (raw < 0.15) return { phase: 0, mix: 0 };
    if (raw < 0.45) {
      return { phase: 0, mix: easeOutExpo((raw - 0.15) / 0.3) };
    }
    if (raw < 0.55) return { phase: 1, mix: 0 };
    if (raw < 0.85) {
      return { phase: 1, mix: easeOutExpo((raw - 0.55) / 0.3) };
    }
    return { phase: 2, mix: 0 };
  }

  function activeRectFor(state) {
    var linux = targets.get("linux");
    var mac = targets.get("mac");
    var terminal = targets.get("terminal");
    if (!linux) return null;

    if (state.phase === 0 && state.mix > 0) {
      if (mac) return blendRect(linux, mac, state.mix);
      return linux;
    }
    if (state.phase === 1 && state.mix === 0) {
      return mac || linux;
    }
    if (state.phase === 1 && state.mix > 0) {
      if (mac && terminal) return blendRect(mac, terminal, state.mix);
      return mac || linux;
    }
    if (state.phase === 2) {
      return terminal || mac || linux;
    }
    return linux;
  }

  function applyScene(scrollY) {
    if (prefersReducedMotion.matches || window.innerWidth <= 1100) {
      root.dataset.sceneMotion = "reduced";
      root.style.setProperty("--scene-scale", "1");
      root.style.setProperty("--scene-translate-x", "0px");
      root.style.setProperty("--scene-translate-y", "0px");
      return;
    }

    delete root.dataset.sceneMotion;

    var state = computeProgress(scrollY);
    var rect = activeRectFor(state);
    if (!rect) return;

    var fitted = fitRect(rect, 1);
    root.style.setProperty("--scene-scale", fitted.scale.toFixed(5));
    root.style.setProperty("--scene-translate-x", Math.round(fitted.translateX) + "px");
    root.style.setProperty("--scene-translate-y", Math.round(fitted.translateY) + "px");
  }

  function measure() {
    viewport = { width: window.innerWidth, height: window.innerHeight };

    var rect = track.getBoundingClientRect();
    metrics = {
      start: window.scrollY + rect.top,
      span: Math.max(1, track.offsetHeight - window.innerHeight),
    };

    captureTargets();
    applyScene(window.scrollY);
  }

  function update() {
    ticking = false;
    applyScene(window.scrollY);
  }

  function schedule() {
    if (ticking) return;
    ticking = true;
    requestAnimationFrame(update);
  }

  function refresh() {
    measure();
    schedule();
  }

  window.addEventListener("scroll", schedule, { passive: true });
  window.addEventListener("resize", refresh, { passive: true });
  window.addEventListener("load", refresh);

  if (typeof ResizeObserver !== "undefined") {
    var obs = new ResizeObserver(refresh);
    obs.observe(track);
    obs.observe(world);
    for (var i = 0; i < targetNames.length; i++) {
      var el = root.querySelector('[data-scene-target="' + targetNames[i] + '"]');
      if (el) obs.observe(el);
    }
  }

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", refresh);
  }

  document.body.addEventListener("htmx:afterSwap", function () {
    requestAnimationFrame(refresh);
  });

  refresh();
})();
