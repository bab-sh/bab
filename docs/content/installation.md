# Installation

Bab is designed to be easy to install across all major platforms. Choose the method that works best for you.

## Prerequisites

For most installation methods, you don't need anything special. However, if you're building from source, you'll need:

- **Go 1.25.1 or later** (for building from source)
- **Git** (for cloning the repository)

## Installation Methods

### macOS

#### Homebrew (Recommended)

The easiest way to install bab on macOS is through Homebrew:

```bash
# Add the bab tap
brew tap bab-sh/tap

# Install bab
brew install --cask bab
```

To verify the installation:

```bash
bab --version
```

To update bab later:

```bash
brew upgrade --cask bab
```

### Linux

#### Build from Source

On Linux, the recommended method is to build from source:

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build bab
go build -o bab

# Move to your PATH (requires sudo)
sudo mv bab /usr/local/bin/

# Verify installation
bab --version
```

#### Manual Installation

If you prefer to install bab to a custom location:

```bash
# Build bab
go build -o bab

# Move to a directory in your PATH
mkdir -p ~/.local/bin
mv bab ~/.local/bin/

# Add to PATH if not already (add this to your ~/.bashrc or ~/.zshrc)
export PATH="$HOME/.local/bin:$PATH"
```

### Windows

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build bab
go build -o bab.exe

# Move to a directory in your PATH
# For example, C:\Program Files\bab\
mkdir "C:\Program Files\bab"
move bab.exe "C:\Program Files\bab\"

# Add to PATH using PowerShell (run as Administrator)
$env:Path += ";C:\Program Files\bab"
[Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::Machine)
```

Verify the installation:

```bash
bab --version
```

## Alternative: Use Compiled Scripts

If you don't want to install bab globally, you can use it in a specific project by compiling your Babfile to standalone scripts:

1. Install bab temporarily or on another machine
2. Run `bab compile` in your project directory
3. Distribute the generated `bab.sh` and `bab.bat` files with your project
4. Team members can run tasks without installing bab:

```bash
./bab.sh <task>    # Unix/Linux/macOS
bab.bat <task>     # Windows
```

## Verifying Installation

After installation, verify that bab is working correctly:

```bash
# Check version
bab --version

# View help
bab --help
```

## Updating Bab

### Homebrew (macOS)

```bash
brew upgrade --cask bab
```

### From Source

```bash
cd bab
git pull origin main
go build -o bab
sudo mv bab /usr/local/bin/
```

## Uninstalling

### Homebrew (macOS)

```bash
brew uninstall --cask bab
brew untap bab-sh/tap
```

### Manual Installation

Simply remove the bab binary:

```bash
# Linux/macOS
sudo rm /usr/local/bin/bab

# Windows
del "C:\Program Files\bab\bab.exe"
```

## Next Steps

Now that you have bab installed, you can:

1. [Get Started](/get-started) with your first Babfile
2. Learn about [Babfile Syntax](/syntax)
3. Explore all [Features](/features)
