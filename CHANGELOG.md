# Changelog

## Unreleased

- Plan the localhost-only transcript web interface.
- Add the first localhost-only transcript web interface MVP.

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
