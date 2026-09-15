const { app, BrowserWindow, ipcMain, session, desktopCapturer } = require("electron");
const path = require("path");
const fs = require("fs");

const CAPTURES = path.join(__dirname, "..", "part-01-the-missing-manual", "recordings", "studio");

let win;

function createWindow() {
  win = new BrowserWindow({
    width: 1600,
    height: 900,
    minWidth: 1100,
    minHeight: 700,
    backgroundColor: "#070707",
    title: "AI Coder Studio",
    titleBarStyle: "hiddenInset",
    trafficLightPosition: { x: 16, y: 16 },
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
      webviewTag: true,
      sandbox: false,
    },
  });

  win.loadFile(path.join(__dirname, "chrome.html"));
}

app.whenReady().then(() => {
  session.defaultSession.setDisplayMediaRequestHandler(async (_req, callback) => {
    const sources = await desktopCapturer.getSources({
      types: ["window"],
      thumbnailSize: { width: 0, height: 0 },
    });
    const self =
      sources.find((s) => win && s.id === win.getMediaSourceId()) ||
      sources.find((s) => /AI Coder Studio/i.test(s.name)) ||
      sources[0];
    callback({ video: self });
  }, { useSystemPicker: false });

  createWindow();
});

app.on("window-all-closed", () => app.quit());

ipcMain.handle("record:save", async (_e, bytes) => {
  fs.mkdirSync(CAPTURES, { recursive: true });
  const stamp = new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);
  const file = path.join(CAPTURES, `studio-${stamp}.webm`);
  fs.writeFileSync(file, Buffer.from(bytes));
  return file;
});
