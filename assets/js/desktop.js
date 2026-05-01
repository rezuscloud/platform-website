(() => {
  var SERVICE_INFO = {
    "shell-app":
      "Renders the proof rail from PostgreSQL-backed shared state. Watches for events on homepage.events via NATS JetStream and updates the summary, proof chips, and event ribbon.",
    "terminal-app":
      "Collects intent and dispatches flows to linux-app via Dapr service invocation. Commands are processed server-side with real PostgreSQL state, Redis locking, and NATS eventing.",
    "mac-app":
      "The inspection surface. Reads PostgreSQL-backed shared state to render the latest artifact or dossier. It never copies or invents data, only reads what linux-app persisted.",
    "linux-app":
      "Owns execution. Accepts invokes, manages shared topology, persists state in PostgreSQL V2, serializes updates through Redis lockstore, and publishes events over NATS JetStream.",
  };

  var ICON_INFO = {
    "Macintosh HD":
      "PostgreSQL V2 State Store\nSession data persisted at homepage/sessions/{id}/state.",
    System:
      "Dapr Sidecar\nService invocation, pub/sub, state management, and distributed locking.",
    Artifacts:
      "Open the artifact viewer to inspect the latest dossier published by linux-app.",
  };

  var MENU_ITEMS = {
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

  var activeMenu = null;

  function closeMenus() {
    var dropdowns = document.querySelectorAll(".mac-dropdown");
    for (var i = 0; i < dropdowns.length; i++) dropdowns[i].remove();
    activeMenu = null;
  }

  function showMenu(anchor, items) {
    closeMenus();
    var parent = anchor.closest(".mac-menubar-shell");
    if (!parent) return;
    var dd = document.createElement("div");
    dd.className = "mac-dropdown";
    var rect = anchor.getBoundingClientRect();
    var parentRect = parent.getBoundingClientRect();
    dd.style.left = rect.left - parentRect.left + "px";
    dd.style.top = rect.bottom - parentRect.top + "px";
    for (var i = 0; i < items.length; i++) {
      (function (item) {
        var row = document.createElement("button");
        row.className = "mac-dropdown-item";
        row.textContent = item.label;
        row.type = "button";
        row.addEventListener("click", function (e) {
          e.stopPropagation();
          closeMenus();
          if (item.info) showToast(anchor, item.info);
        });
        dd.appendChild(row);
      })(items[i]);
    }
    parent.style.position = "relative";
    parent.appendChild(dd);
    activeMenu = dd;
  }

  function showToast(anchor, text) {
    closeMenus();
    var desktop = anchor.closest(".mac-panel") || anchor.closest(".mac-desktop-shell");
    if (!desktop) return;
    var existing = desktop.querySelector(".mac-toast");
    if (existing) existing.remove();
    var toast = document.createElement("div");
    toast.className = "mac-toast";
    toast.textContent = text;
    desktop.style.position = "relative";
    desktop.appendChild(toast);
    setTimeout(function () { toast.remove(); }, 4000);
  }

  function showIconDialog(desktop, title, text) {
    var existing = desktop.querySelector(".mac-dialog");
    if (existing) existing.remove();
    var dlg = document.createElement("div");
    dlg.className = "mac-dialog mac-primary-window";
    var titlebar = document.createElement("div");
    titlebar.className = "mac-titlebar-shell";
    titlebar.innerHTML =
      '<div class="mac-titlebar-box mac-dialog-close" style="cursor:pointer"></div><strong>' +
      title +
      '</strong><div class="mac-titlebar-box"></div>';
    var body = document.createElement("div");
    body.className = "mac-window-body-shell";
    var lines = text.split("\n");
    for (var i = 0; i < lines.length; i++) {
      var p = document.createElement("p");
      p.style.cssText = "margin:0;font-size:0.8rem;line-height:1.35;color:oklch(0.24 0.008 96)";
      p.textContent = lines[i];
      body.appendChild(p);
    }
    dlg.appendChild(titlebar);
    dlg.appendChild(body);
    desktop.appendChild(dlg);
    titlebar.querySelector(".mac-dialog-close").addEventListener("click", function () { dlg.remove(); });
  }

  function initLinux() {
    var rows = document.querySelectorAll(".linux-service-row");
    for (var i = 0; i < rows.length; i++) {
      var row = rows[i];
      if (row.dataset.dtInit) continue;
      row.dataset.dtInit = "1";
      row.style.cursor = "pointer";
      (function (r) {
        r.addEventListener("click", function () {
          var existing = r.querySelector(".linux-service-expand");
          if (existing) {
            existing.remove();
            r.classList.remove("is-expanded");
            return;
          }
          r.classList.add("is-expanded");
          var nameEl = r.querySelector(".linux-service-name");
          if (!nameEl) return;
          var info = SERVICE_INFO[nameEl.textContent.trim()];
          if (!info) return;
          var el = document.createElement("div");
          el.className = "linux-service-expand";
          el.textContent = info;
          r.appendChild(el);
        });
      })(row);
    }

    var ctrls = document.querySelectorAll(".linux-titlebar-controls");
    for (var j = 0; j < ctrls.length; j++) {
      var ctrl = ctrls[j];
      var win = ctrl.closest(".linux-window");
      if (!win || win.dataset.dtInit) continue;
      win.dataset.dtInit = "1";
      var btns = ctrl.querySelectorAll(".linux-control");
      if (btns[0]) {
        btns[0].style.cursor = "pointer";
        (function (w) {
          btns[0].addEventListener("click", function (e) {
            e.stopPropagation();
            w.classList.toggle("is-minimized");
            w.classList.remove("is-maximized");
          });
        })(win);
      }
      if (btns[1]) {
        btns[1].style.cursor = "pointer";
        (function (w) {
          btns[1].addEventListener("click", function (e) {
            e.stopPropagation();
            w.classList.toggle("is-maximized");
            w.classList.remove("is-minimized");
          });
        })(win);
      }
    }
  }

  function initMac() {
    var icons = document.querySelectorAll(".desktop-icon");
    for (var i = 0; i < icons.length; i++) {
      var icon = icons[i];
      if (icon.dataset.dtInit) continue;
      icon.dataset.dtInit = "1";
      icon.style.cursor = "pointer";
      (function (ic) {
        ic.addEventListener("dblclick", function (e) {
          e.preventDefault();
          var label = ic.getAttribute("data-icon") || "";
          var info = ICON_INFO[label];
          if (!info) return;
          var desktop = ic.closest(".mac-desktop-shell");
          if (desktop) showIconDialog(desktop, label, info);
        });
      })(icon);
    }

    var items = document.querySelectorAll(".mac-menubar-left > span");
    for (var j = 0; j < items.length; j++) {
      var item = items[j];
      if (item.classList.contains("mac-apple-mark")) continue;
      if (item.dataset.dtInit) continue;
      item.dataset.dtInit = "1";
      item.style.cursor = "pointer";
      (function (el) {
        var menuName = el.textContent.trim();
        var menuItems = MENU_ITEMS[menuName];
        if (!menuItems) return;
        el.addEventListener("click", function (e) {
          e.stopPropagation();
          if (activeMenu) {
            closeMenus();
            return;
          }
          showMenu(el, menuItems);
        });
      })(item);
    }

    var artWindows = document.querySelectorAll(".mac-artifact-window");
    for (var k = 0; k < artWindows.length; k++) {
      var w = artWindows[k];
      if (w.dataset.dtInit) continue;
      w.dataset.dtInit = "1";
      var titlebar = w.querySelector(".mac-titlebar-shell");
      if (titlebar) {
        titlebar.style.cursor = "pointer";
        (function (win) {
          titlebar.addEventListener("click", function () {
            win.classList.toggle("is-expanded");
          });
        })(w);
      }
    }
  }

  function init() {
    initLinux();
    initMac();
  }

  document.addEventListener("click", function (e) {
    if (activeMenu && !e.target.closest(".mac-dropdown") && !e.target.closest(".mac-menubar-left")) {
      closeMenus();
    }
  });

  document.body.addEventListener("htmx:afterSwap", function () {
    requestAnimationFrame(init);
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
