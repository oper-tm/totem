
# Totem Script

🇺🇸 English



**Totem** is an M3U8 downloader and streaming player powered by **Google Apps Script relays**. It is designed to bypass network restrictions, improve data transfer performance, and provide seamless access to M3U8 media and live streams.

M3U8 (HLS) is one of the most widely used media streaming formats, adopted by many modern video platforms and live streaming services.

With Totem, you can easily download M3U8 videos, watch online media playlists, and view live streams without the usual internet restrictions.

The latest version also supports watching **Twitch live streams** directly through Google Apps Script relays, allowing access without bandwidth throttling or regional restrictions.


## Features

* 🌍 Bypass censorship, sanctions, regional restrictions, and network filtering using Google Apps Script relays.
* 🚀 Boost M3U8 transfer speed by utilizing Google Apps Script relays.
* 🔽 Download M3U8 videos without common internet limitations.
* ▶️ Watch M3U8 videos online without downloading them first.
* 📺 Watch live M3U8 streams in real time.
* ⚡ Batch downloading for improved performance.
* 🟪 Watch Twitch live streams directly through Google Apps Script relays.
* 🎯 Automatically detect playlists and available video qualities.
* ⚙️ Simple JSON-based configuration.
* 🪶 Lightweight, fast, and written entirely in Go.
* 🖥️ Standalone executable with no installation required.
* 📜 Clean console output with detailed progress and error messages.
## Installation
To use m.a.r.d., download one of the releases.

[📥 Download the latest release](../../releases/latest)
## Configuration
- To use m.a.r.d., you need to perform a few steps in Google Apps Script.
These steps are simple and can be completed in just a few minutes.

1. Open https://script.google.com
2. Create a new Apps Script project.
3. Delete the default code and paste this entire file.
4. Click "Deploy" → "New deployment".
5. Select "Web app".
6. Configure:
   - Execute as: Me
   - Who has access: Anyone
7. Click "Deploy" and authorize the requested permissions.
8. Copy the generated Web App key (deployment ID).
9. Open config.json

- appScriptKey
    - Paste the AppScript deployment ID into the "appScriptKey" field.
    - To create an AppScript Key, follow the tutorial on the GitHub page or read the "app-script-code.txt" file.
    - Example:
`"appScriptKey": "AKfycbxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`

- batchSize
    - It is recommended not to change this value.
    - Number of TS segments included in each download batch.
    - During downloads, Totem groups TS segments into batches.
    - This value determines how many segments are included in each batch.
    - Only used in the "download" type.
    - Example:
`"batchSize": 5`

- maxSize
    - It is recommended not to change this value.
    - Some M3U8 playlists may contain a very large number of segments, which can reduce performance.
    - Only used in the "stream" type.
    - Example:
`"maxSize": 10`

- Download mode
    - It is recommended not to change this value.
    - Available values:
        - "go" - Downloads all TS segments in the current batch simultaneously.
        Recommended for faster and more stable internet connections.
        - "nr" - Downloads TS segments one by one within each batch.
        Recommended for slower or less stable internet connections.
    - Example:
`"downloadMode": "go"`

- Low latency
    - Available values:
        - false
        - true
    - Only used in the "stream" type.
    - true enables lower latency and is recommended for faster internet connections.
    - false provides higher latency and is recommended for slower internet connections.
    - Example:
`"lowLatency": false`

- Download path
    - It is recommended not to change this value.
    - This is NOT the output folder for the final MP4 file.
    - It is only used to store temporary ".ts" and ".m3u8" files.
    - Example:
`"downloadPath": "downloads/"`

- Watch port
    - It is recommended not to change this value.
    - Only used in the "stream" and "watch" types.
    - Specifies the local port used by Totem for streaming and watching.
    - Example:
`"watchPort": 1819`

- FFmpeg path
    - Enter the path to "ffmpeg.exe".
    - If you downloaded Totem from the GitHub Releases, you usually do not need to change this value.
    - If FFmpeg is already installed on your system and the "ffmpeg" command is available, you can leave this field empty (`""`).
    - If you are not using Windows, install FFmpeg manually and leave this field empty (`""`).
    - Example:
`"ffmpegPath": "ffmpeg.exe"`


## Usage/Examples

## Command Format

```bash
totem <type> <service> <m3u8_url> <output_file>
```

### type
Available values:
- `download` - Downloads the M3U8 stream and saves it as an MP4 file.
- `watch` - Plays an M3U8 video online.
- `stream` - Watches a live stream.

### service
Available values:
- `normal` - Standard mode. You must provide a direct M3U8 URL.
- `twitch` - Twitch live stream mode. Instead of an M3U8 URL, enter the Twitch channel name in the `<m3u8_url>` field.

### m3u8_url
- If `service` is `normal`, enter the direct M3U8 URL.
- If `service` is `twitch`, enter the Twitch channel name.

### output_file
- Only required when `type` is `download`.
- Specify the output filename and include the `.mp4` extension.
- For `watch` and `stream`, this argument is ignored, so any value can be used (for example, `null`).

### Examples

```bash
totem download normal https://example.com/test.m3u8 test.mp4
```

```bash
totem stream twitch ishowspeed1 null
```
## About the Bypass Mechanism

The component responsible for bypassing internet restrictions (such as censorship or regional access limitations) is the **Google Apps Script**.

Because the requests are made from Google's servers instead of your local machine, the Apps Script can access resources that may be unavailable or restricted from your network. It also acts as a proxy for downloading the M3U8 playlists and video segments before sending the data back to the application.


## RoadMap

- [x] M3U8 Downloader
- [X] Windows CLI version
- [x] M3U8 Stream
- [x] M3U8 Watch
- [x] Linux CLI version
- [x] macOS CLI version
- [ ] Graphical User Interface (GUI)
- [ ] Android application
## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
