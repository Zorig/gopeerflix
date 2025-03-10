package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"gopeerflix/internal/player"
	"gopeerflix/internal/streamer"
	"gopeerflix/internal/torrent"
)

func main() {
	var input string
	var useVLC bool

	rootCmd := &cobra.Command{
		Use:   "gopeerflix [magnet-uri | .torrent file]",
		Short: "Stream torrents directly to VLC",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input = args[0]
			fmt.Println("Fetching metadata...")

			torrentClient := torrent.NewClient()
			tor, torFile, err := torrentClient.Stream(input)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			go streamer.StartServer(torFile)

			if useVLC {
				streamURL := "http://localhost:" + streamer.Port + "/stream"
				player.PlayVLC(streamURL)
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			fmt.Println("Shutting down...")
			tor.Drop()
		},
	}

	rootCmd.Flags().BoolVarP(&useVLC, "vlc", "v", false, "Play in VLC")
	rootCmd.Execute()
}
