# Getting Started

Welcome to **Bab** - the simple, cross-platform task runner for your projects!

## What is Bab?

Bab is a task runner that lets you define custom commands in YAML and execute them with ease. Think of it as a modern alternative to Makefiles and Taskfiles, but with zero dependencies and cross-platform support built-in.

### Key Benefits

- **Zero Dependencies**: No npm, Python, or other runtimes required
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Simple**: Just YAML - no complex syntax to learn
- **Portable**: Compile to standalone scripts for distribution

## Quick Start

### 1. Install Bab

Choose your preferred installation method:

#### macOS (Homebrew)

```bash
brew tap bab-sh/tap
brew install --cask bab
```

#### Build from Source

```bash
git clone https://github.com/bab-sh/bab.git
cd bab
go build -o bab
sudo mv bab /usr/local/bin/
```

For more installation options, see the [Installation Guide](/installation).

### 2. Create Your First Babfile

Create a file named `Babfile` in your project root:

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

### 3. List Available Tasks

Run `bab` without arguments to see all available tasks:

```bash
bab
```

This will display:

```
Available tasks:
  setup      Setup development environment
  dev:start  Start development server
  test       Run tests
  build      Build for production
```

### 4. Run Tasks

Execute any task by name:

```bash
# Run a simple task
bab setup

# Run a nested task
bab dev:start

# Preview what would run (dry-run)
bab build --dry-run

# Run with verbose output
bab test --verbose
```

### 5. (Optional) Compile to Scripts

Want to distribute your tasks without requiring bab installation?

```bash
bab compile
```

This generates:
- `bab.sh` - Standalone script for Unix/Linux/macOS
- `bab.bat` - Standalone script for Windows

Now your team can run tasks without installing bab:

```bash
./bab.sh setup         # Unix/Linux/macOS
bab.bat setup          # Windows
```

## What's Next?

- Learn about [Babfile Syntax](/syntax) to create more complex tasks
- Explore all [Features](/features) available in bab
- Deep dive into [Script Compilation](/compile) for zero-dependency distribution

## Need Help?

- Check out the [GitHub repository](https://github.com/bab-sh/bab)
- Read the full [documentation](/features)
