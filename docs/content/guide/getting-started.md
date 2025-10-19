# Getting Started

## Install

```bash
# macOS / Linux
curl -sSfL https://bab.sh/install.sh | sh

# Windows (PowerShell)
iwr -useb https://bab.sh/install.ps1 | iex
```

See [installation guide](/guide/installation) for other methods.

## Create Babfile

Create a `Babfile` in your project root:

```yaml
setup:
  desc: Install dependencies
  run: npm install

dev:
  desc: Start development server
  deps: setup
  run: npm run dev

test:
  desc: Run tests
  deps: setup
  run: npm test

build:
  desc: Build for production
  deps: [setup, test]
  run: npm run build
```

## Usage

```bash
# List tasks
bab list

# Run a task
bab dev

# Preview without executing
bab build --dry-run

# Verbose output
bab build --verbose
```

## Task Dependencies

Dependencies run automatically before the main task:

```bash
bab build
# Runs: setup → test → build
```

## Nested Tasks

```yaml
dev:
  start:
    desc: Start server
    run: npm run dev

  watch:
    desc: Watch files
    run: npm run watch
```

Run with `bab dev:start`.

## Next Steps

- [Babfile Syntax](/guide/babfile-syntax) - Learn the YAML format
- [CLI Reference](/guide/cli-reference) - See all commands
