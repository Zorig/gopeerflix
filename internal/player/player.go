package player

import (
	"fmt"
	"os/exec"
	"runtime"
)

func PlayVLC(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "vlc", "--network-caching=1000", url)
	case "darwin":
		cmd = exec.Command("open", "-a", "VLC", "--args", "--network-caching=1000", url)
	case "linux":
		cmd = exec.Command("vlc", "--network-caching=1000", "--play-and-exit", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error launching VLC:", err)
		return err
	}

	fmt.Println("VLC launched successfully!")
	return nil
}
