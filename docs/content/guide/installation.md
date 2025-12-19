# Installation

## Quick Install

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

## Package Managers

### Homebrew Cask (macOS/Linux)

```bash
brew tap bab-sh/tap
brew install --cask bab
```

### Homebrew (macOS/Linux)

```bash
brew tap bab-sh/tap
brew install bab
```

### Chocolatey (Windows)

```powershell
choco install bab
```

### Scoop (Windows)

```powershell
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket
scoop install bab
```

### Snapcraft (Linux)

```bash
snap install bab-sh
```

### AUR (Arch Linux)

```bash
# Binary package (pre-built)
yay -S bab-bin

# Source package (builds from source)
yay -S bab
```

### Go

```bash
go install github.com/bab-sh/bab@latest
```

## Manual Download

Download the archive for your platform from [GitHub Releases](https://github.com/bab-sh/bab/releases/latest):

| Platform | File |
|----------|------|
| macOS (Intel) | `bab_*_macOS_x86_64.tar.gz` |
| macOS (Apple Silicon) | `bab_*_macOS_arm64.tar.gz` |
| macOS (Universal) | `bab_*_macOS_universal.tar.gz` |
| Linux (x64) | `bab_*_Linux_x86_64.tar.gz` |
| Linux (ARM64) | `bab_*_Linux_arm64.tar.gz` |
| Linux (ARMv7) | `bab_*_Linux_armv7.tar.gz` |
| Windows (x64) | `bab_*_Windows_x86_64.zip` |

Extract and move to your PATH:

```bash
tar -xzf bab_*.tar.gz
sudo mv bab /usr/local/bin/
```

## Linux Packages

### Debian/Ubuntu

```bash
sudo dpkg -i bab_*_amd64.deb
```

### Fedora/RHEL

```bash
sudo rpm -i bab_*.x86_64.rpm
```

### Alpine

```bash
sudo apk add --allow-untrusted bab_*.apk
```

### Arch Linux

```bash
sudo pacman -U bab_*.pkg.tar.zst
```

## Build from Source

```bash
git clone https://github.com/bab-sh/bab.git
cd bab
go build -o bab
sudo mv bab /usr/local/bin/
```

Requires Go 1.23+.

## Shell Completions

Bab supports bash, zsh, and fish completions. They are automatically installed with package managers.

For manual installation:

::: code-group

```bash [Bash]
bab --completion bash > ~/.local/share/bash-completion/completions/bab
```

```bash [Zsh]
bab --completion zsh > ~/.local/share/zsh/site-functions/_bab
```

```bash [Fish]
bab --completion fish > ~/.config/fish/completions/bab.fish
```

:::
