const tabs = document.querySelectorAll(".mode-tab");
const panels = document.querySelectorAll(".mode-panel");
const outputTitle = document.querySelector("#output-title");
const message = document.querySelector("#message");
const output = document.querySelector("#output");
const outputActions = document.querySelector(".actions");
const copyButton = document.querySelector("#copy-button");
const downloadLink = document.querySelector("#download-link");
const fileList = document.querySelector("#file-list");
const downloadForm = document.querySelector("#download-form");
const convertForm = document.querySelector("#convert-form");
const convertFileInput = document.querySelector("#convert-file");
const pickFileButton = document.querySelector("#pick-file-button");
const clearFileButton = document.querySelector("#clear-file-button");
const convertFileCard = document.querySelector("#convert-file-card");
const convertMediaIcon = document.querySelector("#convert-media-icon");
const convertFileName = document.querySelector("#convert-file-name");
const convertFileMeta = document.querySelector("#convert-file-meta");
const convertFileHelp = document.querySelector("#convert-file-help");
const convertTargetHelp = document.querySelector("#convert-target-help");
const convertSubmit = document.querySelector("#convert-submit");
const convertFormatGroups = document.querySelectorAll("[data-source-kinds]");

let selectedMedia = null;

const mediaLabels = {
  video: "Video",
  audio: "Audio",
  image: "Image",
};

const formatLabels = {
  mp4: "MP4",
  webm: "WebM",
  mkv: "MKV",
  mov: "MOV",
  mp3: "MP3",
  m4a: "M4A",
  wav: "WAV",
  flac: "FLAC",
  ogg: "OGG",
  opus: "OPUS",
  jpg: "JPG",
  png: "PNG",
  gif: "GIF",
  tiff: "TIFF",
  webp: "WebP",
};

let capabilities = {
  webp: false,
};

function selected(form, name) {
  return new FormData(form).get(name);
}

function setActiveMode(mode) {
  tabs.forEach((tab) => tab.classList.toggle("active", tab.dataset.mode === mode));
  panels.forEach((panel) => panel.classList.toggle("active", panel.dataset.modePanel === mode));
  resetActions();
  const label = mode.charAt(0).toUpperCase() + mode.slice(1);
  const emptyOutput = mode === "convert"
    ? "Your converted file will appear here.\nSelect a local media file to begin."
    : "";
  showState(label, `Ready for ${mode}.`, emptyOutput);
}

function setBusy(form, isBusy) {
  form.querySelector("button[type='submit']").disabled = isBusy;
  form.querySelector("button[type='submit']").classList.toggle("is-busy", isBusy);
  form.querySelectorAll("input, button:not([type='submit'])").forEach((input) => {
    input.disabled = isBusy;
  });
}

