package torrent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"
)

type TorrentClient struct {
	Session *torrent.Client
}

func NewClient() *TorrentClient {
	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = false
	cfg.NoDHT = true
	cfg.DisableIPv6 = true
	cfg.DataDir = "/tmp/gopeerflix"

	client, err := torrent.NewClient(cfg)
	if err != nil {
		fmt.Println("Failed to create torrent client", err)
		os.Exit(1)
	}

	return &TorrentClient{Session: client}
}

func (t *TorrentClient) Stream(input string) (*torrent.Torrent, *torrent.File, error) {
	var tor *torrent.Torrent
	var err error

	if isTorrentFile(input) {
		tor, err = t.Session.AddTorrentFromFile(input)
	} else {
		tor, err = t.Session.AddMagnet(input)
	}

	if err != nil {
		return nil, nil, err
	}

	<-tor.GotInfo()
	var selectedFile *torrent.File
	var maxSize int64

	// Select the largest video file.
	for _, file := range tor.Files() {
		if file.Length() > maxSize && isVideo(file.Path()) {
			selectedFile = file
			maxSize = file.Length()
		}
	}

	if selectedFile == nil {
		return nil, nil, fmt.Errorf("No video file found in torrent")
	}

	selectedFile.Download()
	return tor, selectedFile, nil
}

func isVideo(filePath string) bool {
	ext := []string{".mp4", ".mkv", ".avi", ".mov"}
	for _, e := range ext {
		if len(filePath) > len(e) && filePath[len(filePath)-len(e):] == e {
			return true
		}
	}
	return false
}

func isTorrentFile(input string) bool {
	return filepath.Ext(input) == ".torrent"
}
