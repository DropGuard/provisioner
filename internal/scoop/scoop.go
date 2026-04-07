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

// InstallApp installs a software package via Scoop.
func InstallApp(name string, global bool) error {
	args := []string{"install"}
	if global {
		args = append(args, "-g")
	}
	args = append(args, name)

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
			return fmt.Errorf("%s", trimmed) // 直接转发 Scoop 的原话
		}
	}

	return nil
}

// CreateDesktopShortcuts copies any start menu shortcuts created by Scoop for the app to the Desktop.
func CreateDesktopShortcuts(appName string) error {
	psCmd := `
$desktop = [Environment]::GetFolderPath("Desktop")
$startMenu = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Scoop Apps"
$globalStartMenu = "$env:ALLUSERSPROFILE\Microsoft\Windows\Start Menu\Programs\Scoop Apps"

$shell = New-Object -ComObject WScript.Shell
Get-ChildItem -Path $startMenu, $globalStartMenu -Filter "*.lnk" -ErrorAction SilentlyContinue | ForEach-Object {
    $link = $shell.CreateShortcut($_.FullName)
    if ($link.TargetPath -match "\\scoop\\apps\\` + appName + `\\") {
        Copy-Item -Path $_.FullName -Destination "$desktop\$($_.Name)" -Force
    }
}
`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", psCmd)
	return cmd.Run()
}
