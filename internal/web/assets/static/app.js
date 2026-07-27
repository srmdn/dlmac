const tabs = document.querySelectorAll(".mode-tab");
const panels = document.querySelectorAll(".mode-panel");
const outputTitle = document.querySelector("#output-title");
const message = document.querySelector("#message");
const output = document.querySelector("#output");
const copyButton = document.querySelector("#copy-button");
const downloadLink = document.querySelector("#download-link");
const fileList = document.querySelector("#file-list");
const downloadForm = document.querySelector("#download-form");

function selected(form, name) {
  return new FormData(form).get(name);
}

function setActiveMode(mode) {
  tabs.forEach((tab) => tab.classList.toggle("active", tab.dataset.mode === mode));
  panels.forEach((panel) => panel.classList.toggle("active", panel.dataset.modePanel === mode));
  resetActions();
  const label = mode.charAt(0).toUpperCase() + mode.slice(1);
  showState(label, `Ready for ${mode}.`, "");
}

function setBusy(form, isBusy) {
  form.querySelector("button[type='submit']").disabled = isBusy;
  form.querySelector("button[type='submit']").classList.toggle("is-busy", isBusy);
  form.querySelectorAll("input").forEach((input) => {
    input.disabled = isBusy;
  });
}

function resetActions() {
  copyButton.disabled = true;
  downloadLink.classList.add("disabled");
  downloadLink.setAttribute("aria-disabled", "true");
  downloadLink.removeAttribute("href");
  fileList.hidden = true;
  fileList.replaceChildren();
}

function showState(title, body, text = "") {
  outputTitle.textContent = title;
  message.textContent = body;
  output.textContent = text;
}

async function postJSON(path, payload) {
  const response = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const data = await response.json();
  if (!response.ok || !data.ok) {
    throw new Error(data.error || "dlmac request failed.");
  }
  return data;
}

function showFiles(files = []) {
  fileList.replaceChildren();
  if (!files.length) {
    fileList.hidden = true;
    return;
  }

  files.forEach((file) => {
    const link = document.createElement("a");
    link.href = file.download;
    link.textContent = file.file;
    fileList.append(link);
  });
  fileList.hidden = false;
}

function enablePrimaryDownload(data) {
  const first = data.download ? data : (data.files || [])[0];
  if (!first || !first.download) return;
  downloadLink.href = first.download;
  downloadLink.classList.remove("disabled");
  downloadLink.removeAttribute("aria-disabled");
}

function finishSuccess(title, data, fallbackText) {
  outputTitle.textContent = title;
  message.textContent = data.file || (data.files || []).map((file) => file.file).join(" | ") || "Done";
  output.textContent = data.text || data.output || fallbackText;
  showFiles(data.files);
  enablePrimaryDownload(data);
  copyButton.disabled = !output.textContent.trim();
}

tabs.forEach((tab) => {
  tab.addEventListener("click", () => setActiveMode(tab.dataset.mode));
});

document.querySelector("#transcript-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  resetActions();
  setBusy(form, true);
  showState("Working", "Fetching public captions through dlmac...");

  try {
    const data = await postJSON("/api/transcript", {
      url: form.elements.url.value.trim(),
      lang: selected(form, "lang"),
      format: selected(form, "format"),
    });
    finishSuccess("Transcript Saved", data, "Transcript saved. Use Download to open the file.");
  } catch (error) {
    showState("Needs Attention", error.message);
  } finally {
    setBusy(form, false);
  }
});

downloadForm.addEventListener("change", () => {
  const kind = selected(downloadForm, "kind");
  document.querySelector("[data-download-options='video']").hidden = kind !== "video";
  document.querySelector("[data-download-options='audio']").hidden = kind !== "audio";
});

downloadForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  resetActions();
  setBusy(form, true);
  showState("Downloading", "dlmac is downloading through yt-dlp...");

  try {
    const kind = selected(form, "kind");
    const data = await postJSON("/api/download", {
      url: form.elements.url.value.trim(),
      kind,
      quality: selected(form, "quality"),
      audioFormat: selected(form, "audioFormat"),
    });
    finishSuccess(kind === "video" ? "Video Saved" : "Audio Saved", data, "Download finished.");
  } catch (error) {
    showState("Needs Attention", error.message);
  } finally {
    setBusy(form, false);
  }
});

document.querySelector("#convert-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  resetActions();
  setBusy(form, true);
  showState("Converting", "ffmpeg is converting the local file...");

  try {
    const data = await postJSON("/api/convert", {
      file: form.elements.file.value.trim(),
      to: selected(form, "to"),
    });
    finishSuccess("File Converted", data, "Conversion finished.");
  } catch (error) {
    showState("Needs Attention", error.message);
  } finally {
    setBusy(form, false);
  }
});

copyButton.addEventListener("click", async () => {
  if (!output.textContent.trim()) return;
  await navigator.clipboard.writeText(output.textContent);
  const original = copyButton.textContent;
  copyButton.textContent = "Copied";
  setTimeout(() => {
    copyButton.textContent = original;
  }, 1200);
});
