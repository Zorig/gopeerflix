package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
)

// Snapshot is a point-in-time view of download progress.
type Snapshot struct {
	Name          string
	FileBytes     int64
	FileLength    int64
	TorrentBytes  int64
	TorrentLength int64
	ActivePeers   int
	Seeders       int
}

// Provider supplies download snapshots for rendering.
type Provider interface {
	Snapshot() Snapshot
}

// UI renders real-time download progress to a terminal.
type UI struct {
	provider Provider
	w        io.Writer
	interval time.Duration
	enabled  bool
	now      func() time.Time

	rows      int
	lastBytes int64
	lastTime  time.Time
	speed     int64
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// New builds a progress UI that renders to stdout when enabled and stdout is a
// terminal.
func New(provider Provider, enabled bool) *UI {
	return &UI{
		provider: provider,
		w:        os.Stdout,
		interval: 500 * time.Millisecond,
		enabled:  enabled && isTerminal(os.Stdout),
		now:      time.Now,
	}
}

// NewTorrent builds a progress UI for streaming a torrent file.
func NewTorrent(tor *torrent.Torrent, file *torrent.File, enabled bool) *UI {
	return New(&torrentProvider{tor: tor, file: file}, enabled)
}

type torrentProvider struct {
	tor  *torrent.Torrent
	file *torrent.File
}

func (p *torrentProvider) Snapshot() Snapshot {
	s := Snapshot{
		FileBytes:     p.file.BytesCompleted(),
		FileLength:    p.file.Length(),
		TorrentBytes:  p.tor.BytesCompleted(),
		TorrentLength: p.tor.Length(),
	}
	if info := p.tor.Info(); info != nil {
		s.Name = info.Name
	}
	stats := p.tor.Stats()
	s.ActivePeers = stats.ActivePeers
	s.Seeders = stats.ConnectedSeeders
	return s
}

// Start begins rendering progress updates.
func (u *UI) Start() {
	if !u.enabled {
		return
	}
	fmt.Fprint(u.w, "\x1b[?25l") // hide cursor
	u.lastBytes = u.provider.Snapshot().TorrentBytes
	u.lastTime = u.now()
	u.stopCh = make(chan struct{})
	u.doneCh = make(chan struct{})
	go u.run()
}

func (u *UI) run() {
	defer close(u.doneCh)
	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()
	for {
		select {
		case <-u.stopCh:
			return
		case <-ticker.C:
			u.refresh(u.now())
		}
	}
}

// Stop halts rendering and clears the progress output.
func (u *UI) Stop() {
	if u.stopCh == nil {
		return
	}
	close(u.stopCh)
	<-u.doneCh
	u.clear()
	fmt.Fprint(u.w, "\x1b[?25h") // show cursor
	u.stopCh = nil
}

// clear removes the previously rendered block from the terminal.
func (u *UI) clear() {
	for i := 0; i < u.rows; i++ {
		fmt.Fprint(u.w, "\x1b[1A\x1b[2K")
	}
	u.rows = 0
}

func (u *UI) refresh(now time.Time) {
	s := u.provider.Snapshot()

	dt := now.Sub(u.lastTime)
	if dt > 0 {
		delta := s.TorrentBytes - u.lastBytes
		if delta >= 0 {
			u.speed = int64(float64(delta) / dt.Seconds())
		}
	}
	u.lastBytes = s.TorrentBytes
	u.lastTime = now

	lines := u.renderLines(s)
	u.clear()
	for _, line := range lines {
		fmt.Fprintln(u.w, line)
	}
	u.rows = len(lines)
}

func (u *UI) renderLines(s Snapshot) []string {
	var lines []string
	if s.Name != "" {
		lines = append(lines, s.Name)
	}
	lines = append(lines, progressBar("file", s.FileBytes, s.FileLength))
	lines = append(lines, fmt.Sprintf(
		"speed  %s   eta  %s   peers  %d (%d seeders)",
		humanizeRate(u.speed),
		formatETA(eta(s.FileBytes, s.FileLength, u.speed)),
		s.ActivePeers,
		s.Seeders,
	))
	lines = append(lines, fmt.Sprintf(
		"downloaded  %s of %s",
		humanize(s.TorrentBytes),
		humanize(s.TorrentLength),
	))
	return lines
}

func progressBar(label string, done, total int64) string {
	const width = 40
	var pct float64
	if total > 0 {
		pct = float64(done) / float64(total)
		if pct > 1 {
			pct = 1
		}
	}
	filled := int(pct * width)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
	return fmt.Sprintf("%-6s [%s] %5.1f%%", label, bar, pct*100)
}

func eta(done, total, speed int64) int64 {
	if speed <= 0 {
		return -1
	}
	remaining := total - done
	if remaining <= 0 {
		return 0
	}
	return remaining / speed
}

func formatETA(secs int64) string {
	if secs < 0 {
		return "--:--:--"
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func humanizeRate(b int64) string {
	return humanize(b) + "/s"
}

func humanize(n int64) string {
	f := float64(n)
	switch {
	case f >= 1<<30:
		return fmt.Sprintf("%.2f GB", f/(1<<30))
	case f >= 1<<20:
		return fmt.Sprintf("%.2f MB", f/(1<<20))
	case f >= 1<<10:
		return fmt.Sprintf("%.2f KB", f/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
