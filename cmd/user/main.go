package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func isAdmin() bool {
	// Simple check for admin rights on Windows
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

func getAdminGroupName() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "(New-Object System.Security.Principal.SecurityIdentifier('S-1-5-32-544')).Translate([System.Security.Principal.NTAccount]).Value")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get admin group name: %v", err)
	}

	name := strings.TrimSpace(string(out))
	// Name should be in format "DOMAIN\Group" or "BUILTIN\Group"
	parts := strings.Split(name, "\\")
	if len(parts) == 2 {
		return parts[1], nil
	}
	return name, nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandHidden(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(string(out)))
	}
	return err
}

func main() {
	if !isAdmin() {
		fmt.Println("[-] Please run this program as Administrator.")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

	// This is the daily-use account. It will be added to the Administrators group
	// so it has full privileges, but it is separate from the built-in Administrator
	// account which we use only for initial system setup.
	username := "DailyUser"

	fmt.Printf("[*] Creating user %s...\n", username)
	// Create the account with no password (intentional: single-user personal machine).
	err := runCommandHidden("net", "user", username, "/add")
	if err != nil {
		fmt.Printf("[!] Warning: user creation returned an error (user may already exist): %v\n", err)
	} else {
		fmt.Println("[+] User created successfully!")
	}

	adminGroup, err := getAdminGroupName()
	if err != nil {
		fmt.Printf("[-] FATAL: Failed to get the system administrator group name: %v\n", err)
		fmt.Println("[-] Execution aborted for safety. Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

	// Grant the daily account full administrator privileges.
	fmt.Printf("[*] Adding %s to the Administrators group (%s)...\n", username, adminGroup)
	err = runCommandHidden("net", "localgroup", adminGroup, username, "/add")
	if err != nil {
		fmt.Printf("[!] Warning: failed to add to Administrators group (may already be a member): %v\n", err)
	} else {
		fmt.Println("[+] Successfully added to the Administrators group!")
	}

	fmt.Println("[+] Setup complete!")
	fmt.Printf("[*] Daily administrator account '%s' has been created (no password).\n", username)
	fmt.Println("[*] Log off and switch to this account to use it as your daily driver.")
	fmt.Println("==================================================")
	fmt.Print("Do you want to log off now? (y/N): ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		fmt.Println("[*] Logging off...")
		runCommand("logoff")
	} else {
		fmt.Println("[*] Please log off manually later (Start -> Profile -> Sign out).")
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}
}
