# 🚀 GoPeerflix

> **The Lightweight, High-Performance Torrent Streaming CLI for VLC**

[![GitHub Release](https://img.shields.io/github/v/release/zorig/gopeerflix)](https://github.com/zorig/gopeerflix/releases)
[![Build Status](https://github.com/zorig/gopeerflix/actions/workflows/release.yml/badge.svg)](https://github.com/zorig/gopeerflix/actions)
![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8.svg?style=flat&logo=go)

---

## ✨ Features

- 🚀 **Blazing Fast:** Stream torrents instantly—no need to wait for full downloads!
- 🦋 **Lightweight & Efficient:** Minimal memory usage and optimized chunk streaming.
- 📺 **VLC Ready:** Seamlessly integrates with VLC media player.
- 🔄 **Auto Updates:** Automatically checks and updates itself.
- 🛠️ **Cross-Platform:** Supports Windows, Linux, and macOS.
- 🌎 **Magnet & Torrent Support:** Stream from magnet links or `.torrent` files.
- 📈 **Real-time TUI:** Interactive terminal UI showing real-time progress, speed, and ETA.

---

## 📦 Installation

### 📥 Download Pre-Built Binaries

Visit the [Releases](https://github.com/zorig/gopeerflix/releases) page to download binaries for Linux, Windows, or macOS.

### 🛠️ Build from Source

Ensure you have [Go installed](https://golang.org/dl/):

```bash
git clone https://github.com/zorig/gopeerflix.git
cd gopeerflix
go build -o gopeerflix ./cmd
```

## 🚩 Usage

```sh
./gopeerflix [magnet-link | torrent-file] --vlc
```

### 🎬 Example

Stream from a magnet link:

```sh
./gopeerflix "magnet:?xt=urn:btih:yourmagnetlinkhere" --vlc
```

Or stream from a local torrent file:

```sh
./gopeerflix ./myvideo.torrent --vlc
```

---

### 📺 VLC Integration

Ensure VLC is installed:

- Linux:

```sh
sudo apt install vlc
```

- macOS:

```sh
brew install vlc
```

- Windows
  [Download VLC](https://www.videolan.org/vlc/download-windows.html)

GoPeerflix automatically opens VLC for instant streaming.

### ⚡ Performance Optimizations

- Efficient Buffering: Low memory footprint.
- Direct Chunk Streaming: Streams immediately without full downloads.
- Connection Management: Limited torrent connections for optimized performance.
- HTTP Range Requests: Smooth seeking and playback.

### 🔨 Contributing

Contributions are welcome! Please fork, submit pull requests, or open issues if you encounter bugs or have suggestions.

🌟 Star the Project!
If you find this useful, please ⭐️ star the repository—it helps keep me motivated!

### 📝 License

MIT License
