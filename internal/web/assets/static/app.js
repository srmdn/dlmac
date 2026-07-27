const form = document.querySelector("#transcript-form");
const urlInput = document.querySelector("#url");
const submitButton = document.querySelector("#submit-button");
const outputTitle = document.querySelector("#output-title");
const message = document.querySelector("#message");
const transcript = document.querySelector("#transcript");
const copyButton = document.querySelector("#copy-button");
const downloadLink = document.querySelector("#download-link");

function selected(name) {
  return new FormData(form).get(name);
}

function setBusy(isBusy) {
  submitButton.disabled = isBusy;
  submitButton.classList.toggle("is-busy", isBusy);
  urlInput.disabled = isBusy;
  form.querySelectorAll("input[type='radio']").forEach((input) => {
    input.disabled = isBusy;
  });
}

function resetActions() {
  copyButton.disabled = true;
  downloadLink.classList.add("disabled");
  downloadLink.setAttribute("aria-disabled", "true");
  downloadLink.removeAttribute("href");
}

function showState(title, body, text = "") {
  outputTitle.textContent = title;
  message.textContent = body;
  transcript.textContent = text;
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  resetActions();
  setBusy(true);
  showState("Working", "Fetching public captions through dlmac...");

  try {
    const response = await fetch("/api/transcript", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        url: urlInput.value.trim(),
        lang: selected("lang"),
        format: selected("format"),
      }),
    });

    const data = await response.json();
    if (!response.ok || !data.ok) {
      throw new Error(data.error || "Transcript request failed.");
    }

    outputTitle.textContent = "Saved";
    message.textContent = data.file;
    transcript.textContent = data.text || "Transcript saved. Use Download to open the file.";

    downloadLink.href = data.download;
    downloadLink.classList.remove("disabled");
    downloadLink.removeAttribute("aria-disabled");
    copyButton.disabled = !data.text;
  } catch (error) {
    showState("Needs Attention", error.message);
  } finally {
    setBusy(false);
  }
});

copyButton.addEventListener("click", async () => {
  if (!transcript.textContent.trim()) return;
  await navigator.clipboard.writeText(transcript.textContent);
  const original = copyButton.textContent;
  copyButton.textContent = "Copied";
  setTimeout(() => {
    copyButton.textContent = original;
  }, 1200);
});
