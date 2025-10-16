# Getting Started

Welcome to **Bab** - the modern task runner for all your projects.

## What is Bab?

Bab is a task runner that replaces Makefiles and npm scripts with a simple, universal solution. It lets you define custom commands in YAML and execute them with ease across any language or project.

### Key Features

- **Zero Dependencies** - No npm, Python, or other runtimes required
- **Cross-Platform** - Works on Windows, macOS, and Linux
- **Simple Syntax** - Just YAML, no complex syntax to learn
- **Fast** - Built with Go for speed and efficiency
- **Universal** - Works with any programming language or project type

## Quick Start

### 1. Install Bab

Choose your preferred installation method:

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```bash [Homebrew]
brew tap bab-sh/tap
brew install --cask bab
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

For more installation options, see the [Installation Guide](/guide/installation).

### 2. Create Your First Babfile

Create a file named `Babfile` in your project root:

```yaml
# Babfile
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

::: tip
You can also name your file `Babfile.yaml` or `Babfile.yml` - Bab will find it automatically.
:::

### 3. List Available Tasks

Run `bab list` to see all your tasks:

```bash
bab list
```

Output:

```
build Build for production
dev Start development server
setup Install dependencies
test Run test suite
```

### 4. Run Tasks

Execute any task by name:

```bash
# Run setup
bab setup

# Start development server
bab dev

# Build for production
bab build
```

### 5. Preview Commands (Dry Run)

Want to see what a task will do before running it?

```bash
bab build --dry-run
```

This shows you all the commands that would be executed without actually running them.

::: tip
Use `--dry-run` (or `-n`) to safely preview any task before execution.
:::

## Understanding the Basics

### Task Structure

Every task in a Babfile has:
- A **name** (the key, like `build` or `test`)
- A **run** field with the command(s) to execute
- An optional **desc** field for documentation

```yaml
task-name:
  desc: What this task does
  run: command to execute
```

### Multiple Commands

You can run multiple commands in sequence:

```yaml
deploy:
  desc: Build and deploy
  run:
    - npm run test
    - npm run build
    - npm run deploy:prod
```

Commands execute in order. If any command fails, execution stops.

### Nested Tasks

Organize related tasks using colon notation:

```yaml
dev:
  start:
    desc: Start dev server
    run: npm run dev

  watch:
    desc: Watch for changes
    run: npm run watch

test:
  unit:
    desc: Run unit tests
    run: npm run test:unit

  e2e:
    desc: Run E2E tests
    run: npm run test:e2e
```

Run nested tasks:

```bash
bab dev:start
bab test:unit
```

## Common Patterns

### Node.js Project

```yaml
setup:
  desc: Install dependencies
  run: npm install

dev:
  desc: Start development
  run: npm run dev

test:
  desc: Run tests
  run: npm test

build:
  desc: Build for production
  run: npm run build

clean:
  desc: Clean build artifacts
  run: rm -rf dist node_modules
```

### Go Project

```yaml
setup:
  desc: Download dependencies
  run: go mod download

dev:
  desc: Run with auto-reload
  run: air

test:
  desc: Run tests
  run: go test ./...

build:
  desc: Build binary
  run: go build -o app

clean:
  desc: Clean build artifacts
  run: rm -f app
```

### Python Project

```yaml
setup:
  desc: Create virtualenv and install deps
  run:
    - python -m venv venv
    - source venv/bin/activate && pip install -r requirements.txt

dev:
  desc: Run development server
  run: python manage.py runserver

test:
  desc: Run tests
  run: pytest

lint:
  desc: Lint code
  run: flake8 .
```

## Command-Line Options

### Verbose Output

Get detailed execution information:

```bash
bab build --verbose
# or
bab build -v
```

### Dry Run

Preview commands without executing:

```bash
bab deploy --dry-run
# or
bab deploy -n
```

### Get Help

```bash
bab --help          # General help
bab list --help     # Command-specific help
```

### Check Version

```bash
bab --version
```

## What's Next?

Now that you have the basics down:

- **[Installation Guide](/guide/installation)** - Learn about all installation methods
- **[Babfile Syntax](/guide/babfile-syntax)** - Master the Babfile syntax
- **[CLI Reference](/guide/cli-reference)** - Explore all CLI commands and flags

## Need Help?

- Join our [Discord community](https://discord.bab.sh)
- Check the [GitHub repository](https://github.com/bab-sh/bab)
- Report issues on [GitHub Issues](https://github.com/bab-sh/bab/issues)
