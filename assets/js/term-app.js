(() => {
  let term = null;
  let fitAddon = null;
  let history = [];
  let histIdx = -1;
  let currentLine = "";
  let prompt = "rezus@terminal $ ";
  let busy = false;

  function apiBase() {
    const el = document.querySelector("[data-term-api]");
    return el ? el.getAttribute("data-term-api") : "/apps/terminal/api/run";
  }

  function getMount() {
    return document.getElementById("xterm-mount");
  }

  function shouldInit() {
    return getMount() && typeof Terminal !== "undefined";
  }

  function create() {
    const mount = getMount();
    if (!mount || term) return;

    term = new Terminal({
      cursorBlink: true,
      cursorStyle: "block",
      fontSize: 18,
      fontFamily: '"VT323", "Courier New", monospace',
      lineHeight: 1.1,
      letterSpacing: 1,
      scrollback: 500,
      theme: {
        background: "#060d08",
        foreground: "#1cff77",
        cursor: "#1cff77",
        cursorAccent: "#060d08",
        selectionBackground: "rgba(28, 255, 119, 0.25)",
        black: "#060d08",
        red: "#ff5555",
        green: "#1cff77",
        yellow: "#f0c040",
        blue: "#5588ff",
        magenta: "#cc77ff",
        cyan: "#55ffdd",
        white: "#d4d4d4",
        brightBlack: "#444444",
        brightRed: "#ff8888",
        brightGreen: "#66ff99",
        brightYellow: "#ffdd66",
        brightBlue: "#88aaff",
        brightMagenta: "#dd99ff",
        brightCyan: "#88ffee",
        brightWhite: "#ffffff",
      },
      allowProposedApi: true,
    });

    if (typeof FitAddon !== "undefined") {
      fitAddon = new FitAddon.FitAddon();
      term.loadAddon(fitAddon);
    }

    term.open(mount);

    if (fitAddon) {
      requestAnimationFrame(() => {
        fitAddon.fit();
      });
    }

    const bootEl = document.querySelector("[data-term-boot]");
    if (bootEl) {
      const raw = bootEl.getAttribute("data-term-boot") || "[]";
      const lines = JSON.parse(raw);
      for (const line of lines) {
        term.writeln(line);
      }
    }

    term.writeln("");
    writePrompt();

    term.onData(handleInput);

    window.addEventListener("resize", () => {
      if (fitAddon) fitAddon.fit();
    });
  }

  function writePrompt() {
    term.write("\x1b[32m" + prompt + "\x1b[0m");
  }

  function clearCurrentLine() {
    if (currentLine.length > 0) {
      term.write("\x1b[" + currentLine.length + "D");
      term.write("\x1b[K");
    }
  }

  function handleInput(data) {
    if (busy) return;

    const code = data.charCodeAt(0);

    if (code === 13) {
      term.writeln("");
      const cmd = currentLine.trim();
      currentLine = "";
      histIdx = -1;

      if (cmd) {
        history.push(cmd);
        histIdx = history.length;
        runCommand(cmd);
      } else {
        writePrompt();
      }
      return;
    }

    if (code === 127 || code === 8) {
      if (currentLine.length > 0) {
        currentLine = currentLine.slice(0, -1);
        term.write("\b \b");
      }
      return;
    }

    if (data === "\x1b[A") {
      if (history.length === 0) return;
      clearCurrentLine();
      if (histIdx > 0) histIdx--;
      else if (histIdx === -1) histIdx = history.length - 1;
      currentLine = history[histIdx];
      term.write(currentLine);
      return;
    }

    if (data === "\x1b[B") {
      if (history.length === 0) return;
      clearCurrentLine();
      if (histIdx < history.length - 1) {
        histIdx++;
        currentLine = history[histIdx];
      } else {
        histIdx = history.length;
        currentLine = "";
      }
      term.write(currentLine);
      return;
    }

    if (data === "\x1b" || data === "\x18") {
      clearCurrentLine();
      currentLine = "";
      return;
    }

    if (code >= 32 && code < 127 && data.length === 1) {
      currentLine += data;
      term.write(data);
    }
  }

  function runCommand(command) {
    busy = true;
    term.write("\x1b[2m  \u2318 processing...\x1b[0m\r\n");

    const url = apiBase() + "?command=" + encodeURIComponent(command);

    fetch(url, { method: "POST", credentials: "same-origin" })
      .then((r) => {
        if (!r.ok) throw new Error("HTTP " + r.status);
        return r.json();
      })
      .then((data) => {
        term.write("\x1b[1A\x1b[K");

        if (data.history && data.history.length > 0) {
          const bootEl = document.querySelector("[data-term-boot]");
          const prevCount = bootEl
            ? JSON.parse(bootEl.getAttribute("data-term-boot") || "[]").length
            : 0;
          const newLines = data.history.slice(prevCount);
          for (const line of newLines) {
            term.writeln(line);
          }
          if (bootEl) {
            bootEl.setAttribute(
              "data-term-boot",
              JSON.stringify(data.history)
            );
          }
        }

        if (data.prompt) {
          prompt = data.prompt + " $ ";
        }

        term.writeln("");
        writePrompt();

        if (typeof htmx !== "undefined") {
          htmx.trigger(document.body, "session-updated");
        }
      })
      .catch((err) => {
        term.write("\x1b[1A\x1b[K");
        term.writeln(
          "\x1b[31m[error]\x1b[0m " + (err.message || "request failed")
        );
        term.writeln("");
        writePrompt();
      })
      .finally(() => {
        busy = false;
      });
  }

  function injectCommand(cmd) {
    if (busy) return;
    clearCurrentLine();
    currentLine = cmd;
    term.write(currentLine);
    term.writeln("");
    currentLine = "";
    history.push(cmd);
    histIdx = history.length;
    runCommand(cmd);
  }

  document.body.addEventListener("htmx:afterSwap", () => {
    requestAnimationFrame(() => {
      if (!term && shouldInit()) create();
      if (fitAddon) fitAddon.fit();
    });
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", () => {
      if (shouldInit()) create();
    });
  } else {
    if (shouldInit()) create();
  }
})();
