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

Create a `Babfile.yml` in your project root:

```yaml
env:
  NODE_ENV: production

tasks:
  setup:
    desc: Install dependencies
    run:
      - cmd: npm install

  dev:
    desc: Start development server
    deps: [setup]
    env:
      PORT: "3000"
    run:
      - cmd: npm run dev

  test:
    desc: Run tests
    deps: [setup]
    run:
      - cmd: npm test

  build:
    desc: Build for production
    deps: [setup, test]
    run:
      - cmd: npm run build
```

Environment variables can be set at global, task, or command level. See [Babfile Syntax](/guide/babfile-syntax#environment-variables) for details.

## Usage

```bash
# Browse tasks interactively
bab

# List tasks
bab --list

# Run a task
bab dev

# Preview without executing
bab build --dry-run

# Verbose output
bab build --verbose
```

## Interactive Mode

Running `bab` with no arguments launches an interactive fuzzy finder that lets you browse and select tasks. This is especially useful when you have many tasks or can't remember the exact task name. Type to search, use arrow keys to navigate, and press `Enter` to execute the selected task.

## Task Dependencies

Dependencies run automatically before the main task:

```bash
bab build
# Runs: setup → test → build
```

## Namespaced Tasks

Use colon notation for logical groupings:

```yaml
tasks:
  dev:start:
    desc: Start server
    run:
      - cmd: npm run dev

  dev:watch:
    desc: Watch files
    run:
      - cmd: npm run watch
```

Run with `bab dev:start`.

## Next Steps

- [Babfile Syntax](/guide/babfile-syntax) - Learn the YAML format
- [CLI Reference](/guide/cli-reference) - See all commands
