package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type fakeProvider struct {
	s Snapshot
}

func (f *fakeProvider) Snapshot() Snapshot { return f.s }

func TestProgressBar(t *testing.T) {
	if got := progressBar("file", 0, 100); !strings.Contains(got, "0.0%") || !strings.Contains(got, "[----") {
		t.Errorf("progressBar(0/100) = %q", got)
	}
	half := progressBar("file", 50, 100)
	if !strings.Contains(half, "50.0%") {
		t.Errorf("progressBar(50/100) missing 50.0%%: %q", half)
	}
	if got := progressBar("file", 100, 100); !strings.Contains(got, "100.0%") || strings.Contains(got, "-") {
		t.Errorf("progressBar(100/100) = %q", got)
	}
	if got := progressBar("file", 1000, 100); !strings.Contains(got, "100.0%") {
		t.Errorf("progressBar clamped incorrectly: %q", got)
	}
}

func TestHumanize(t *testing.T) {
	cases := map[int64]string{
		0:                      "0 B",
		512:                    "512 B",
		1024:                   "1.00 KB",
		5 * 1024 * 1024:        "5.00 MB",
		2 * 1024 * 1024 * 1024: "2.00 GB",
	}
	for n, want := range cases {
		if got := humanize(n); got != want {
			t.Errorf("humanize(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestFormatETA(t *testing.T) {
	cases := map[int64]string{
		-1:    "--:--:--",
		0:     "00:00:00",
		59:    "00:00:59",
		3661:  "01:01:01",
		86400: "24:00:00",
	}
	for secs, want := range cases {
		if got := formatETA(secs); got != want {
			t.Errorf("formatETA(%d) = %q, want %q", secs, got, want)
		}
	}
}

func TestEta(t *testing.T) {
	cases := []struct {
		done, total, speed, want int64
	}{
		{50, 100, 10, 5},
		{100, 100, 10, 0},
		{50, 100, 0, -1},
		{50, 100, -5, -1},
	}
	for _, c := range cases {
		if got := eta(c.done, c.total, c.speed); got != c.want {
			t.Errorf("eta(%d,%d,%d) = %d, want %d", c.done, c.total, c.speed, got, c.want)
		}
	}
}

func TestRefreshComputesSpeed(t *testing.T) {
	f := &fakeProvider{s: Snapshot{TorrentBytes: 100}}
	u := &UI{provider: f, w: &bytes.Buffer{}, interval: time.Second}

	t0 := time.Unix(0, 0)
	u.lastTime = t0
	u.lastBytes = 0
	u.refresh(t0.Add(2 * time.Second))

	if u.speed != 50 {
		t.Errorf("speed = %d, want 50 bytes/s", u.speed)
	}
}

func TestRenderLines(t *testing.T) {
	f := &fakeProvider{s: Snapshot{
		Name:          "Big Buck Bunny",
		FileBytes:     50,
		FileLength:    100,
		TorrentBytes:  60,
		TorrentLength: 100,
		ActivePeers:   4,
		Seeders:       1,
	}}
	u := &UI{provider: f, w: &bytes.Buffer{}, interval: time.Second, speed: 1024}

	lines := u.renderLines(f.s)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Big Buck Bunny") {
		t.Errorf("render missing torrent name: %q", joined)
	}
	if !strings.Contains(joined, "50.0%") {
		t.Errorf("render missing file progress: %q", joined)
	}
	if !strings.Contains(joined, "1.00 KB/s") {
		t.Errorf("render missing speed: %q", joined)
	}
	if !strings.Contains(joined, "peers  4 (1 seeders)") {
		t.Errorf("render missing peer stats: %q", joined)
	}
}

func TestStartStopNoOutputWhenDisabled(t *testing.T) {
	f := &fakeProvider{s: Snapshot{TorrentBytes: 100}}
	u := &UI{provider: f, w: &bytes.Buffer{}, interval: time.Hour, enabled: false}

	u.Start()
	u.Stop()

	if u.w.(*bytes.Buffer).Len() != 0 {
		t.Errorf("disabled UI produced output: %q", u.w.(*bytes.Buffer).String())
	}
}
