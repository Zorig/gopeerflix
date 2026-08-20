package streamer

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var fileData = []byte("0123456789")

type fakeStreamReader struct {
	*bytes.Reader
	responsive bool
	readahead  int64
}

func (f *fakeStreamReader) SetResponsive()       { f.responsive = true }
func (f *fakeStreamReader) SetReadahead(n int64) { f.readahead = n }
func (f *fakeStreamReader) Close() error         { return nil }

func newTestHandler() (http.Handler, *fakeStreamReader) {
	fr := &fakeStreamReader{Reader: bytes.NewReader(fileData)}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveStream(w, r, fr, "movie.mkv")
	})
	return h, fr
}

func doRequest(t *testing.T, h http.Handler, method, rangeHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/stream", nil)
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestFullRequest(t *testing.T) {
	h, fr := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != string(fileData) {
		t.Fatalf("body = %q, want %q", rec.Body.String(), fileData)
	}
	if got := rec.Header().Get("Content-Length"); got != "10" {
		t.Errorf("Content-Length = %q, want 10", got)
	}
	if got := rec.Header().Get("Content-Range"); got != "" {
		t.Errorf("unexpected Content-Range = %q", got)
	}
	if got := rec.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges = %q, want bytes", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "video/x-matroska" {
		t.Errorf("Content-Type = %q, want video/x-matroska", got)
	}
	if !fr.responsive {
		t.Error("reader was not configured responsive")
	}
	if fr.readahead != readaheadBytes {
		t.Errorf("readahead = %d, want %d", fr.readahead, readaheadBytes)
	}
}

func TestHeadRequest(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodHead, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD request returned a body: %q", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Length"); got != "10" {
		t.Errorf("Content-Length = %q, want 10", got)
	}
}

func TestRangeRequest(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "bytes=2-5")

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	if rec.Body.String() != "2345" {
		t.Fatalf("body = %q, want 2345", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes 2-5/10" {
		t.Errorf("Content-Range = %q, want bytes 2-5/10", got)
	}
	if got := rec.Header().Get("Content-Length"); got != "4" {
		t.Errorf("Content-Length = %q, want 4", got)
	}
}

func TestOpenEndedRange(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "bytes=3-")

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	if rec.Body.String() != "3456789" {
		t.Fatalf("body = %q, want 3456789", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes 3-9/10" {
		t.Errorf("Content-Range = %q, want bytes 3-9/10", got)
	}
}

func TestSuffixRange(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "bytes=-3")

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	if rec.Body.String() != "789" {
		t.Fatalf("body = %q, want 789", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes 7-9/10" {
		t.Errorf("Content-Range = %q, want bytes 7-9/10", got)
	}
}

func TestRangeBeyondEOF(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "bytes=99-")

	if rec.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", rec.Code)
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes */10" {
		t.Errorf("Content-Range = %q, want bytes */10", got)
	}
}

func TestReversedRange(t *testing.T) {
	h, _ := newTestHandler()
	rec := doRequest(t, h, http.MethodGet, "bytes=5-2")

	if rec.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", rec.Code)
	}
}

func TestContentTypeFor(t *testing.T) {
	cases := map[string]string{
		"movie.mp4":          "video/mp4",
		"series/episode.mkv": "video/x-matroska",
		"clip.avi":           "video/x-msvideo",
		"trailer.MOV":        "video/quicktime",
		"film.webm":          "video/webm",
		"broadcast.ts":       "video/mp2t",
		"audio.mp3":          "audio/mpeg",
		"sound.flac":         "audio/flac",
		"voice.ogg":          "audio/ogg",
		"readme.txt":         "application/octet-stream",
		"noextension":        "application/octet-stream",
	}
	for name, want := range cases {
		if got := contentTypeFor(name); got != want {
			t.Errorf("contentTypeFor(%q) = %q, want %q", name, got, want)
		}
	}
}

var _ io.ReadSeekCloser = (*fakeStreamReader)(nil)

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	srv := newServer(func() streamReader {
		return &fakeStreamReader{Reader: bytes.NewReader(fileData)}
	}, "movie.mkv")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv.listener = ln
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		if err := <-serveErr; err != http.ErrServerClosed {
			t.Errorf("Serve returned %v, want ErrServerClosed", err)
		}
	})

	base := "http://" + ln.Addr().String()
	if got := srv.URL(); got != base+"/stream" {
		t.Errorf("URL() = %q, want %q", got, base+"/stream")
	}
	return srv, base
}

func TestServerServesFullRequest(t *testing.T) {
	_, base := newTestServer(t)
	resp, err := http.Get(base + "/stream")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if string(body) != string(fileData) {
		t.Fatalf("body = %q, want %q", body, fileData)
	}
	if got := resp.Header.Get("Content-Type"); got != "video/x-matroska" {
		t.Errorf("Content-Type = %q, want video/x-matroska", got)
	}
	if got := resp.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Errorf("Accept-Ranges = %q, want bytes", got)
	}
}

func TestServerServesRangeRequest(t *testing.T) {
	_, base := newTestServer(t)
	req, err := http.NewRequest(http.MethodGet, base+"/stream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Range", "bytes=2-5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	if string(body) != "2345" {
		t.Fatalf("body = %q, want 2345", body)
	}
	if got := resp.Header.Get("Content-Range"); got != "bytes 2-5/10" {
		t.Errorf("Content-Range = %q, want bytes 2-5/10", got)
	}
}

func TestServerServesSuffixRange(t *testing.T) {
	_, base := newTestServer(t)
	req, err := http.NewRequest(http.MethodGet, base+"/stream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Range", "bytes=-3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	if string(body) != "789" {
		t.Fatalf("body = %q, want 789", body)
	}
}

func TestServerRejectsInvalidRange(t *testing.T) {
	_, base := newTestServer(t)
	req, err := http.NewRequest(http.MethodGet, base+"/stream", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Range", "bytes=99-")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Range"); got != "bytes */10" {
		t.Errorf("Content-Range = %q, want bytes */10", got)
	}
}

func TestServerUnknownRoute(t *testing.T) {
	_, base := newTestServer(t)
	resp, err := http.Get(base + "/other")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	srv, base := newTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	if _, err := client.Get(base + "/stream"); err == nil {
		t.Fatal("expected connection error after shutdown")
	}
}
