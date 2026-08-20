package streamer

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
)

// readaheadBytes is how far ahead of the current playback position pieces are
// prioritized for sequential download, so playback never waits on rarest-first
// piece selection.
const readaheadBytes = 64 * 1024 * 1024

// streamReader is the subset of torrent.Reader needed to serve a stream.
type streamReader interface {
	io.ReadSeekCloser
	SetResponsive()
	SetReadahead(int64)
}

// Server streams a torrent file over HTTP with byte-range support.
type Server struct {
	server    *http.Server
	listener  net.Listener
	newReader func() streamReader
	name      string
}

// NewServer builds a streaming server for the torrent file.
func NewServer(torFile *torrent.File) *Server {
	return newServer(func() streamReader { return torFile.NewReader() }, torFile.Path())
}

// newServer builds a streaming server from a reader factory and file name.
func newServer(newReader func() streamReader, name string) *Server {
	s := &Server{newReader: newReader, name: name}
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		serveStream(w, r, s.newReader(), s.name)
	})
	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

// StartServer listens on an ephemeral loopback port, serves the torrent file,
// and returns the running server so the caller can shut it down. Using a fresh
// port per run keeps the stream URL unique, so players don't resume a stale
// position from a previous torrent.
func StartServer(torFile *torrent.File) (*Server, error) {
	srv := NewServer(torFile)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("binding stream listener: %w", err)
	}
	srv.listener = ln
	fmt.Println("Streaming at " + srv.URL())
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
		}
	}()
	return srv, nil
}

// URL returns the full streaming URL for this server.
func (s *Server) URL() string {
	return "http://" + s.listener.Addr().String() + "/stream"
}

// ListenAndServe blocks serving requests on Addr until Shutdown, returning nil
// on a clean shutdown.
func (s *Server) ListenAndServe() error {
	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Serve blocks serving requests on ln until Shutdown.
func (s *Server) Serve(ln net.Listener) error {
	return s.server.Serve(ln)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// serveStream serves the file with proper HTTP range semantics: 200 for full
// responses, 206 for valid byte ranges, and 416 for unsatisfiable ranges. The
// reader is configured responsively with a readahead window so the client
// downloads the playback pieces sequentially instead of rarest-first.
func serveStream(w http.ResponseWriter, r *http.Request, sr streamReader, name string) {
	w.Header().Set("Content-Type", contentTypeFor(name))
	sr.SetResponsive()
	sr.SetReadahead(readaheadBytes)
	defer sr.Close()

	http.ServeContent(w, r, name, time.Time{}, sr)
}

// contentTypeFor returns the media type for a torrent-internal file name.
func contentTypeFor(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".ts":
		return "video/mp2t"
	case ".mp3":
		return "audio/mpeg"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg", ".oga":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	default:
		return "application/octet-stream"
	}
}
