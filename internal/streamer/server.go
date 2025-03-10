package streamer

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
)

var Port = "18090"

func StartServer(torFile *torrent.File) {
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		fileSize := torFile.Length()
		rangeHeader := r.Header.Get("Range")
		start, end := parseRange(rangeHeader, fileSize)

		percent := float64(start) / float64(fileSize) * 100
		fmt.Printf("Streaming request: %.2f%% (%d-%d bytes)\n", percent, start, end)

		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Type", "video/*")
		w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.WriteHeader(http.StatusPartialContent)

		fileReader := torFile.NewReader()
		defer fileReader.Close()

		fileReader.Seek(start, io.SeekStart)
		buf := make([]byte, 32*1024) // 32KB buffer
		bytesLeft := end - start + 1

		for bytesLeft > 0 {
			n, err := fileReader.Read(buf)
			if err != nil && err != io.EOF {
				break
			}
			if int64(n) > bytesLeft {
				n = int(bytesLeft)
			}
			w.Write(buf[:n])
			bytesLeft -= int64(n)
			if err == io.EOF {
				break
			}
		}
	})

	fmt.Println("Streaming at http://localhost:" + Port + "/stream")
	if err := http.ListenAndServe(":"+Port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func parseRange(header string, fileSize int64) (int64, int64) {
	if header == "" || !strings.HasPrefix(header, "bytes=") {
		return 0, fileSize - 1
	}
	ranges := strings.Split(strings.TrimPrefix(header, "bytes="), "-")
	start, _ := strconv.ParseInt(ranges[0], 10, 64)
	end := fileSize - 1
	if len(ranges) > 1 && ranges[1] != "" {
		end, _ = strconv.ParseInt(ranges[1], 10, 64)
	}
	return start, end
}
