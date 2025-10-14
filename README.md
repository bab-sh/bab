<div align="center">

<img src="assets/bab.png" alt="Bab Logo" width="200"/>

<h1>Bab</h1>

**Custom commands for every project**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bab-sh/bab)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/bab-sh/bab)](https://github.com/bab-sh/bab/releases)
[![Status](https://img.shields.io/badge/Status-Pre--Alpha-red.svg)](https://github.com/bab-sh/bab)

[Website](https://bab.sh) • [Documentation](https://github.com/bab-sh/bab#readme) • [Installation](#installation) • [Discord](https://discord.bab.sh) • [Contributing](#contributing)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Update](#update)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [Babfile Structure](#babfile-structure)
  - [Command Reference](#command-reference)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

---

## Overview
**Bab** is a modern task runner built for seamless development workflows. It replaces the clunky syntax of Makefiles and the limitations of npm scripts with a universal, dependency free solution that works across any language or project. Designed with developer experience at its core, Bab is simple when you want it to be and powerful when you need it to be. Whether you're running small scripts or orchestrating hundreds of tasks, Bab scales effortlessly, keeping your workflow smooth, organized, and maintainable.

---
## Quick Start

### 1. Install Bab

```bash
# macOS / Linux
curl -sSfL https://bab.sh/install.sh | sh

# Windows (PowerShell)
iwr -useb https://bab.sh/install.ps1 | iex
```

### 2. Create a Babfile

Create a `Babfile` in your project root:

```yaml
setup:
  desc: Install dependencies
  run: npm install

dev:
  desc: Start development server
  run: npm run dev

test:
  desc: Run test suite
  run: npm test

build:
  desc: Build for production
  run: npm run build
```

### 3. Run Your Tasks

```bash
bab               # List all available tasks in the handy fuzzy finder
bab setup         # Run the setup task
bab dev           # Start development server
```

---

## Installation

Bab provides multiple installation methods to fit your preferred workflow.

### Quick Install

#### macOS / Linux

```bash
# Install latest version
curl -sSfL https://bab.sh/install.sh | sh

# Install specific version
curl -sSfL https://bab.sh/install.sh | sh -s -- v1.0.0

# Using wget
wget -qO- https://bab.sh/install.sh | sh
```

#### Windows (PowerShell)

```powershell
# Install latest version
iwr -useb https://bab.sh/install.ps1 | iex

# Install specific version
iwr -useb https://bab.sh/install.ps1 | iex -Version v1.0.0
```

### Package Managers

#### Homebrew (macOS/Linux)

```bash
# Tap the repository
brew tap bab-sh/tap

# Install as formula
brew install bab

# Or install as cask (recommended)
brew install --cask bab
```

#### Chocolatey (Windows)

```powershell
choco install bab
```

#### Scoop (Windows)

```powershell
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket
scoop install bab
```

#### Snapcraft (Linux)

```bash
snap install bab-sh
```

#### AUR (Arch Linux)

```bash
# Pre-built binary
yay -S bab-bin

# Build from source
yay -S bab
```

#### Go Install

```bash
go install github.com/bab-sh/bab@latest
```

### Linux Packages

Download packages from [GitHub Releases](https://github.com/bab-sh/bab/releases/latest):

```bash
# Debian/Ubuntu (.deb)
sudo dpkg -i bab_*_amd64.deb

# Fedora/RHEL (.rpm)
sudo rpm -i bab_*.x86_64.rpm

# Alpine (.apk)
sudo apk add --allow-untrusted bab_*.apk

# Arch Linux (.pkg.tar.zst)
sudo pacman -U bab_*.pkg.tar.zst
```

### Build from Source

Requires Go 1.21 or later:

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build
go build -o bab

# Install (optional)
sudo mv bab /usr/local/bin/
```

---

## Update

### Quick Update

The install scripts automatically detect existing installations and update them:

#### macOS / Linux

```bash
# Update to latest version
curl -sSfL https://bab.sh/install.sh | sh

# Update to specific version
curl -sSfL https://bab.sh/install.sh | sh -s -- v1.0.0
```

#### Windows (PowerShell)

```powershell
# Update to latest version
iwr -useb https://bab.sh/install.ps1 | iex

# Update to specific version
iwr -useb https://bab.sh/install.ps1 | iex -Version v1.0.0
```

### Package Manager Updates

#### Homebrew (macOS/Linux)

```bash
# Update via formula
brew upgrade bab

# Update via cask
brew upgrade --cask bab
```

#### Chocolatey (Windows)

```powershell
choco upgrade bab
```

#### Scoop (Windows)

```powershell
scoop update bab
```

#### Snapcraft (Linux)

```bash
snap refresh bab-sh
```

#### AUR (Arch Linux)

```bash
# Update with yay
yay -Syu bab-bin

# Or with paru
paru -Syu bab-bin
```

#### Go Install

```bash
go install github.com/bab-sh/bab@latest
```

#### Linux Packages

Download the latest package from [releases](https://github.com/bab-sh/bab/releases/latest) and reinstall:

```bash
# Debian/Ubuntu
sudo dpkg -i bab_*_amd64.deb

# Fedora/RHEL
sudo rpm -U bab_*.x86_64.rpm

# Alpine
sudo apk add --allow-untrusted bab_*.apk

# Arch Linux
sudo pacman -U bab_*.pkg.tar.zst
```

---

## Usage

### Basic Usage

```bash
# List all available tasks
bab

# Run a task
bab <task-name>

# Run nested tasks
bab <parent>:<child>

# Get help
bab --help

# Show version
bab --version
```

### Babfile Structure

A Babfile is a YAML file that defines your project's tasks. Here's the basic structure:

```yaml
# Simple task
task-name:
  desc: Description of what this task does
  run: command to execute

# Task with multiple commands
build:
  desc: Build the project
  run: |
    echo "Building..."
    go build -o app
    echo "Build complete!"

# Nested tasks
dev:
  start:
    desc: Start development server
    run: npm run dev

  test:
    desc: Run tests in watch mode
    run: npm run test:watch

# Task without description
quick-test:
  run: go test ./...
```

### Command Reference

```bash
# Basic commands
bab                              # List all tasks
bab <task>                       # Run a task
bab <parent>:<child>             # Run nested task

# Options
bab <task> --dry-run             # Show what would run without executing
bab <task> --verbose             # Show detailed output
bab --file custom.yaml <task>    # Use a custom Babfile
bab completion bash              # Generate bash completion script
bab completion zsh               # Generate zsh completion script
bab completion fish              # Generate fish completion script

# Help and information
bab --help                       # Show help information
bab --version                    # Show version information
```

---

## Roadmap

Bab is under active development. Some completed features may be refined or reimplemented as the project matures.

- [x] **Interactive Mode** - Fuzzy search interface for browsing and selecting tasks
- [x] **Task History Tracking** - Per-project execution history with status and duration
- [x] **Nested Task Support** - Organize tasks hierarchically with colon notation
- [x] **Cross-Platform Execution** - Native shell execution on Linux, macOS, and Windows
- [x] **Shell Completions** - Tab completion for bash, zsh, and fish shells
- [x] **Flexible Babfile Formats** - Support for Babfile, Babfile.yaml, and Babfile.yml
- [x] **Dry-Run Mode** - Preview commands without executing them
- [x] **Custom File Paths** - Specify alternative Babfile locations
- [ ] **Babfile Schema Validation** - Structured schema for validating Babfile syntax and configuration
- [ ] **Advanced Multi-Babfile Support** - Import and compose multiple Babfiles for complex project structures
- [ ] **Platform-Specific Tasks** - Define tasks that run only on specific operating systems (Windows, Linux, macOS)
- [ ] **Compiled Binary Execution** - Support for executing compiled binaries beyond shell scripts
- [ ] **Environment Variable Management** - Built-in handling, interpolation, and validation of environment variables
- [ ] **Task Dependencies** - Automatic execution of prerequisite tasks in the correct order
- [ ] **Module System** - Share and reuse Babfiles across projects as importable modules
- [ ] **Performance Profiling** - Built-in performance monitoring and profiling for task execution
- [ ] **Interactive Babfile Generator** - Create and modify Babfiles through interactive command-line forms
- [ ] **Enhanced Output Formatting** - Improved colored output with better visual hierarchy

See our [GitHub Issues](https://github.com/bab-sh/bab/issues) for the complete list of planned features and to suggest new ideas.

---

## Contributing

We welcome contributions from the community. Whether it's bug reports, feature requests, or code contributions, we appreciate your help.

### How to Contribute

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/bab-sh/bab.git
cd bab

# Install dependencies
go mod download

# Build
go build -o bab

# Run tests
go test ./...
```

### Guidelines

- Write clear, concise commit messages
- Add tests for new features
- Update documentation as needed
- Follow Go best practices and idioms
- Run tests and lint before submitting PRs

---

## Support

Need help? We're here for you:

- 💬 **Discord** - Join our [Discord community](https://discord.bab.sh) for questions and discussions
- 🐛 **Bug Reports** - [GitHub Issues](https://github.com/bab-sh/bab/issues) for bug reports
- 🌐 **Website** - Visit [bab.sh](https://bab.sh) for more information

---

## License

Bab is released under the [MIT License](LICENSE). This means you're free to use, modify, and distribute this software for personal or commercial purposes.

---

<div align="center">

**[⬆ back to top](#table-of-contents)**

Built with ❤️ by AIO

</div>
