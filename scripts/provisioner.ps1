# scripts/provisioner.ps1 - Bootstrapper to download and run provisioner-[arch].exe
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
}

$repo = "DropGuard/provisioner"
$url = "https://github.com/$repo/releases/latest/download/provisioner-$arch.exe"
$tempDir = Join-Path $env:SystemDrive "provisioner-setup"
if (-not (Test-Path $tempDir)) {
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
}
$output = Join-Path $tempDir "provisioner.exe"

Write-Host "Windows Architecture Detected: $arch"
Write-Host "Downloading provisioner-$arch.exe from GitHub..."
Invoke-WebRequest -Uri $url -OutFile $output -UseBasicParsing

Write-Host "Starting system provisioning..."
& $output

