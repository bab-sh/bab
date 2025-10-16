# Installation

Bab provides multiple installation methods to fit your workflow. Choose the one that works best for you.

## Quick Install

The fastest way to get started:

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```bash [wget]
wget -qO- https://bab.sh/install.sh | sh
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

::: tip Version-Specific Installation
Install a specific version by appending the version:

```bash
# macOS/Linux
curl -sSfL https://bab.sh/install.sh | sh -s -- v1.0.0

# Windows
iwr -useb https://bab.sh/install.ps1 | iex -Version v1.0.0
```
:::

## Package Managers

### Homebrew (macOS/Linux)

```bash
# Add the bab tap
brew tap bab-sh/tap

# Install as cask (recommended)
brew install --cask bab

# Or install as formula
brew install bab
```

**Update:**
```bash
brew upgrade --cask bab
```

### Chocolatey (Windows)

```powershell
# Install
choco install bab

# Update
choco upgrade bab
```

### Scoop (Windows)

```powershell
# Add bucket
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket

# Install
scoop install bab

# Update
scoop update bab
```

### Snapcraft (Linux)

```bash
# Install
snap install bab-sh

# Update
snap refresh bab-sh
```

### AUR (Arch Linux)

::: code-group

```bash [Pre-built Binary]
yay -S bab-bin
```

```bash [Build from Source]
yay -S bab
```

```bash [paru]
paru -S bab-bin
```

:::

**Update:**
```bash
yay -Syu bab-bin
```

### Go Install

```bash
# Install
go install github.com/bab-sh/bab@latest

# This automatically installs to $GOPATH/bin
# Make sure $GOPATH/bin is in your PATH
```

## Linux Packages

Download packages from [GitHub Releases](https://github.com/bab-sh/bab/releases/latest).

### Debian/Ubuntu (.deb)

```bash
# Download latest .deb package
wget https://github.com/bab-sh/bab/releases/latest/download/bab_*_amd64.deb

# Install
sudo dpkg -i bab_*_amd64.deb

# Update (download new version and reinstall)
sudo dpkg -i bab_*_amd64.deb
```

### Fedora/RHEL (.rpm)

```bash
# Download latest .rpm package
wget https://github.com/bab-sh/bab/releases/latest/download/bab_*.x86_64.rpm

# Install
sudo rpm -i bab_*.x86_64.rpm

# Update
sudo rpm -U bab_*.x86_64.rpm
```

### Alpine (.apk)

```bash
# Download latest .apk package
wget https://github.com/bab-sh/bab/releases/latest/download/bab_*.apk

# Install
sudo apk add --allow-untrusted bab_*.apk
```

### Arch Linux (.pkg.tar.zst)

```bash
# Download latest package
wget https://github.com/bab-sh/bab/releases/latest/download/bab_*.pkg.tar.zst

# Install
sudo pacman -U bab_*.pkg.tar.zst
```

## Build from Source

Requirements:
- Go 1.25 or later
- Git

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build
go build -o bab

# Install (optional)
sudo mv bab /usr/local/bin/

# Or install to custom location
mkdir -p ~/.local/bin
mv bab ~/.local/bin/
# Add ~/.local/bin to your PATH if not already there
```

::: tip
Building from source gives you the latest development version. For stable releases, use one of the package manager methods above.
:::

## Verify Installation

After installing, verify that bab is working:

```bash
# Check version
bab --version

# View help
bab --help

# List available commands
bab list
```

You should see output similar to:

```
Custom commands for every project

Usage:
  bab [command]

Available Commands:
  completion  Generate autocompletion script
  help        Help about any command
  list        List all available tasks

Flags:
  -n, --dry-run   Show commands without executing
  -h, --help      help for bab
  -v, --verbose   Enable verbose output
      --version   version for bab
```

## Updating Bab

### Quick Update

The install scripts automatically detect and update existing installations:

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

### Package Manager Updates

See the package manager sections above for update commands specific to each tool.

## Uninstalling

### Homebrew

```bash
brew uninstall --cask bab
brew untap bab-sh/tap
```

### Chocolatey

```powershell
choco uninstall bab
```

### Scoop

```powershell
scoop uninstall bab
```

### Snap

```bash
snap remove bab-sh
```

### Manual Installation

Simply remove the bab binary:

::: code-group

```bash [macOS/Linux]
sudo rm /usr/local/bin/bab
```

```powershell [Windows]
del "C:\Program Files\bab\bab.exe"
```

:::

## Troubleshooting

### Command not found

If you get "command not found" after installation:

1. **Check if bab is in your PATH:**
   ```bash
   which bab    # macOS/Linux
   where bab    # Windows
   ```

2. **Add to PATH if needed:**

   For `~/.local/bin`:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="$HOME/.local/bin:$PATH"
   ```

   For Go install:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="$GOPATH/bin:$PATH"
   ```

3. **Restart your terminal** after updating PATH.

### Permission denied

If you get permission errors:

```bash
# Make bab executable
chmod +x /path/to/bab

# Or if installing to system directory
sudo mv bab /usr/local/bin/
```

### Go version mismatch

If building from source fails due to Go version:

```bash
# Check your Go version
go version

# Update Go to 1.25 or later
# Visit: https://golang.org/dl/
```

## Next Steps

Now that you have bab installed:

- **[Get Started](/guide/getting-started)** - Create your first Babfile
- **[Babfile Syntax](/guide/babfile-syntax)** - Learn the Babfile syntax
- **[CLI Reference](/guide/cli-reference)** - Explore all CLI commands

## Need Help?

- Join our [Discord community](https://discord.bab.sh)
- Report issues on [GitHub](https://github.com/bab-sh/bab/issues)
- Check the [documentation](https://docs.bab.sh)
