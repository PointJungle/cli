// Package browser opens a URL in the person's browser.
package browser

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Open starts the browser on url without waiting for it. $BROWSER wins when
// set; otherwise the operating system's opener is used (on WSL, the Windows
// browser through rundll32.exe when no Linux opener exists).
func Open(url string) error {
	for _, command := range candidates() {
		path, err := exec.LookPath(command[0])
		if err != nil {
			continue
		}
		cmd := exec.Command(path, append(command[1:], url)...)
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return nil
		}
	}
	return errors.New("no browser found")
}

func candidates() [][]string {
	if b := os.Getenv("BROWSER"); b != "" {
		return [][]string{strings.Fields(b)}
	}
	switch runtime.GOOS {
	case "darwin":
		return [][]string{{"open"}}
	case "windows":
		return [][]string{{"rundll32", "url.dll,FileProtocolHandler"}}
	default:
		return [][]string{{"xdg-open"}, {"wslview"}, {"rundll32.exe", "url.dll,FileProtocolHandler"}}
	}
}
