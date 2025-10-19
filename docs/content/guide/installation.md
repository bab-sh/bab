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
yay -S bab-bin
```

### Go

```bash
go install github.com/bab-sh/bab@latest
```

## Linux Packages

Download from [GitHub Releases](https://github.com/bab-sh/bab/releases/latest).

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

Requires Go 1.25+.
