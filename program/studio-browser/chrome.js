const view = document.getElementById("view");
const url = document.getElementById("url");
const status = document.getElementById("status");
const recBtn = document.getElementById("rec");

let recorder = null;
let chunks = [];

const PART01 = "http://127.0.0.1:8765/part-01-the-missing-manual/slides.html";

function normalize(raw) {
  const t = raw.trim();
  if (!t) return PART01;
  if (/^[a-z]+:\/\//i.test(t)) return t;
  if (t.includes(" ") || !t.includes(".")) {
    return `https://www.google.com/search?q=${encodeURIComponent(t)}`;
  }
  return `https://${t}`;
}

document.getElementById("go").addEventListener("submit", (e) => {
  e.preventDefault();
  view.src = normalize(url.value);
});
document.getElementById("back").addEventListener("click", () => view.goBack());
document.getElementById("fwd").addEventListener("click", () => view.goForward());
document.getElementById("reload").addEventListener("click", () => view.reload());
document.getElementById("part").addEventListener("click", () => {
  url.value = PART01;
  view.src = PART01;
});
document.getElementById("theme").addEventListener("click", () => {
  const next = document.body.dataset.theme === "dark" ? "light" : "dark";
  document.body.dataset.theme = next;
  status.textContent = `chrome theme · ${next}`;
});

view.addEventListener("did-navigate", (e) => { url.value = e.url; });
view.addEventListener("did-navigate-in-page", (e) => { url.value = e.url; });
view.addEventListener("page-title-updated", (e) => {
  document.title = `${e.title} · AI Coder Studio`;
});

async function startRecord() {
  const stream = await navigator.mediaDevices.getDisplayMedia({
    video: { frameRate: 30 },
    audio: false,
  });
  chunks = [];
  recorder = new MediaRecorder(stream, { mimeType: "video/webm;codecs=vp9" });
  recorder.ondataavailable = (e) => { if (e.data.size) chunks.push(e.data); };
  recorder.onstop = async () => {
    stream.getTracks().forEach((t) => t.stop());
    const blob = new Blob(chunks, { type: "video/webm" });
    const bytes = new Uint8Array(await blob.arrayBuffer());
    const file = await window.studio.saveRecording(bytes);
    status.textContent = `saved · ${file}`;
    recBtn.dataset.on = "0";
    recBtn.textContent = "Record";
    recorder = null;
  };
  recorder.start(250);
  recBtn.dataset.on = "1";
  recBtn.textContent = "Stop";
  status.textContent = "recording this window only";
}

recBtn.addEventListener("click", async () => {
  try {
    if (recorder) recorder.stop();
    else await startRecord();
  } catch (err) {
    status.textContent = err.message || "record failed";
  }
});
