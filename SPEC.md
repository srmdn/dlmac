# SPEC.md - dlmac

## Purpose

`dlmac` is a local macOS CLI wrapper for `yt-dlp` and `ffmpeg`. It downloads
permitted online media and extracts audio from local video files.

## Current release: v0.1

### Dependencies

- macOS
- Homebrew
- `yt-dlp`
- `ffmpeg`
- Bash

### Supported commands

```text
dlmac info <url>
dlmac formats <url>
dlmac video <url>
dlmac video <url> --quality 360p
dlmac video <url> --quality 480p
dlmac video <url> --quality 720p
dlmac video <url> --quality 1080p
dlmac audio <url>
dlmac audio <url> --format mp3
dlmac audio <url> --format m4a
dlmac audio <url> --format wav
dlmac convert <file> --to mp3
dlmac convert <file> --to m4a
dlmac convert <file> --to wav
```

### Expected behavior

`dlmac` with no arguments shows usage text.

`dlmac info <url>` shows basic metadata using `yt-dlp`.

`dlmac formats <url>` shows available formats using `yt-dlp -F`.

`dlmac video <url>` downloads the best video and audio, merges them to MP4,
and saves the result in `./downloads/`.

`dlmac video <url> --quality 360p|480p|720p|1080p` downloads the best
available video at or below the requested height, merges it with audio, and
saves the result in `./downloads/`.

`dlmac audio <url>` downloads the best audio, converts it to MP3, and saves it
in `./downloads/`.

`dlmac audio <url> --format mp3|m4a|wav` downloads the best audio and converts
it to the selected format.

`dlmac convert <file> --to mp3|m4a|wav` extracts audio from a local video file.
The output goes to `./downloads/`, preserves the base filename, and does not
overwrite an existing output file.

### Error cases

| Condition | Behavior |
| --- | --- |
| Missing `yt-dlp` | Show a clear install command |
| Missing `ffmpeg` | Show a clear install command |
| Missing URL argument | Show usage and exit non-zero |
| Invalid URL | Pass through the `yt-dlp` error |
| Missing local file | Show a clear file-not-found error |
| Unsupported format | List valid formats |
| Unsupported quality | List valid qualities |
| Missing `./downloads/` | Create it automatically |
| Existing conversion output file | Warn and do not overwrite |

## Planned release: v0.2 transcript

The next proposed feature is transcript extraction from public YouTube captions.
This section is a planning contract, not current behavior.

### Proposed commands

```text
dlmac transcript <url>
dlmac transcript <url> --lang id
dlmac transcript <url> --lang en
dlmac transcript <url> --format txt
dlmac transcript <url> --format vtt
dlmac transcript <url> --format srt
```

### Proposed behavior

- Download public captions or auto captions through `yt-dlp`.
- Prefer manually provided captions when available.
- Fall back to auto captions when manual captions are unavailable.
- Support Indonesian (`id`) and English (`en`) first.
- Save output to `./downloads/`.
- Do not download video for transcript-only commands.
- Return a clear error when no public transcript is available.

### Transcript non-goals

- No login or cookie support.
- No private, members-only, or restricted videos.
- No DRM or access restriction bypass.
- No speech-to-text dependency in v0.2.
- No AI summary in v0.2.
- No web UI in v0.2.

## General non-goals

- No GUI in the CLI package.
- No npm or Electron in the CLI package.
- No playlist support.
- No interactive format selector.
- No batch conversion.
- No cloud sync, daemon, or background service.
- No metadata editor or library management.

## Safety and legal boundary

This tool is for local use with content the user owns, has permission to
download, or that is legally available. It does not bypass DRM, paywalls,
private content, or access restrictions. Users are responsible for respecting
copyright, platform terms, and applicable laws.

## Verification checklist

```bash
rtk bash -n dlmac install.sh
rtk shellcheck dlmac install.sh
rtk ./dlmac --help
rtk ./dlmac --version
```

Manual behavior checks:

```bash
rtk ./dlmac info "URL"
rtk ./dlmac formats "URL"
rtk ./dlmac video "URL" --quality 720p
rtk ./dlmac video "URL" --quality 4k
rtk ./dlmac audio "URL" --format mp3
rtk ./dlmac audio "URL" --format flac
rtk ./dlmac convert sample.mp4 --to mp3
rtk ./dlmac convert missing.mp4 --to mp3
```
