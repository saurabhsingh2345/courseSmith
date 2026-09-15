const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("studio", {
  saveRecording: (bytes) => ipcRenderer.invoke("record:save", bytes),
});
