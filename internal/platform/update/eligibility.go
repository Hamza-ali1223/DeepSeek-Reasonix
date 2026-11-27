package update

import (
	"os"
	"path/filepath"
	"runtime"
)

// AutoUpdateEligible reports whether this machine's Studio install has a
// published fork updater path. Every other install can still read release notes
// and open the release page, but must not stage an in-app replacement.
func AutoUpdateEligible(layout Layout, line Line) bool {
	if layout.Executable == "" || layout.Root == "" || runtime.GOARCH != "amd64" {
		return false
	}
	switch runtime.GOOS {
	case "windows":
		marker, err := os.Stat(filepath.Join(filepath.Dir(layout.Executable), "Uninstall Reasonix Studio.exe"))
		return err == nil && marker.Mode().IsRegular()
	case "linux":
		return line.Deb.Package == "reasonix-studio" && line.OwnsInstalledPath(layout.Executable)
	default:
		return false
	}
}
