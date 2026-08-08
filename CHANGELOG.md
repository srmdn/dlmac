# Changelog

## Unreleased

- Expand local conversion to curated video, audio, and image target formats.
- Add WebP image output through the official Homebrew `webp` tools.
- Detect WebP encoder availability before showing the target in the web UI.
- Add a native macOS file picker and media inspection to the local workbench.
- Update the Convert UI with selected-file metadata and dynamic target groups.

## v0.3.0

- Add a localhost-only web workbench for transcript, download, and convert
  workflows.
- Support transcript language and format selection, video quality limits,
  audio formats, and local media conversion.
- Display text transcripts, command output, errors, and saved file links.
- Add responsive desktop, tablet, and mobile layouts.
- Redesign the workbench with a calmer local media studio interface.
- Fix web form payload capture, audio format handling, error presentation, and
  saved transcript file listings.
- Build and run a compiled `dlmac-web` helper without requiring the source tree
  or Go at runtime.

## v0.2.0

- Add `dlmac transcript` for public YouTube captions and auto captions.
- Support transcript languages with `--lang id|en`.
- Support transcript output formats with `--format txt|vtt|srt`.
- Prefer manual captions and fall back to auto captions when manual captions
  are unavailable.
- Convert WebVTT captions to plain text without adding dependencies.
- Update public project workflow docs, roadmap, and specification.

## v0.1.0

- Add the initial macOS CLI for `yt-dlp` and `ffmpeg`.
- Support media info, format listing, video download, audio download, and local
  video-to-audio conversion.
- Prefer H.264 video and AAC audio for QuickTime-compatible MP4 output.
- Add installation guidance, compatibility notes, troubleshooting, and legal
  boundaries.
