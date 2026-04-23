package scoop

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCreateDesktopShortcut_SubfolderGit verifies that CreateDesktopShortcut correctly
// resolves a shortcut whose Scoop manifest encodes a subfolder in shortcuts[1]:
//
//	git manifest: shortcuts[0][1] = "Git\Git Bash" → folder="Git", name="Git Bash"
//	expected .lnk path: <Scoop Apps>\Git\Git Bash.lnk
//
// Run with: go test ./internal/scoop/
func TestCreateDesktopShortcut_SubfolderGit(t *testing.T) {
	if _, err := exec.LookPath("scoop"); err != nil {
		t.Skip("scoop not found in PATH")
	}

	installViaScoopIfMissing(t, "git")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	// git's manifest: shortcuts[0][1] = "Git\Git Bash" → Base = "Git Bash"
	desktopLnk := filepath.Join(home, "Desktop", "Git Bash.lnk")

	// Setup: remove any leftover shortcut so a pre-existing file can't cause a false positive.
	_ = os.Remove(desktopLnk)

	// Teardown: remove the shortcut the test created to leave the desktop clean.
	t.Cleanup(func() {
		if err := os.Remove(desktopLnk); err != nil && !os.IsNotExist(err) {
			t.Logf("[WARN] cleanup: failed to remove %q: %v", desktopLnk, err)
		}
	})

	if err := CreateDesktopShortcut("git"); err != nil {
		t.Fatalf("CreateDesktopShortcut: %v", err)
	}

	if _, err := os.Stat(desktopLnk); os.IsNotExist(err) {
		t.Errorf("shortcut not created: expected file at %q", desktopLnk)
	}
}

// installViaScoopIfMissing installs appName via Scoop if it is not already in PATH.
// The app is intentionally never uninstalled — this helper assumes dev-time dependencies.
func installViaScoopIfMissing(t *testing.T, appName string) {
	t.Helper()

	if _, err := exec.LookPath(appName); err == nil {
		t.Logf("'%s' already in PATH, skipping install.", appName)
		return
	}

	t.Logf("'%s' not found in PATH, installing via Scoop...", appName)
	cmd := exec.Command("scoop", "install", appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("scoop install %s: %v", appName, err)
	}
}