function resetActions() {
  outputActions.hidden = true;
  copyButton.disabled = true;
  copyButton.hidden = true;
  downloadLink.classList.add("disabled");
  downloadLink.setAttribute("aria-disabled", "true");
  downloadLink.removeAttribute("href");
  downloadLink.hidden = true;
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

async function getJSON(path) {
  const response = await fetch(path);
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
  downloadLink.hidden = false;
  downloadLink.classList.remove("disabled");
  downloadLink.removeAttribute("aria-disabled");
}

function finishSuccess(title, data, fallbackText) {
  const files = data.files || (data.file ? [{ file: data.file, download: data.download }] : []);
  outputTitle.textContent = title;
  message.textContent = data.file || files.map((file) => file.file).join(" | ") || "Done";
  output.textContent = data.text || data.output || fallbackText;
  showFiles(files);
  enablePrimaryDownload(data);
  copyButton.hidden = !data.text;
  copyButton.disabled = !data.text;
  outputActions.hidden = copyButton.hidden && downloadLink.hidden;
}

function finishError(error) {
  showState("Needs Attention", "dlmac reported an error.", error.message);
}

function visibleConvertInputs() {
  return Array.from(convertForm.querySelectorAll("input[name='to']")).filter((input) => {
    const group = input.closest("[data-source-kinds]");
    const requirement = input.dataset.requires;
    return group && !group.hidden && (!requirement || capabilities[requirement]);
  });
}

function updateConvertFormats() {
  const kind = selectedMedia ? selectedMedia.kind : "";
  convertFormatGroups.forEach((group) => {
    const kinds = group.dataset.sourceKinds.split(" ");
    group.hidden = !kind || !kinds.includes(kind);
  });

  convertForm.querySelectorAll("input[name='to']").forEach((input) => {
    const requirement = input.dataset.requires;
    const available = !requirement || capabilities[requirement];
    input.closest("label").hidden = !available;
    input.disabled = !available;
  });

  const available = visibleConvertInputs();
  const current = available.find((input) => input.checked);
  if (!current && available[0]) {
    available[0].checked = true;
  }

  const target = selected(convertForm, "to");
  convertSubmit.disabled = !selectedMedia || !target;
  convertSubmit.querySelector(".button-label").textContent = selectedMedia && target
    ? `Convert to ${formatLabels[target] || target.toUpperCase()}`
    : "Choose a file";

  if (!selectedMedia) {
    convertTargetHelp.textContent = "Choose a file to see compatible targets.";
  } else if (kind === "video") {
    convertTargetHelp.textContent = "Choose a video format or extract the audio track.";
  } else if (kind === "image" && !capabilities.webp) {
    convertTargetHelp.textContent = "WebP output needs cwebp. Run ./install.sh or brew install webp.";
  } else {
    convertTargetHelp.textContent = `Choose an output format for this ${mediaLabels[kind].toLowerCase()}.`;
  }
}

async function loadCapabilities() {
  try {
    const data = await getJSON("/api/capabilities");
    capabilities = {
      webp: Boolean(data.webp),
    };
  } catch (error) {
    capabilities = { webp: false };
  }
  updateConvertFormats();
}

function mediaSummary(media) {
  const details = [mediaLabels[media.kind] || "Media", media.sizeLabel];
  if (media.dimensions) details.push(media.dimensions);
  if (media.duration) details.push(media.duration);
  return details.filter(Boolean).join(" · ");
}

function setSelectedMedia(media) {
  selectedMedia = media;
  convertFileInput.value = media.path;
  convertFileCard.hidden = false;
  convertMediaIcon.textContent = mediaLabels[media.kind] || "Media";
  convertFileName.textContent = media.name;
  convertFileMeta.textContent = mediaSummary(media);
  convertFileHelp.textContent = "Selected locally. The media bytes stay on this Mac.";
  updateConvertFormats();
}

function clearSelectedMedia(clearInput = true) {
  selectedMedia = null;
  if (clearInput) {
    convertFileInput.value = "";
  }
  convertFileCard.hidden = true;
  convertMediaIcon.textContent = "—";
  convertFileName.textContent = "No file selected";
  convertFileMeta.textContent = "Choose a local media file to inspect it.";
  convertFileHelp.textContent = "Use this when you already know the local path.";
  updateConvertFormats();
}

async function inspectConvertFile(path) {
  const data = await postJSON("/api/inspect", { file: path });
  setSelectedMedia(data.media);
  return data.media;
}

pickFileButton.addEventListener("click", async () => {
  pickFileButton.disabled = true;
  showState("Choose a file", "Opening the native macOS file picker...", "");

  try {
    const data = await postJSON("/api/pick-file", {});
    if (data.cancelled) {
      showState("Convert", "File selection canceled.", "Your converted file will appear here.\nSelect a local media file to begin.");
      return;
    }

    setSelectedMedia(data.media);
    showState("Ready to convert", `${data.media.name} is ready.`, `${data.media.name}\n${mediaSummary(data.media)}\n\nChoose a target format, then start conversion.`);
  } catch (error) {
    finishError(error);
  } finally {
    pickFileButton.disabled = false;
  }
});

clearFileButton.addEventListener("click", () => {
  clearSelectedMedia();
  convertFileInput.focus();
  showState("Convert", "Choose a local media file to begin.", "Your converted file will appear here.\nSelect a local media file to begin.");
});

convertFileInput.addEventListener("input", () => {
  if (selectedMedia && convertFileInput.value.trim() !== selectedMedia.path) {
    clearSelectedMedia(false);
  }
});

convertFileInput.addEventListener("blur", async () => {
  const path = convertFileInput.value.trim();
  if (!path || (selectedMedia && selectedMedia.path === path)) return;

  showState("Inspecting file", "Reading local media information...", "");
  try {
    const media = await inspectConvertFile(path);
    showState("Ready to convert", `${media.name} is ready.`, `${media.name}\n${mediaSummary(media)}\n\nChoose a target format, then start conversion.`);
  } catch (error) {
    finishError(error);
  }
});

convertForm.addEventListener("change", (event) => {
  if (event.target.matches("input[name='to']")) {
    updateConvertFormats();
  }
});

tabs.forEach((tab) => {
  tab.addEventListener("click", () => setActiveMode(tab.dataset.mode));
});

document.querySelector("#transcript-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  const payload = {
    url: form.elements.url.value.trim(),
    lang: selected(form, "lang"),
    format: selected(form, "format"),
  };
  resetActions();
  setBusy(form, true);
  showState("Working", "Fetching public captions through dlmac...");

  try {
    const data = await postJSON("/api/transcript", payload);
    finishSuccess("Transcript Saved", data, "Transcript saved. Use Download to open the file.");
  } catch (error) {
    finishError(error);
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
  const kind = selected(form, "kind");
  const payload = {
    url: form.elements.url.value.trim(),
    kind,
    quality: selected(form, "quality"),
    audioFormat: selected(form, "audioFormat"),
  };
  resetActions();
  setBusy(form, true);
  showState("Downloading", "dlmac is downloading through yt-dlp...");

  try {
    const data = await postJSON("/api/download", payload);
    finishSuccess(kind === "video" ? "Video Saved" : "Audio Saved", data, "Download finished.");
  } catch (error) {
    finishError(error);
  } finally {
    setBusy(form, false);
  }
});

convertForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;

  if (!convertFileInput.value.trim()) {
    finishError(new Error("Choose a local media file."));
    return;
  }

  const payload = {
    file: form.elements.file.value.trim(),
    to: selected(form, "to"),
  };
  resetActions();
  setBusy(form, true);
  showState(
    "Converting",
    payload.to === "webp" ? "cwebp is converting the local image..." : "ffmpeg is converting the local file...",
  );

  try {
    if (!selectedMedia || selectedMedia.path !== payload.file) {
      showState("Inspecting file", "Reading local media information...", "");
      await inspectConvertFile(payload.file);
      payload.to = selected(form, "to");
    }

    const data = await postJSON("/api/convert", payload);
    finishSuccess("File Converted", data, "Conversion finished.");
  } catch (error) {
    finishError(error);
  } finally {
    setBusy(form, false);
  }
});

updateConvertFormats();
loadCapabilities();

copyButton.addEventListener("click", async () => {
  if (!output.textContent.trim()) return;
  await navigator.clipboard.writeText(output.textContent);
  const original = copyButton.textContent;
  copyButton.textContent = "Copied";
  setTimeout(() => {
    copyButton.textContent = original;
  }, 1200);
});
