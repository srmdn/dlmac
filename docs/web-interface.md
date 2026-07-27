# Local web interface

## Purpose

The local web interface gives `dlmac` a friendly transcript workflow:

```text
paste YouTube URL -> choose language -> choose format -> get transcript
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
needs a calmer transcript workspace.

Use the `elayadesign/redesign-skill` approach when polishing the interface:

- Diagnose before changing.
- Work inside the existing stack.
- Avoid generic AI design patterns.
- Include hover, focus, loading, empty, success, and error states.
- Keep changes targeted and reviewable.

Reference: <https://github.com/elayadesign/redesign-skill>

## Product scope

### v0.3 MVP

The first web release should support only transcripts:

- Paste a public YouTube URL.
- Choose `English` or `Indonesian`.
- Choose `Text`, `WebVTT`, or `SRT`.
- Run transcript extraction.
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

## Proposed command

```text
dlmac serve
dlmac ui
```

`serve` is the primary command. `ui` can be an alias.

Expected behavior:

1. Start a local server on `127.0.0.1`.
2. Pick an available port.
3. Print the URL.
4. Open the browser on macOS.
5. Serve the transcript interface.
6. Stop when the terminal process exits.

## Proposed architecture

The current CLI is Bash. A web server needs a stronger structure than Bash can
comfortably provide.

Recommended v0.3 architecture:

```text
dlmac                 # existing Bash CLI remains available
cmd/dlmac-web/        # local web server entry point
internal/web/         # handlers and templates
internal/transcript/  # shared transcript runner or wrapper
web/templates/        # HTML templates
web/static/           # CSS and small client script
```

Keep the first implementation simple:

- Prefer Go standard library for the local server if the maintainer approves
  adding Go as a development dependency.
- Do not add a JavaScript framework for v0.3.
- Do not add npm, Vite, React, or Electron for v0.3.
- Do not rewrite the existing Bash CLI before the web UI proves useful.

If Go is approved, the local server can call the existing `dlmac transcript`
command first, then extract shared logic later when duplication becomes painful.

## Proposed endpoints

```text
GET  /               Render interface
POST /api/transcript Run transcript extraction
GET  /downloads/...  Serve saved transcript files from downloads/
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

## Interface direction

Design a transcript workspace, not a dashboard.

Recommended layout:

- Single-column command surface on small screens.
- Two-panel layout on desktop: request form on the left, transcript output on
  the right.
- Compact header with product name, version, and local-only status.
- URL input as the primary object.
- Segmented controls for language and format.
- One primary action: get transcript.
- Output area with copy and download buttons.

Visual direction:

- Different from `plong`'s GitHub-like dark dashboard.
- Calm, editorial, and reading-focused.
- Avoid purple/blue AI gradients.
- Avoid decorative cards inside cards.
- Keep controls obvious and dense enough for repeated use.
- Use complete states: idle, loading, success, empty, and error.

## Verification

Before opening a PR for the web MVP:

```bash
rtk bash -n dlmac install.sh
rtk shellcheck dlmac install.sh
rtk ./dlmac --help
rtk ./dlmac transcript --help
```

If Go is added:

```bash
rtk go test ./...
rtk go build ./...
```

Manual browser checks:

- Page loads at localhost.
- URL validation shows inline errors.
- English transcript request succeeds for a video with public captions.
- Indonesian request shows a clear error when captions are unavailable.
- Text transcript can be copied.
- VTT and SRT transcripts can be downloaded.
- The interface works on desktop and mobile widths.
