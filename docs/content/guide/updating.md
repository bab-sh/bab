# Updating

## Package Managers

### Homebrew Cask (macOS/Linux)

```bash
brew upgrade --cask bab
```

### Homebrew (macOS/Linux)

```bash
brew upgrade bab
```

### Chocolatey (Windows)

```powershell
choco upgrade bab
```

### Scoop (Windows)

```powershell
scoop update bab
```

### Snapcraft (Linux)

```bash
snap refresh bab
```

### AUR (Arch Linux)

```bash
yay -S bab-bin
```

### Go

```bash
go install github.com/bab-sh/bab@latest
```

## Install Script

Re-run the install script:

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

## Linux Packages

Download the latest release from [GitHub Releases](https://github.com/bab-sh/bab/releases/latest) and reinstall:

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
