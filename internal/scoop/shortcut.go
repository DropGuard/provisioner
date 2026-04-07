package scoop

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is a partial representation of a Scoop manifest.
type Manifest struct {
	Shortcuts [][]string `json:"shortcuts"`
}

// CreateDesktopShortcut tries to find and copy shortcuts created by Scoop to the user's desktop.
func CreateDesktopShortcut(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	desktopDir := filepath.Join(home, "Desktop")

	// Get all possible locations for manifests and shortcuts
	manifestPaths := getPossibleManifestPaths(name, home)
	shortcutDirs := getPossibleShortcutDirs(home)

	var shortcuts []string
	for _, p := range manifestPaths {
		if data, err := os.ReadFile(p); err == nil {
			shortcuts = extractShortcutNames(string(data))
			if len(shortcuts) > 0 {
				break
			}
		}
	}

	// Case 1: No shortcuts found in manifest, fallback to fuzzy matching
	if len(shortcuts) == 0 {
		for _, dir := range shortcutDirs {
			_ = copyMatchingShortcuts(name, dir, desktopDir)
		}
		return nil
	}

	// Case 2: Found shortcuts in manifest, take the first one (intentional)
	scName := shortcuts[0]
	lnkName := scName + ".lnk"

	for _, dir := range shortcutDirs {
		if found, src := findShortcutRecursively(dir, lnkName); found {
			dst := filepath.Join(desktopDir, lnkName)
			fmt.Printf("      Copying shortcut: %s -> %s\n", scName, dst)
			if err := copyFile(src, dst); err != nil {
				fmt.Printf("      [WARN] Failed to copy shortcut %s: %v\n", scName, err)
			}
			return nil
		}
	}

	return nil
}

func getPossibleManifestPaths(name, home string) []string {
	var paths []string

	// Local scoop
	localRoot := os.Getenv("SCOOP")
	if localRoot == "" {
		localRoot = filepath.Join(home, "scoop")
	}
	paths = append(paths, filepath.Join(localRoot, "apps", name, "current", "manifest.json"))

	// Global scoop
	globalRoot := os.Getenv("SCOOP_GLOBAL")
	if globalRoot == "" {
		globalRoot = `C:\ProgramData\scoop`
	}
	paths = append(paths, filepath.Join(globalRoot, "apps", name, "current", "manifest.json"))

	return paths
}

func getPossibleShortcutDirs(home string) []string {
	var dirs []string

	// User start menu
	dirs = append(dirs, filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Scoop Apps"))

	// System-wide start menu (for global apps)
	dirs = append(dirs, `C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Scoop Apps`)

	return dirs
}

func extractShortcutNames(manifestContent string) []string {
	var m Manifest
	if err := json.Unmarshal([]byte(manifestContent), &m); err != nil {
		return nil
	}

	var names []string
	for _, s := range m.Shortcuts {
		if len(s) >= 2 {
			names = append(names, s[1])
		}
	}
	return names
}

func copyMatchingShortcuts(name string, srcDir, dstDir string) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	normalizedName := normalizeName(name)

	for _, f := range files {
		path := filepath.Join(srcDir, f.Name())

		if f.IsDir() {
			_ = copyMatchingShortcuts(name, path, dstDir)
			continue
		}

		if !strings.HasSuffix(strings.ToLower(f.Name()), ".lnk") {
			continue
		}

		pureName := strings.TrimSuffix(f.Name(), ".lnk")
		normalizedPure := normalizeName(pureName)

		if !strings.Contains(normalizedPure, normalizedName) && !strings.Contains(normalizedName, normalizedPure) {
			continue
		}

		dst := filepath.Join(dstDir, f.Name())
		fmt.Printf("      Found matching shortcut: %s. Copying to desktop...\n", f.Name())
		_ = copyFile(path, dst)
	}
	return nil
}

func findShortcutRecursively(dir, target string) (bool, string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, ""
	}

	for _, f := range files {
		path := filepath.Join(dir, f.Name())

		if f.IsDir() {
			if found, foundPath := findShortcutRecursively(path, target); found {
				return true, foundPath
			}
			continue
		}

		if strings.EqualFold(f.Name(), target) {
			return true, path
		}
	}
	return false, ""
}

func normalizeName(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, "-rev")
	s = strings.TrimSuffix(s, "-portable")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
