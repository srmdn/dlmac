# Roadmap

## Direction

`dlmac` stays a small local media CLI first. The next valuable direction is a
transcript workflow for public YouTube captions, followed by an optional local
web interface that makes the same workflow easier to use.

Long term, `dlmac` can grow into a local-first YouTube learning and clipping
tool. The product should move in layers: transcript first, simple web UI next,
then clip extraction, then optional AI-assisted workflows.

## Current state

v0.2 is complete as a shippable CLI:

- Show media info.
- List formats.
- Download QuickTime-compatible MP4 video.
- Download MP3, M4A, or WAV audio.
- Download public YouTube transcripts in English or Indonesian.
- Extract audio from local video files.

## Current: transcript

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

## Next: local web tool

The next release should build a localhost-only web interface around the
transcript command. This should happen before clipper work.

MVP screen:

- URL input
- Language selector
- Format selector
- Run button
- Transcript viewer
- Copy and download buttons

Recommended boundary:

- Keep `dlmac` as the engine.
- Keep the web app thin and localhost-only.
- Do not deploy it.
- Do not visually copy `plong`.
- Use `plong` only as an architecture reference for local server behavior.
- Apply the redesign-skill approach when polishing the interface: diagnose
  before changing, avoid generic AI design patterns, and ship complete states.
- Do not add hosted accounts, subscriptions, or team features.
- Do not copy broad all-in-one AI product scope.

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

Consider these only after the transcript workflow works well:

- Local transcript history
- Markdown export
- Chapter extraction from timestamps
- Optional AI summary
- Optional OpenAI or Whisper transcription for files without captions
- Manual YouTube clipper
- AI-assisted clip suggestion
- Caption burn-in
- 9:16 vertical export
- Hosted web version
