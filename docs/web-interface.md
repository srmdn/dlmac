# Local web interface

## Purpose

The local web interface gives `dlmac` a friendly local workbench for the core
single-item workflows:

```text
paste URL -> choose transcript/download options -> run dlmac
choose local media file -> choose target format -> convert
```

It is a localhost-only tool. It is not a hosted product, and it must not add
accounts, cloud storage, deployment, private-video workflows, or login/cookie
support.

## References

Use `plong` as an architecture reference, not a visual reference:

- `plong serve` starts a local server on `127.0.0.1`.
- It chooses an available port automatically.
- It prints the URL.
- It opens the browser.
- It keeps the dashboard embedded and dependency-light.
- It uses small JSON endpoints for actions.

Do not copy `plong`'s visual design. `plong` is a dark disk dashboard. `dlmac`
uses a calmer media workspace.

Use the `elayadesign/redesign-skill` approach when polishing the interface:

- Diagnose before changing.
- Work inside the existing stack.
- Avoid generic AI design patterns.
- Include hover, focus, loading, empty, success, and error states.
- Keep changes targeted and reviewable.

Reference: <https://github.com/elayadesign/redesign-skill>

## Product scope

### v0.3 MVP

The first web release supports the existing safe CLI workflows:

- Paste a public YouTube URL.
- Run transcript extraction with `English` or `Indonesian` captions.
- Save transcripts as `Text`, `WebVTT`, or `SRT`.
- Download video with optional quality selection.
- Download audio as `MP3`, `M4A`, or `WAV`.
- Choose a local media file with a native macOS file picker.
- Convert local video, audio, or image media to a supported target format.
- Show WebP as an image target when the local `cwebp` tools are available.
- Send only the local path to the localhost server; do not upload media bytes.
- Show progress while the command runs.
- Show inline errors from `yt-dlp` or `dlmac`.
- Display plain text transcripts in the page.
- Provide copy and download controls.
- Link to the saved file in `downloads/` when possible.

### Out of scope

- Hosted deployment.
- Accounts.
- Database.
- Transcript history.
- AI summary.
- Speech-to-text.
- Clipper.
- Batch jobs.
- Login, cookie, or private video support.
- Browser-session or credential import.

## Current commands

```text
dlmac serve
dlmac ui
```

`serve` is the primary command. `ui` can be an alias.

Behavior:

1. Start a local server on `127.0.0.1`.
2. Pick an available port.
3. Print the URL.
4. Open the browser on macOS.
5. Serve the local workbench interface.
6. Stop when the terminal process exits.

## Current architecture

The current CLI is Bash. The v0.3 MVP uses a small Go standard-library server
while keeping the Bash CLI as the engine.

Current v0.3 architecture:

```text
dlmac                 # existing Bash CLI and helper launcher
dlmac-web             # compiled local web server helper
cmd/dlmac-web/        # local web server source entry point
internal/web/         # handlers and embedded UI assets
internal/web/assets/  # HTML, CSS, and small client script
```

Keep the implementation simple:

- Use Go standard library for the local server.
- Do not add a JavaScript framework for v0.3.
- Do not add npm, Vite, React, or Electron for v0.3.
- Do not rewrite the existing Bash CLI before the web UI proves useful.

The local server calls the existing `dlmac transcript`, `video`, `audio`, and
`convert` commands. It uses `ffprobe` to inspect the selected local file and
`osascript` to open the native macOS file picker. Extract shared logic later
only when duplication becomes painful.

`install.sh` builds `dlmac-web` next to `dlmac`. The CLI runs that helper
without requiring Go or the source tree at runtime. A development checkout can
fall back to `go run ./cmd/dlmac-web` when the compiled helper is absent.

## Current endpoints

```text
GET  /                  Render interface
POST /api/transcript    Run transcript extraction
POST /api/download      Run video or audio download
POST /api/pick-file     Open the native local file picker
POST /api/inspect       Inspect a local media file
GET  /api/capabilities  Report optional local conversion capabilities
POST /api/convert       Run local file conversion
GET  /downloads/...     Serve saved output files from downloads/
```

`POST /api/transcript` input:

```json
{
  "url": "https://www.youtube.com/watch?v=...",
  "lang": "en",
  "format": "txt"
}
```

Response on success:

```json
{
  "ok": true,
  "file": "downloads/title.videoid.en.txt",
  "text": "Transcript text when format is txt"
}
```

Response on error:

```json
{
  "ok": false,
  "error": "No public transcript found for language 'id'."
}
```

`POST /api/pick-file` returns the selected path and media metadata. A canceled
selection returns `cancelled: true` and does not start a conversion.

`POST /api/inspect` input:

```json
{
  "file": "/absolute/path/to/media.webm"
}
```

`GET /api/capabilities` response:

```json
{
  "ok": true,
  "webp": true
}
```

The server reports WebP support when it finds both `cwebp` and `gif2webp`.
The interface hides the WebP target when either tool is unavailable.

`POST /api/convert` input:

```json
{
  "file": "/absolute/path/to/media.webm",
  "to": "mp4"
}
```

The conversion response uses the same saved-file shape as the download
response:

```json
{
  "ok": true,
  "file": "downloads/source.mp4",
  "download": "/downloads/source.mp4"
}
```

## Interface direction

The interface uses a compact media workbench rather than a dashboard.

Current layout:

- Single-column command surface on small screens.
- Two-panel layout on desktop: mode controls on the left, run output on the
  right.
- Compact header with product name, version, and local-only status.
- Mode tabs for transcript, download, and convert.
- URL or local media file input as the primary object.
- Native macOS file picker for path-only selection.
- Selected-file card with filename, media type, size, and dimensions or
  duration when available.
- Target format groups that adapt to video, audio, or image input.
- WebP image output when the local `webp` tools are installed.
- Segmented controls for language and format.
- One primary action per mode.
- Output area with copy, download, and saved file links.

Visual direction:

- Different from `plong`'s GitHub-like dark dashboard.
- Calm, editorial, and reading-focused.
- Avoid purple/blue AI gradients.
- Avoid decorative cards inside cards.
- Keep controls obvious and dense enough for repeated use.
- Use complete states: idle, loading, success, empty, and error.

## Verification

Before changing web behavior, run:

```bash
rtk bash -n dlmac install.sh
rtk shellcheck dlmac install.sh
rtk ./dlmac --help
rtk ./dlmac transcript --help
rtk go test ./...
rtk go build ./...
rtk ./install.sh
```

Manual browser checks:

- Page loads at localhost.
- URL validation shows inline errors.
- Layout does not overflow at desktop, tablet, or mobile widths.
- English transcript request succeeds for a video with public captions.
- Indonesian request shows a clear error when captions are unavailable.
- Text transcript can be copied.
- VTT and SRT transcripts can be downloaded.
- Video and audio download requests show saved files when available.
- The native picker selects a local file without uploading media bytes.
- Video, audio, and image conversion returns a saved output file.
- WebP output appears only when `cwebp` and `gif2webp` are available.
- The interface works on desktop and mobile widths.
