# Roadmap

## Direction

`dlmac` stays a small local media CLI first. The CLI remains the engine, and the
localhost web workbench provides a friendlier interface for the same safe,
single-item workflows.

Long term, `dlmac` can grow into a local-first YouTube learning and clipping
tool. The product moves in layers: release the current web workbench, stabilize
local workflows, add manual clip extraction, and consider optional AI-assisted
features only after the core workflows are reliable.

## Current state

v0.3 is the latest release. It provides a shippable CLI and local web
workbench that can:

- Show media info.
- List formats.
- Download QuickTime-compatible MP4 video.
- Download MP3, M4A, or WAV audio.
- Download public YouTube transcripts in English or Indonesian.
- Download transcripts in Text, WebVTT, or SRT format.
- Download video with a quality limit.
- Download audio as MP3, M4A, or WAV.
- Convert local video, audio, and image files to curated target formats.
- Convert supported local images to WebP through the official `cwebp` tools.
- Hide WebP output when the local WebP tools are unavailable.
- Select local files through the native macOS picker without uploading media.
- Display text transcripts and link to saved output files.
- Run entirely on `127.0.0.1` without accounts or hosted services.

## Released: v0.2 transcript

The `transcript` command turns a public YouTube URL into a downloadable caption
file.

Target workflow:

```text
paste YouTube URL -> choose language -> download transcript -> copy or save
```

Initial scope:

- `dlmac transcript <url>`
- `--lang id|en`
- `--format txt|vtt|srt`
- Public captions through `yt-dlp`
- Manual captions preferred over auto captions
- Clear failure when captions are unavailable

Out of scope for v0.2:

- Speech-to-text generation
- AI summary
- Web UI
- Accounts
- Database
- Background jobs

## Released: v0.3 local web workbench

The local web workbench wraps the existing safe CLI workflows in a
localhost-only browser interface.

Implemented:

- Mode tabs for transcript, download, and convert
- URL input for transcript/download
- Local file path input for convert
- Native local file picker for convert
- Language, quality, and format selectors
- Run buttons
- Output viewer
- Copy and download buttons

Release boundaries:

- Keep the workbench thin, localhost-only, and dependency-light.
- Keep accounts, hosted services, private-video workflows, AI summary, and
  speech-to-text outside the release scope.

## Candidate after v0.3: local activity history

Consider local activity history only after v0.3 is released. If approved, it
can record recent transcript, download, and convert runs without adding
accounts or a database.

This remains a candidate, not committed release scope.

## Later: YouTube clipper

After the transcript and local web tool are stable, add a clipper workflow for
turning long videos into selected short clips.

Target workflow:

```text
paste YouTube URL -> fetch transcript -> select segment -> cut clip -> export
```

Initial scope:

- Use transcript timestamps when available.
- Let the user select start and end times manually.
- Cut clips with `ffmpeg`.
- Export MP4 clips to `downloads/clips/`.
- Preserve QuickTime-compatible H.264 and AAC output.
- Keep this local-only and single-user.

Out of scope for the first clipper release:

- Auto-posting to social platforms
- User accounts
- Cloud storage
- Scheduled uploads
- AI actors
- Voice cloning
- Image or video generation
- Team workflows

## Later: AI-assisted clipping

Once manual clipping works, add optional AI assistance.

Useful features:

- Detect high-potential moments from transcripts.
- Suggest short titles and hook text.
- Generate chapter lists.
- Burn captions into clips.
- Reframe horizontal video to 9:16.
- Score clips by clarity, length, and topic focus.

Recommended boundary:

- Make AI features optional.
- Keep provider keys local.
- Start with transcript-only analysis before adding visual analysis.
- Add each provider behind a small interface so it can be swapped later.

## Long-term architecture

If the project outgrows Bash, split it into layers:

```text
dlmac CLI
  -> media engine
  -> transcript engine
  -> clip engine
  -> optional local web UI
```

Potential package layout:

```text
bin/
  dlmac
src/
  media/
  transcript/
  clips/
  web/
docs/
  roadmap.md
  workflows.md
```

Do not rewrite into a larger stack until one of these is true:

- Transcript parsing becomes hard to maintain in Bash.
- Clipping needs structured job state.
- The web UI needs a backend API.
- Tests need reusable functions instead of shell-only integration checks.

## Future ideas

Consider these only after the v0.3 workbench is released and stable:

- Local activity history
- Markdown export
- Chapter extraction from timestamps
- Optional AI summary
- Optional OpenAI or Whisper transcription for files without captions
- Manual YouTube clipper
- AI-assisted clip suggestion
- Caption burn-in
- 9:16 vertical export
- Hosted web version
