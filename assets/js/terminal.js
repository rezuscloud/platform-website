(() => {
  const COMMANDS = [
    "rezus sync demo",
    "rezus fanout edge",
    "rezus inspect dossier",
  ];
  let history = [];
  let histIdx = -1;
  let buffer = "";
  let focused = false;
  let submitting = false;

  function q(sel, root) {
    return (root || document).querySelector(sel);
  }

  function screen() {
    return q(".terminal-screen-shell");
  }

  function scrollback() {
    return q(".terminal-scrollback");
  }

  function inputSpan() {
    return q("[data-term-input]");
  }

  function hiddenInput() {
    return q("[data-term-cmd]");
  }

  function promptLine() {
    return q("[data-term-prompt]");
  }

  function form() {
    return q(".terminal-form");
  }

  function syncPrompt() {
    const span = inputSpan();
    if (span) span.textContent = buffer;
  }

  function scrollEnd() {
    const sb = scrollback();
    if (sb) sb.scrollTop = sb.scrollHeight;
  }

  function focusOn() {
    focused = true;
    const s = screen();
    if (s) s.classList.add("term-focused");
  }

  function focusOff() {
    focused = false;
    const s = screen();
    if (s) s.classList.remove("term-focused");
  }

  function submit(cmd) {
    const f = form();
    const hi = hiddenInput();
    if (!f || !hi || submitting) return;

    const trimmed = cmd.trim();
    if (trimmed) {
      history.push(trimmed);
      histIdx = history.length;
    }

    hi.value = trimmed;
    submitting = true;

    const pl = promptLine();
    if (pl) pl.setAttribute("data-term-busy", "");

    const sb = scrollback();
    if (sb) {
      const wait = document.createElement("p");
      wait.className = "term-waiting";
      wait.textContent = "\u2318 processing\u2026";
      sb.appendChild(wait);
      scrollEnd();
    }

    const parent = q("[data-term-parent]");
    if (parent) parent.value = currentHostRoute();

    f.requestSubmit();
  }

  function currentHostRoute() {
    const el = q("[data-term-host]");
    return el ? el.value : "";
  }

  function onKey(e) {
    if (!focused || submitting) return;

    const tag = (document.activeElement || {}).tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

    switch (e.key) {
      case "Enter":
        e.preventDefault();
        const cmd = buffer;
        buffer = "";
        syncPrompt();
        submit(cmd);
        break;
      case "ArrowUp":
        e.preventDefault();
        if (histIdx > 0) {
          histIdx--;
          buffer = history[histIdx];
          syncPrompt();
        }
        break;
      case "ArrowDown":
        e.preventDefault();
        if (histIdx < history.length - 1) {
          histIdx++;
          buffer = history[histIdx];
        } else {
          histIdx = history.length;
          buffer = "";
        }
        syncPrompt();
        break;
      case "Tab":
        e.preventDefault();
        for (const c of COMMANDS) {
          if (c.startsWith(buffer) && c !== buffer) {
            buffer = c;
            syncPrompt();
            break;
          }
        }
        break;
      case "Backspace":
        e.preventDefault();
        buffer = buffer.slice(0, -1);
        syncPrompt();
        break;
      case "Escape":
        e.preventDefault();
        buffer = "";
        histIdx = history.length;
        syncPrompt();
        break;
      default:
        if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
          e.preventDefault();
          buffer += e.key;
          syncPrompt();
        }
    }

    scrollEnd();
  }

  function onSuggestionClick(e) {
    const btn = e.target.closest(".terminal-suggestions button");
    if (!btn) return;
    e.preventDefault();
    const hi = hiddenInput();
    if (hi) hi.value = "";
    const cmd = btn.getAttribute("data-cmd") || btn.value || btn.textContent.trim();
    buffer = "";
    syncPrompt();
    submit(cmd);
  }

  function init() {
    const s = screen();
    if (!s) return;

    s.addEventListener("click", focusOn);
    s.addEventListener("focus", focusOn);

    const pl = promptLine();
    if (pl) pl.removeAttribute("data-term-busy");

    submitting = false;
    scrollEnd();
  }

  function teardown() {
    const s = screen();
    if (s) {
      s.removeEventListener("click", focusOn);
      s.removeEventListener("focus", focusOn);
    }
  }

  document.addEventListener("click", (e) => {
    const s = screen();
    if (s && !s.contains(e.target)) focusOff();
  });

  document.addEventListener("keydown", onKey);
  document.addEventListener("click", onSuggestionClick);

  document.body.addEventListener("htmx:afterSwap", (e) => {
    const target = e.detail && e.detail.target;
    if (
      target &&
      (target.id === "terminal-panel" ||
        (target.querySelector && target.querySelector("#terminal-panel")))
    ) {
      teardown();
      requestAnimationFrame(init);
    }
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
