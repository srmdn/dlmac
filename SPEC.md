# SPEC.md - dlmac

## Purpose

`dlmac` is a local macOS CLI wrapper for `yt-dlp` and `ffmpeg`. It downloads
permitted online media and extracts audio from local video files.

## Current release: v0.3

### Dependencies

- macOS
- Homebrew
- `yt-dlp`
- `ffmpeg`
- Bash
- Go 1.26 or newer to build the local web helper during installation

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
dlmac transcript <url>
dlmac transcript <url> --lang id
dlmac transcript <url> --lang en
dlmac transcript <url> --format txt
dlmac transcript <url> --format vtt
dlmac transcript <url> --format srt
dlmac serve
dlmac ui
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

`dlmac transcript <url>` downloads public English captions or auto captions,
converts them to plain text, and saves the result in `./downloads/`.

`dlmac transcript <url> --lang id|en` downloads public captions or auto
captions for the selected language.

`dlmac transcript <url> --format txt|vtt|srt` saves the transcript in the
selected format. Plain text is the default.

`dlmac serve` starts the localhost-only web interface.

`dlmac ui` is an alias for `dlmac serve`.

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
| Missing public captions | Show a clear no-transcript error |
| Missing web helper in an installed layout | Ask the user to run the installer |
| Missing `./downloads/` | Create it automatically |
| Existing conversion output file | Warn and do not overwrite |

## Transcript non-goals

- No login or cookie support.
- No private, members-only, or restricted videos.
- No DRM or access restriction bypass.
- No speech-to-text dependency.
- No AI summary.
## Local web workbench

The current release includes a localhost-only web interface for the main
`dlmac` workflows.

### Commands

```text
dlmac serve
dlmac ui
```

`dlmac ui` is an alias for `dlmac serve`.

### Behavior

- Start a localhost server on `127.0.0.1`.
- Choose an available port automatically.
- Print the local URL in the terminal.
- Open the local URL in the default browser on macOS.
- Let the user choose transcript, download, or convert mode.
- Let the user paste a YouTube URL.
- Let the user choose `id` or `en`.
- Let the user choose `txt`, `vtt`, or `srt`.
- Run the existing transcript workflow.
- Let the user download video as MP4 with optional quality selection.
- Let the user download audio as `mp3`, `m4a`, or `wav`.
- Let the user convert a local video path to `mp3`, `m4a`, or `wav`.
- Show loading, success, empty, and error states.
- Display the transcript for `txt` output.
- Provide copy and download controls when output files are available.

Implementation:

- Uses Go standard library for the localhost server.
- Keeps the existing Bash CLI as the workflow engine.
- Installs a compiled `dlmac-web` helper next to the Bash CLI.
- Runs the compiled helper without requiring Go or the source tree at runtime.
- Falls back to `go run ./cmd/dlmac-web` in a development checkout when the
  compiled helper is unavailable.

### Web helper packaging

`install.sh` builds `dlmac-web` into a temporary file and moves it into place
only after the build succeeds. Re-running the installer safely replaces both
the helper and its executable permissions.

`dlmac serve` and `dlmac ui` look for an executable named `dlmac-web` in the
same directory as `dlmac`. If the helper is unavailable, a development
checkout can run the Go source directly. An installed CLI without either form
shows a clear error and asks the user to run the installer.

To move an installed copy to another directory, copy both `dlmac` and
`dlmac-web`.

### Web interface non-goals

- No hosted deployment.
- No accounts.
- No database.
- No cloud storage.
- No AI summary.
- No speech-to-text.
- No clipper.
- No login, cookie, or private video support.
- No browser credential, cookie, or session import.
- No design reuse from `plong`.

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
rtk go test ./...
rtk go build ./...
```

Manual behavior checks:

```bash
rtk ./dlmac info "URL"
rtk ./dlmac formats "URL"
rtk ./dlmac video "URL" --quality 720p
rtk ./dlmac video "URL" --quality 4k
rtk ./dlmac audio "URL" --format mp3
rtk ./dlmac audio "URL" --format flac
rtk ./dlmac transcript "URL" --lang en
rtk ./dlmac transcript "URL" --lang id
rtk ./dlmac transcript "URL" --format srt
rtk ./dlmac transcript "URL" --lang ar
rtk ./dlmac convert sample.mp4 --to mp3
rtk ./dlmac convert missing.mp4 --to mp3
```
