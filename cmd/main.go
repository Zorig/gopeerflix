package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"gopeerflix/internal/player"
	"gopeerflix/internal/streamer"
	"gopeerflix/internal/torrent"
	"gopeerflix/internal/ui"
)

func main() {
	var input string
	var useVLC bool
	var useIINA bool
	var noProgress bool

	rootCmd := &cobra.Command{
		Use:   "gopeerflix [magnet-uri | .torrent file]",
		Short: "Stream torrents directly to VLC or IINA",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			input = args[0]

			if useVLC && useIINA {
				fmt.Println("Error: --vlc and --iina cannot be used together.")
				os.Exit(1)
			}
			fmt.Println("Fetching metadata...")

			torrentClient := torrent.NewClient()
			tor, torFile, err := torrentClient.Stream(input)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			srv, err := streamer.StartServer(torFile)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if useVLC || useIINA {
				streamURL := srv.URL()
				var err error
				if useVLC {
					fmt.Println("Launching VLC...")
					err = player.PlayVLC(streamURL)
				} else {
					fmt.Println("Launching IINA...")
					err = player.PlayIINA(streamURL)
				}
				if err != nil {
					fmt.Println("Player launch failed:", err)
				} else {
					fmt.Println("Player launched successfully!")
				}
			}

			progress := ui.NewTorrent(tor, torFile, !noProgress)
			progress.Start()
			defer progress.Stop()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			progress.Stop()
			fmt.Println("Shutting down...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Shutdown(ctx)
			tor.Drop()
		},
	}

	rootCmd.Flags().BoolVarP(&useVLC, "vlc", "v", false, "Play in VLC")
	rootCmd.Flags().BoolVar(&useIINA, "iina", false, "Play in IINA (macOS only)")
	rootCmd.Flags().BoolVar(&noProgress, "no-progress", false, "Disable the real-time progress display")
	rootCmd.Execute()
}
