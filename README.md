# dlmac

macOS CLI wrapper for [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [ffmpeg](https://ffmpeg.org). Download online media and extract audio from local video files.

## Requirements

- macOS
- [Homebrew](https://brew.sh)
- yt-dlp
- ffmpeg

## Installation

```bash
git clone https://github.com/srmdn/dlmac.git
cd dlmac
./install.sh
```

`install.sh` checks dependencies and offers to install missing ones via Homebrew.

## Usage

```
dlmac info <url>                    Show video metadata
dlmac formats <url>                 Show available formats
dlmac video <url>                   Download best video (mp4)
dlmac video <url> --quality 360p    Download best video up to 360p
dlmac video <url> --quality 480p    Download best video up to 480p
dlmac video <url> --quality 720p    Download best video up to 720p
dlmac video <url> --quality 1080p   Download best video up to 1080p
dlmac audio <url>                   Download best audio (mp3)
dlmac audio <url> --format mp3      Download audio as mp3
dlmac audio <url> --format m4a      Download audio as m4a
dlmac audio <url> --format wav      Download audio as wav
dlmac convert <file> --to mp3       Extract audio from video to mp3
dlmac convert <file> --to m4a       Extract audio from video to m4a
dlmac convert <file> --to wav       Extract audio from video to wav
```

All downloads saved to `./downloads/`.

## Examples

```bash
# Get video info
./dlmac info "https://example.com/video"

# Download video at 720p
./dlmac video "https://example.com/video" --quality 720p

# Download audio as mp3
./dlmac audio "https://example.com/video" --format mp3

# Extract audio from local video
./dlmac convert myvideo.mp4 --to mp3
```

## Compatibility

**Video downloads produce H.264 video with AAC audio in an MP4 container.**
This is the format QuickTime, Safari, and Apple's media framework expect.

If the source offers H.264+AAC natively, dlmac downloads it directly with
zero re-encoding — fast and lossless. If H.264+AAC is unavailable, dlmac
automatically re-encodes the output with ffmpeg.

Players like VLC, IINA, and mpv can play any codec combination. The H.264+AAC
preference only matters for QuickTime and Apple apps.

## Troubleshooting

**QuickTime can't open the file**
The output is already H.264+AAC and should work with QuickTime. If you
still have issues, check that ffmpeg is installed (`brew install ffmpeg`)
and that your yt-dlp is up to date (`brew upgrade yt-dlp`).

**No H.264 formats available**
Some videos (particularly at 1080p and above) only offer VP9/AV1 video.
dlmac falls back to whatever is available and re-encodes automatically.
You'll see a "Re-encoding to H.264+AAC" message during download.

**"Re-encoding to H.264+AAC" message appears**
This is normal when the source doesn't provide H.264+AAC natively.
The download will take longer because ffmpeg must re-encode the entire
video. The result is still a QuickTime-compatible MP4.

## Limitations

- macOS only
- No playlist support
- No interactive format selector
- No login/cookie support
- Quality depends on source video availability; falls back to best available
  below the requested resolution

## Legal Notice

This tool is for local use with content you own, have permission to download, or that is legally available. It does **not** bypass DRM, paywalls, private content, or access restrictions.

Users are responsible for respecting copyright, platform terms of service, and applicable laws.

## License

GPL-3.0-or-later
