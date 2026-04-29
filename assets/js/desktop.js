(() => {
  const SERVICE_INFO = {
    "shell-app":
      "Renders the proof rail from PostgreSQL-backed shared state. Watches for events on homepage.events via NATS JetStream and updates the summary, proof chips, and event ribbon.",
    "terminal-app":
      "Collects intent and dispatches flows to linux-app via Dapr service invocation. Commands are processed server-side with real PostgreSQL state, Redis locking, and NATS eventing.",
    "mac-app":
      "The inspection surface. Reads PostgreSQL-backed shared state to render the latest artifact or dossier. It never copies or invents data, only reads what linux-app persisted.",
    "linux-app":
      "Owns execution. Accepts invokes, manages shared topology, persists state in PostgreSQL V2, serializes updates through Redis lockstore, and publishes events over NATS JetStream.",
  };

  const ICON_INFO = {
    "Macintosh HD":
      "PostgreSQL V2 State Store\nSession data persisted at homepage/sessions/{id}/state. Backed by PostgreSQL with the Dapr state management API (v2).",
    System:
      "Dapr Sidecar\nService invocation, pub/sub, state management, and distributed locking. Components: PostgreSQL (state), NATS JetStream (pubsub), Redis (locking).",
    Artifacts: "Open the artifact viewer to inspect the latest dossier published by linux-app.",
  };

  const MENU_ITEMS = {
    File: [
      { label: "Open State", info: "Opens the PostgreSQL-backed session state viewer." },
      { label: "Close Window", info: null },
    ],
    Edit: [
      { label: "Copy Session ID", info: "Copies the current session identifier to clipboard." },
    ],
    View: [
      { label: "Refresh State", info: "Re-reads shared state from PostgreSQL via Dapr state API." },
      { label: "View Source", info: "Shows the raw JSON state document for this session." },
    ],
    Special: [
      { label: "Inspect Dapr", info: "Dapr sidecar: PostgreSQL state, NATS JetStream pubsub, Redis lockstore." },
      { label: "Check Topology", info: "Same-origin topology: shell, terminal, mac, linux apps sharing one state store." },
      { label: "Publish Event", info: "linux-app publishes to homepage.events on NATS JetStream after each command." },
    ],
  };

  let activeMenu = null;

  function closeMenus() {
    document.querySelectorAll(".mac-dropdown").forEach((d) => d.remove());
    activeMenu = null;
  }

  function showMenu(anchor, items) {
    closeMenus();
    const parent = anchor.closest(".mac-menubar-shell");
    if (!parent) return;
    const dd = document.createElement("div");
    dd.className = "mac-dropdown";
    const rect = anchor.getBoundingClientRect();
    const parentRect = parent.getBoundingClientRect();
    dd.style.left = rect.left - parentRect.left + "px";
    dd.style.top = rect.bottom - parentRect.top + "px";
    items.forEach((item) => {
      const row = document.createElement("button");
      row.className = "mac-dropdown-item";
      row.textContent = item.label;
      row.type = "button";
      row.addEventListener("click", (e) => {
        e.stopPropagation();
        closeMenus();
        if (item.info) showToast(anchor, item.info);
      });
      dd.appendChild(row);
    });
    parent.style.position = "relative";
    parent.appendChild(dd);
    activeMenu = dd;
  }

  function showToast(anchor, text) {
    closeMenus();
    const desktop = anchor.closest(".mac-panel") || anchor.closest(".mac-desktop-shell");
    if (!desktop) return;
    const existing = desktop.querySelector(".mac-toast");
    if (existing) existing.remove();
    const toast = document.createElement("div");
    toast.className = "mac-toast";
    toast.textContent = text;
    desktop.style.position = "relative";
    desktop.appendChild(toast);
    setTimeout(() => toast.remove(), 4000);
  }

  function showIconDialog(desktop, title, text) {
    const existing = desktop.querySelector(".mac-dialog");
    if (existing) existing.remove();
    const dlg = document.createElement("div");
    dlg.className = "mac-dialog mac-primary-window";
    const titlebar = document.createElement("div");
    titlebar.className = "mac-titlebar-shell";
    titlebar.innerHTML =
      '<div class="mac-titlebar-box mac-dialog-close" style="cursor:pointer"></div><strong>' +
      title +
      "</strong><div class=\"mac-titlebar-box\"></div>";
    const body = document.createElement("div");
    body.className = "mac-window-body-shell";
    const lines = text.split("\n");
    lines.forEach((l) => {
      const p = document.createElement("p");
      p.style.cssText = "margin:0;font-size:0.8rem;line-height:1.35;color:oklch(0.24 0.008 96)";
      p.textContent = l;
      body.appendChild(p);
    });
    dlg.appendChild(titlebar);
    dlg.appendChild(body);
    desktop.appendChild(dlg);
    titlebar.querySelector(".mac-dialog-close").addEventListener("click", () => dlg.remove());
  }

  function showServiceDetail(row) {
    const existing = row.querySelector(".linux-service-expand");
    if (existing) {
      existing.remove();
      row.classList.remove("is-expanded");
      return;
    }
    row.classList.add("is-expanded");
    const name = row.querySelector(".linux-service-name");
    if (!name) return;
    const info = SERVICE_INFO[name.textContent.trim()];
    if (!info) return;
    const el = document.createElement("div");
    el.className = "linux-service-expand";
    el.textContent = info;
    row.appendChild(el);
  }

  function initLinux() {
    document.querySelectorAll(".linux-service-row").forEach((row) => {
      if (row.dataset.dtInit) return;
      row.dataset.dtInit = "1";
      row.style.cursor = "pointer";
      row.addEventListener("click", () => showServiceDetail(row));
    });

    document.querySelectorAll(".linux-titlebar-controls").forEach((ctrls) => {
      const win = ctrls.closest(".linux-window");
      if (!win || win.dataset.dtInit) return;
      win.dataset.dtInit = "1";
      const btns = ctrls.querySelectorAll(".linux-control");
      if (btns[0]) {
        btns[0].style.cursor = "pointer";
        btns[0].addEventListener("click", (e) => {
          e.stopPropagation();
          win.classList.toggle("is-minimized");
        });
      }
      if (btns[1]) {
        btns[1].style.cursor = "pointer";
        btns[1].addEventListener("click", (e) => {
          e.stopPropagation();
          win.classList.toggle("is-maximized");
        });
      }
    });
  }

  function initMac() {
    document.querySelectorAll(".desktop-icon").forEach((icon) => {
      if (icon.dataset.dtInit) return;
      icon.dataset.dtInit = "1";
      icon.style.cursor = "pointer";
      icon.addEventListener("dblclick", (e) => {
        e.preventDefault();
        const label = icon.querySelector(".desktop-icon-label");
        if (!label) return;
        const name = label.textContent.trim();
        const info = ICON_INFO[name];
        if (!info) return;
        const desktop = icon.closest(".mac-desktop-shell");
        if (desktop) showIconDialog(desktop, name, info);
      });
    });

    document.querySelectorAll(".mac-menubar-left > span").forEach((item) => {
      if (item.classList.contains("mac-apple-mark")) return;
      if (item.dataset.dtInit) return;
      item.dataset.dtInit = "1";
      item.style.cursor = "pointer";
      const menuName = item.classList.contains("mac-menu-strong")
        ? item.textContent.trim()
        : item.textContent.trim();
      const items = MENU_ITEMS[menuName];
      if (!items) return;
      item.addEventListener("click", (e) => {
        e.stopPropagation();
        if (activeMenu) {
          closeMenus();
          return;
        }
        showMenu(item, items);
      });
    });

    document.querySelectorAll(".mac-artifact-window").forEach((w) => {
      if (w.dataset.dtInit) return;
      w.dataset.dtInit = "1";
      const titlebar = w.querySelector(".mac-titlebar-shell");
      if (titlebar) {
        titlebar.style.cursor = "pointer";
        titlebar.addEventListener("click", () => w.classList.toggle("is-expanded"));
      }
    });
  }

  function init() {
    initLinux();
    initMac();
  }

  document.addEventListener("click", (e) => {
    if (activeMenu && !e.target.closest(".mac-dropdown") && !e.target.closest(".mac-menubar-left")) {
      closeMenus();
    }
  });

  document.body.addEventListener("htmx:afterSwap", () => {
    requestAnimationFrame(init);
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
