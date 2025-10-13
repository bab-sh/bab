# Bab - Custom Commands for your Projects

**Zero dependencies. Cross-platform by design.**

Write custom commands in YAML, execute them with bab. Just like Makefile but with more features and Zero dependencies. (soon)

ITS PRE PRE ALPHA DO NOT USE IT FRFR

### TODOS
- [x] Last commands executed history
- [ ] Basic babfile schema
- [ ] Advanced nesting and multi Babfile support
- [ ] Platform flags for commands so you can define commands specifically for OS
- [ ] Binary Compile (so its not just shell and batch scripts)
- [ ] Babfile configs and environment variable handeling
- [ ] Homepage
- [ ] Being a good markdown/taskfile replacement

### IDEAS
- [ ] Performance profiling
- [x] Interactive mode (fuzzy search or something)
- [ ] Modules (multi babfile support in module format)
- [ ] Generate babfiles/generate and add new commands to babfile via forms


## Installation

### Quick Install (Recommended)

One line install script for all platforms

#### macOS / Linux

```bash
# Install latest version
curl -sSfL https://bab.sh/install.sh | sh

# Install specific version
curl -sSfL https://bab.sh/install.sh | sh -s v1.0.0

# Using wget
wget -qO- https://bab.sh/install.sh | sh
```

#### Windows (PowerShell)

```powershell
# Install latest version
iwr -useb https://bab.sh/install.ps1 | iex

# Install specific version
$version="v1.0.0"; iwr -useb https://bab.sh/install.ps1 | iex
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

### Scoop (Windows)

```bash
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket
scoop install bab
```

### Snapcraft (Linux)

```bash
snap install bab-sh
```

### Linux Packages

```bash
# Debian/Ubuntu (.deb)
# Download from https://github.com/bab-sh/bab/releases/latest
sudo dpkg -i bab_*_amd64.deb

# Fedora/RHEL (.rpm)
# Download from https://github.com/bab-sh/bab/releases/latest
sudo rpm -i bab_*.x86_64.rpm

# Alpine (.apk)
# Download from https://github.com/bab-sh/bab/releases/latest
sudo apk add --allow-untrusted bab_*.apk

# Arch Linux (.pkg.tar.zst)
# Download from https://github.com/bab-sh/bab/releases/latest
sudo pacman -U bab_*.pkg.tar.zst

# Or use AUR
yay -S bab-bin       # Pre-built binary
yay -S bab           # Build from source
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build bab
go build -o bab

# Move to your PATH (optional)
sudo mv bab /usr/local/bin/
```

### 2. Create a Babfile

```yaml
# Babfile
setup:
  desc: Setup development environment
  run: npm install

dev:
  start:
    desc: Start development server
    run: npm run dev

test:
  desc: Run tests
  run: npm test

build:
  desc: Build for production
  run: npm run build
```

### 3. Run your tasks

```bash
bab                    # List all tasks
bab setup              # Run setup task
bab dev:start          # Run nested task
```

## 📋 Commands

```bash
# Basic usage
bab                    # List all tasks
bab <task>             # Run a task

# Example task usage
bab setup              # Run setup task
bab dev:start          # Run nested task

# Options
bab <task> --dry-run   # Show what would run
bab <task> --verbose   # Verbose output
bab --file custom.yaml <task>  # Use custom Babfile
```

## 📜 License

MIT License - Free for personal and commercial use.

---

**Made with ❤️ by aio for developers who value simplicity and reliability.**
