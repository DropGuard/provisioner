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

// ShortcutEntry holds the resolved shortcut name and its optional Start Menu subfolder.
//
// Scoop encodes the subfolder directly inside shortcuts[1] as a path, e.g. "JetBrains\\IDEA".
// We split that into Folder ("JetBrains") and Name ("IDEA") using filepath.Dir/Base.
type ShortcutEntry struct {
	Name   string // shortcut filename without extension → becomes <Name>.lnk
	Folder string // optional subdirectory under "Scoop Apps\" (may be empty / ".")
}

// CreateDesktopShortcut tries to find and copy shortcuts created by Scoop to the user's desktop.
func CreateDesktopShortcut(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	desktopDir := filepath.Join(home, "Desktop")
	shortcutDirs := getScoopShortcutDirs(home)

	// Try to resolve shortcut metadata from the installed manifest.
	shortcuts := readShortcutEntries(getManifestPaths(name, home))

	if len(shortcuts) > 0 {
		// Case 1: manifest found — build the exact path directly, no filesystem traversal.
		// Scoop may nest shortcuts under a subfolder (manifest shortcuts[2]).
		entry := shortcuts[0]
		lnkName := entry.Name + ".lnk"
		for _, dir := range shortcutDirs {
			src := buildShortcutPath(dir, entry.Folder, lnkName)
			if _, err := os.Stat(src); err == nil {
				dst := filepath.Join(desktopDir, lnkName)
				fmt.Printf("      Copying shortcut: %s -> %s\n", entry.Name, dst)
				if err := copyFile(src, dst); err != nil {
					fmt.Printf("      [WARN] Failed to copy shortcut %s: %v\n", entry.Name, err)
				}
				return nil
			}
		}
		return nil
	}

	// Case 2: no manifest — fall back to recursive fuzzy name matching.
	// Recursion is necessary here because the subfolder is unknown without a manifest.
	for _, dir := range shortcutDirs {
		_ = copyMatchingShortcuts(name, dir, desktopDir)
	}
	return nil
}

// buildShortcutPath constructs the full path to a shortcut file.
// If folder is non-empty, the shortcut is nested under that subdirectory.
func buildShortcutPath(baseDir, folder, lnkName string) string {
	if folder != "" {
		return filepath.Join(baseDir, folder, lnkName)
	}
	return filepath.Join(baseDir, lnkName)
}

// getManifestPaths returns candidate manifest.json paths for both user-level and global Scoop installs.
func getManifestPaths(name, home string) []string {
	localRoot := os.Getenv("SCOOP")
	if localRoot == "" {
		localRoot = filepath.Join(home, "scoop")
	}
	globalRoot := os.Getenv("SCOOP_GLOBAL")
	if globalRoot == "" {
		globalRoot = `C:\ProgramData\scoop`
	}
	return []string{
		filepath.Join(localRoot, "apps", name, "current", "manifest.json"),
		filepath.Join(globalRoot, "apps", name, "current", "manifest.json"),
	}
}

// getScoopShortcutDirs returns the Start Menu directories where Scoop places shortcuts,
// covering both user-level and global installs.
func getScoopShortcutDirs(home string) []string {
	return []string{
		filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Scoop Apps"),
		`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Scoop Apps`,
	}
}

// readShortcutEntries parses the first readable manifest from the given paths and returns
// its shortcut entries.
//
// Scoop encodes an optional subfolder directly in shortcuts[1] as a path separator,
// e.g. "JetBrains\\IDEA" → folder "JetBrains", name "IDEA".
func readShortcutEntries(manifestPaths []string) []ShortcutEntry {
	for _, p := range manifestPaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		var entries []ShortcutEntry
		for _, s := range m.Shortcuts {
			if len(s) < 2 {
				continue
			}
			// s[1] may be a plain name ("Google Chrome") or a path ("JetBrains\\IDEA").
			folder := filepath.Dir(s[1])  // "JetBrains" or "."
			name := filepath.Base(s[1])   // "IDEA" or "Google Chrome"
			if folder == "." {
				folder = ""
			}
			entries = append(entries, ShortcutEntry{Name: name, Folder: folder})
		}
		if len(entries) > 0 {
			return entries
		}
	}
	return nil
}

// copyMatchingShortcuts recursively scans srcDir for .lnk files whose names fuzzy-match
// the given app name, and copies any matches to dstDir.
//
// Recursion is needed here because without a manifest we don't know the subfolder.
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
