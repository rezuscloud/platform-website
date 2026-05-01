(() => {
  const root = document.querySelector("[data-scene-root]");
  if (!root) return;

  const track = root.querySelector(".snap-track");
  if (!track) return;

  const panels = Array.from(track.querySelectorAll(".snap-panel"));
  const dots = Array.from(root.querySelectorAll(".snap-dot"));
  if (panels.length === 0) return;

  const prefersReducedMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  );
  let currentIndex = 0;
  let isScrolling = false;
  let scrollTimeout = null;
  let wheelAccum = 0;
  const WHEEL_THRESHOLD = 60;
  const WHEEL_COOLDOWN = 600;

  function updateDots(index) {
    dots.forEach((dot, i) => {
      dot.classList.toggle("is-active", i === index);
    });
  }

  function scrollToPanel(index) {
    if (index < 0 || index >= panels.length) return;
    if (index === currentIndex && isScrolling) return;

    isScrolling = true;
    currentIndex = index;
    updateDots(index);

    const panel = panels[index];
    const offset = panel.offsetLeft;

    if (prefersReducedMotion.matches || window.innerWidth <= 1100) {
      panel.scrollIntoView({ behavior: "auto", block: "nearest", inline: "start" });
      setTimeout(() => { isScrolling = false; }, 100);
      return;
    }

    track.scrollTo({
      left: offset,
      behavior: "smooth",
    });

    setTimeout(() => { isScrolling = false; }, WHEEL_COOLDOWN);
  }

  function handleWheel(e) {
    if (window.innerWidth <= 1100) return;
    if (isScrolling) {
      e.preventDefault();
      return;
    }

    const delta = Math.abs(e.deltaX) > Math.abs(e.deltaY) ? e.deltaX : e.deltaY;
    wheelAccum += delta;

    if (Math.abs(wheelAccum) >= WHEEL_THRESHOLD) {
      e.preventDefault();
      const direction = wheelAccum > 0 ? 1 : -1;
      wheelAccum = 0;
      scrollToPanel(currentIndex + direction);
    }

    clearTimeout(scrollTimeout);
    scrollTimeout = setTimeout(() => { wheelAccum = 0; }, 200);
  }

  function handleKey(e) {
    if (window.innerWidth <= 1100) return;
    if (e.key === "ArrowRight" || e.key === "ArrowDown") {
      e.preventDefault();
      scrollToPanel(currentIndex + 1);
    }
    if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
      e.preventDefault();
      scrollToPanel(currentIndex - 1);
    }
  }

  function handleDotClick(e) {
    e.preventDefault();
    const dot = e.currentTarget;
    const index = parseInt(dot.dataset.snapGoto, 10);
    if (!isNaN(index)) scrollToPanel(index);
  }

  function handleScrollEnd() {
    if (!track) return;
    const scrollLeft = track.scrollLeft;
    let closest = 0;
    let closestDist = Infinity;
    panels.forEach((panel, i) => {
      const dist = Math.abs(panel.offsetLeft - scrollLeft);
      if (dist < closestDist) {
        closestDist = dist;
        closest = i;
      }
    });
    if (closest !== currentIndex) {
      currentIndex = closest;
      updateDots(closest);
    }
  }

  function handleTouchStart(e) {
    if (window.innerWidth <= 1100) return;
    const touch = e.touches[0];
    track._touchStartX = touch.clientX;
    track._touchStartY = touch.clientY;
  }

  function handleTouchEnd(e) {
    if (window.innerWidth <= 1100) return;
    if (!track._touchStartX) return;
    const touch = e.changedTouches[0];
    const dx = touch.clientX - track._touchStartX;
    const dy = touch.clientY - track._touchStartY;

    if (Math.abs(dx) > Math.abs(dy) && Math.abs(dx) > 50) {
      if (dx < 0) scrollToPanel(currentIndex + 1);
      else scrollToPanel(currentIndex - 1);
    }
    track._touchStartX = null;
  }

  root.addEventListener("wheel", handleWheel, { passive: false });
  root.addEventListener("keydown", handleKey);
  dots.forEach((dot) => dot.addEventListener("click", handleDotClick));
  track.addEventListener("touchstart", handleTouchStart, { passive: true });
  track.addEventListener("touchend", handleTouchEnd, { passive: true });

  let scrollEndTimer = null;
  track.addEventListener("scroll", () => {
    clearTimeout(scrollEndTimer);
    scrollEndTimer = setTimeout(handleScrollEnd, 100);
  }, { passive: true });

  if (typeof prefersReducedMotion.addEventListener === "function") {
    prefersReducedMotion.addEventListener("change", () => scrollToPanel(currentIndex));
  }

  document.body.addEventListener("htmx:afterSwap", () => {
    requestAnimationFrame(() => scrollToPanel(currentIndex));
  });

  updateDots(0);

  if (typeof ResizeObserver !== "undefined") {
    new ResizeObserver(() => scrollToPanel(currentIndex)).observe(track);
  }
})();
