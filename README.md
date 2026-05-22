# Windows Provisioning Tool

A fast, lightweight, and automated system setup tool for Windows written in Go. It automates user creation, Scoop configuration, and application installations.

---

## 🚀 Quick Start

Setting up a fresh Windows machine is split into **two phases** to handle user account creation and account switching lifecycles cleanly.

### Phase 1: Create Daily Administrator Account
1. Log in to your fresh Windows system with the built-in **Administrator** account.
2. Open **PowerShell** (the script will automatically request Administrator elevation).
3. Copy and run the following command to run the user creation tool:
   ```powershell
   irm https://raw.githubusercontent.com/DropGuard/provisioner/main/scripts/user.ps1 | iex
   ```
4. Follow the prompt to create the account (defaults to `DailyUser`). When prompted, type `y` to log off.

### Phase 2: System Provisioning
1. Log in to your newly created **DailyUser** account.
2. Open **PowerShell** (runs as a standard user).
3. Copy and run the following command to start installing Scoop and all configured applications:
   ```powershell
   irm https://raw.githubusercontent.com/DropGuard/provisioner/main/scripts/provisioner.ps1 | iex
   ```
4. Wait for the installation to complete. The script will automatically configure Scoop, add required buckets, install your apps, create desktop shortcuts, and run post-setup commands.

---

## 🛠️ Development & Compilation

The project uses a standard `Makefile` for local development.

### Local Makefile Targets:
* **Build both tools**:
  ```bash
  make build
  ```
  *(Binaries are compiled into the `bin/` directory)*
* **Clean up build artifacts**:
  ```bash
  make clean
  ```
* **Run unit tests**:
  ```bash
  make test
  ```
* **Format & Lint**:
  ```bash
  make lint
  ```
* **Run components directly**:
  ```bash
  make run-provisioner
  make run-user
  ```

---

## ⚙️ Custom Configuration

Modify [config.yaml](config.yaml) to customize the software list and commands executed during setup. The configuration file is automatically embedded into the binary during compilation using Go's `//go:embed` feature.
