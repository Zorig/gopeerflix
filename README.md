# 🚀 GoPeerflix

> **The Lightweight, High-Performance Torrent Streaming CLI for VLC and IINA**

[![GitHub Release](https://img.shields.io/github/v/release/zorig/gopeerflix)](https://github.com/zorig/gopeerflix/releases)
[![Build Status](https://github.com/zorig/gopeerflix/actions/workflows/release.yml/badge.svg)](https://github.com/zorig/gopeerflix/actions)
![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8.svg?style=flat&logo=go)

---

## ✨ Features

- 🚀 **Blazing Fast:** Stream torrents instantly—no need to wait for full downloads!
- 🦋 **Lightweight & Efficient:** Minimal memory usage and optimized chunk streaming.
- 📺 **Media Player Ready:** Seamlessly integrates with VLC and IINA players.
- 🛠️ **Cross-Platform:** Supports Windows, Linux, and macOS.
- 🌎 **Magnet & Torrent Support:** Stream from magnet links or `.torrent` files.

## Planning to implement

- 🔄 **Auto Updates:** Automatically checks and updates itself.
- 📈 **Real-time TUI:** Interactive terminal UI showing real-time progress, speed, and ETA.

---

## 📦 Installation

### 🍺 Homebrew (macOS and Linux)

Install GoPeerflix from the official tap:

```sh
brew install zorig/tap/gopeerflix
```

After installation, run `gopeerflix` from anywhere in your terminal.

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
gopeerflix [magnet-link | torrent-file] --vlc
# or
gopeerflix [magnet-link | torrent-file] --iina
```

If you downloaded or built the binary locally, use `./gopeerflix` instead.

### 🎬 Example

Stream from a magnet link:

```sh
gopeerflix "magnet:?xt=urn:btih:yourmagnetlinkhere" --vlc
# or
gopeerflix "magnet:?xt=urn:btih:yourmagnetlinkhere" --iina
```

Or stream from a local torrent file:

```sh
gopeerflix ./myvideo.torrent --vlc
# or
gopeerflix ./myvideo.torrent --iina
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

### 🍿 IINA Integration (macOS)

Install IINA via Homebrew:

```sh
brew install --cask iina
```

Launch IINA automatically with the `--iina` flag when streaming on macOS.

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
