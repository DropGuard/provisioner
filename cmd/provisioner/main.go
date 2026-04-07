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

	fmt.Println("--------------------------------")
	
	// Phase 2: Add buckets and install apps
	for _, app := range cfg.Apps {
		fmt.Printf("Processing %s...\n", app.Name)

		// Add bucket if specified
		if app.Bucket != "" && app.Bucket != "main" {
			fmt.Printf(" -> Adding bucket: %s\n", app.Bucket)
			if err := scoop.AddBucket(app.Bucket); err != nil {
				fmt.Printf(" -> [WARN] Failed to add bucket %s: %v\n", app.Bucket, err)
			}
		}

		// Install app
		status := "local"
		if app.Global {
			status = "GLOBAL"
		}
		fmt.Printf(" -> Installing %s (%s)...\n", app.Name, status)
		if err := scoop.InstallApp(app.Name, app.Global); err != nil {
			fmt.Printf(" -> [ERROR] Failed to install %s: %v\n", app.Name, err)
		} else {
			fmt.Printf(" -> Successfully installed %s.\n", app.Name)
			if cfg.CreateDesktopShortcuts {
				if err := scoop.CreateDesktopShortcuts(app.Name); err == nil {
					fmt.Printf(" -> [INFO] Checked/Created desktop shortcuts for %s.\n", app.Name)
				}
			}
		}
	}

	fmt.Println("--------------------------------")
	fmt.Println("Provisioning complete!")
	waitAndExit(0)
}

func waitAndExit(code int) {
	fmt.Println("\n按回车键 (Enter) 退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(code)
}
