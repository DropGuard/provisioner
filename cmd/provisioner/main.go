package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"provisioner"
	"provisioner/internal/config"
	"provisioner/internal/scoop"
)

func main() {
	fmt.Println("== Windows Provisioning Tool ==")

	fmt.Println("Loading embedded configuration...")
	cfg, err := config.LoadBytes(provisioner.EmbeddedConfig)

	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		waitAndExit(1)
	}

	fmt.Printf("Loaded configured apps: %d\n", len(cfg.Apps))
	fmt.Println("--------------------------------")

	// Phase 0: Run Setup Commands
	if len(cfg.SetupCommands) > 0 {
		fmt.Printf("Running %d setup commands...\n", len(cfg.SetupCommands))
		for _, cmdStr := range cfg.SetupCommands {
			fmt.Printf(" -> Executing: %s\n", cmdStr)
			cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", cmdStr)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf(" -> [WARN] Setup command failed: %v\n", err)
			}
		}
		fmt.Println("--------------------------------")
	}

	// Phase 1: Check and Install Scoop
	if !scoop.IsInstalled() {
		fmt.Println("Scoop is not installed. Attempting to install...")
		if err := scoop.InstallScoop(); err != nil {
			fmt.Printf("Failed to install Scoop: %v\n", err)
			waitAndExit(1)
		}
		fmt.Println("Scoop installed successfully!")
	} else {
		fmt.Println("Scoop is already installed.")
	}

	if err := scoop.ConfigureGitHubToken(); err != nil {
		fmt.Printf("Warning: Failed to configure GitHub token: %v\n", err)
	}

	fmt.Println("--------------------------------")

	// Phase 2: Add buckets
	buckets := make(map[string]struct{})
	for _, app := range cfg.Apps {
		if app.Bucket != "" && app.Bucket != "main" {
			buckets[app.Bucket] = struct{}{}
		}
	}

	if len(buckets) > 0 {
		fmt.Printf("Adding required buckets (%d)...\n", len(buckets))
		for b := range buckets {
			fmt.Printf(" -> Adding bucket: %s\n", b)
			if err := scoop.AddBucket(b); err != nil {
				fmt.Printf(" -> [WARN] Failed to add bucket %s: %v\n", b, err)
			}
		}
		fmt.Println("--------------------------------")
	}

	// Phase 3: Install Scoop apps
	fmt.Printf("Installing %d Scoop apps...\n", len(cfg.Apps))
	for _, app := range cfg.Apps {
		fmt.Printf(" -> Processing %s...\n", app.Name)
		if err := scoop.InstallApp(app.Name); err != nil {
			fmt.Printf("    [ERROR] Failed to install %s: %v\n", app.Name, err)
			continue
		}

		fmt.Printf("    [OK] Successfully installed %s.\n", app.Name)
		if !app.DesktopShortcut {
			continue
		}

		fmt.Printf("    Creating desktop shortcut for %s...\n", app.Name)
		if err := scoop.CreateDesktopShortcut(app.Name); err != nil {
			fmt.Printf("    [WARN] Failed to create desktop shortcut for %s: %v\n", app.Name, err)
		}
	}

	fmt.Println("--------------------------------")

	// Phase 4: Run Post-Setup Commands
	if len(cfg.PostSetupCommands) > 0 {
		fmt.Printf("Running %d post-setup commands...\n", len(cfg.PostSetupCommands))
		for _, cmdStr := range cfg.PostSetupCommands {
			fmt.Printf(" -> Executing: %s\n", cmdStr)
			cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", cmdStr)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf(" -> [WARN] Post-setup command failed: %v\n", err)
			}
		}
		fmt.Println("--------------------------------")
	}

	fmt.Println("Provisioning complete!")
	waitAndExit(0)
}

func waitAndExit(code int) {
	fmt.Println("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(code)
}
