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
		return fmt.Errorf("VLC is not supported on platform %q", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launching VLC: %w", err)
	}

	return nil
}

func PlayIINA(url string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("IINA is only supported on macOS")
	}

	cmd := exec.Command("open", "-a", "IINA", url)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launching IINA: %w", err)
	}

	return nil
}
