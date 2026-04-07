package scoop

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// IsInstalled checks if scoop is available in the system PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("scoop")
	return err == nil
}

// InstallScoop attempts to install scoop via PowerShell.
func InstallScoop() error {
	// The standard way to install Scoop. If running as an administrator, -RunAsAdmin is required.
	psCmd := `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser; iex "& {$(irm get.scoop.sh)} -RunAsAdmin"`

	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", psCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Installing Scoop...")
	return cmd.Run()
}

// AddBucket adds a scoop bucket.
func AddBucket(name string) error {
	cmd := exec.Command("scoop", "bucket", "add", name)
	// Suppress output if it's already added, we will just try to add it.
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.ToLower(string(output))
		// If the bucket already exists, scoop typically throws an error with "already exists".
		if strings.Contains(outStr, "already exists") {
			return nil
		}
		return fmt.Errorf("failed to add bucket %s: %s", name, outStr)
	}
	return nil
}

// ConfigureGitHubToken sets the github_token for scoop. 
// It first tries to read from the GITHUB_TOKEN environment variable.
// If not found, it prompts the user for input.
func ConfigureGitHubToken() error {
	token := os.Getenv("GITHUB_TOKEN")

	if token == "" {
		fmt.Print("No GITHUB_TOKEN found in environment variables.\n")
		fmt.Print("Enter your GitHub token (or press Enter to skip): ")
		
		var input string
		fmt.Scanln(&input)
		token = strings.TrimSpace(input)
	}

	if token == "" {
		fmt.Println("Skipping GitHub token configuration.")
		return nil
	}

	fmt.Printf("Configuring GitHub token for Scoop...\n")
	cmd := exec.Command("scoop", "config", "github_token", token)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure GitHub token: %s", string(output))
	}
	fmt.Println("GitHub token configured successfully!")
	return nil
}

// InstallApp installs a software package via Scoop.
func InstallApp(name string) error {
	args := []string{"install", name}

	cmd := exec.Command("scoop", args...)

	// Create a buffer to capture the output for error detection,
	// while still streaming output to the terminal in real-time.
	var outBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &outBuf)

	err := cmd.Run()

	outStr := outBuf.String()

	// Check the text manually since scoop can swallow errors and exit with 0.
	if err != nil {
		return err
	}

	// If Scoop exited with 0 but actually failed, we extract its original error message and forward it directly.
	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ERROR:") || strings.HasPrefix(trimmed, "Couldn't find manifest") {
			return fmt.Errorf("%s", trimmed)
		}
	}

	return nil
}
