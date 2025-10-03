# Script Compilation

One of Bab's most powerful features is the ability to compile your Babfile into standalone shell scripts. This enables **zero-dependency distribution** of your tasks.

## What is Compilation?

Compilation transforms your Babfile into two standalone scripts:

- **`bab.sh`** - Shell script for Unix/Linux/macOS
- **`bab.bat`** - Batch file for Windows

These scripts:
- Contain all your tasks
- Require **no external dependencies**
- Work on any system with a shell
- Can be distributed with your project

## Why Compile?

### 1. Zero Dependencies

Team members don't need to install bab to run tasks:

```bash
# Without compilation - requires bab installation
bab build

# With compilation - no installation needed
./bab.sh build
```

### 2. Easy Distribution

Include compiled scripts in your repository:

```bash
git add bab.sh bab.bat
git commit -m "Add task runner scripts"
```

Now everyone can run tasks immediately after cloning.

### 3. CI/CD Integration

Use in build pipelines without installing bab:

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build
        run: ./bab.sh build
```

### 4. Portability

Scripts work anywhere:
- Docker containers
- CI/CD runners
- Restricted environments
- Air-gapped systems

## How to Compile

### Basic Usage

In your project directory with a Babfile:

```bash
bab compile
```

Output:

```
INFO  Using Babfile path=Babfile
INFO  Output directory path=.
INFO  Compiling Babfile to scripts
INFO  Successfully compiled Babfile to scripts!
INFO  Generated script path=bab.sh
INFO  Generated script path=bab.bat
```

### Custom Output Directory

Generate scripts in a specific directory:

```bash
bab compile -o scripts/
# or
bab compile --output dist/
```

This creates:
- `scripts/bab.sh`
- `scripts/bab.bat`

### Disable Colors

Generate scripts without color output:

```bash
bab compile --no-color
```

Useful for:
- Systems without color support
- Log files
- CI/CD environments with plain text output

## Using Compiled Scripts

### Unix/Linux/macOS

Make the script executable (first time only):

```bash
chmod +x bab.sh
```

Run tasks:

```bash
# List all tasks
./bab.sh

# Run a specific task
./bab.sh build

# Run nested tasks
./bab.sh dev:start
```

### Windows

Run tasks directly:

```cmd
REM List all tasks
bab.bat

REM Run a specific task
bab.bat build

REM Run nested tasks
bab.bat dev:start
```

## Script Contents

### Shell Script (bab.sh)

The generated shell script includes:

- Task function definitions
- Task listing functionality
- Error handling
- Color output (unless `--no-color`)
- Help information

Example structure:

```bash
#!/bin/sh
set -e

# Color codes (if colors enabled)
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Task functions
task_build() {
  echo "${BLUE}Running task: build${NC}"
  go build -o myapp
  echo "${GREEN}Task completed: build${NC}"
}

task_test() {
  echo "${BLUE}Running task: test${NC}"
  go test ./...
  echo "${GREEN}Task completed: test${NC}"
}

# Task listing and execution logic
# ...
```

### Batch File (bab.bat)

The generated batch file includes similar functionality for Windows:

```batch
@echo off
setlocal enabledelayedexpansion

REM Task implementations
:task_build
  echo Running task: build
  go build -o myapp.exe
  echo Task completed: build
  goto :eof

:task_test
  echo Running task: test
  go test ./...
  echo Task completed: test
  goto :eof

REM Task listing and execution logic
REM ...
```

## Complete Example

### Setup

Create a Babfile:

```yaml
# Babfile
setup:
  desc: Install dependencies
  run: npm install

dev:
  start:
    desc: Start development server
    run: npm run dev

build:
  desc: Build for production
  run:
    - npm run test
    - npm run build
```

Compile it:

```bash
bab compile
```

### Using the Scripts

List tasks:

```bash
$ ./bab.sh
Available tasks:
  setup       Install dependencies
  dev:start   Start development server
  build       Build for production
```

Run a task:

```bash
$ ./bab.sh setup
▶ Running task: setup
  npm install
  added 234 packages
✓ Task completed: setup
```

Run nested task:

```bash
$ ./bab.sh dev:start
▶ Running task: dev:start
  Starting development server on http://localhost:3000
```

## Workflow Integration

### Development Workflow

1. **Developer**: Create Babfile with tasks
2. **Developer**: Compile to scripts: `bab compile`
3. **Developer**: Commit scripts to repository
4. **Team**: Clone repository
5. **Team**: Run tasks without installing bab

### CI/CD Example

```yaml
# .github/workflows/build.yml
name: Build

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Setup Node
        uses: actions/setup-node@v2
        with:
          node-version: '18'

      # No bab installation needed!
      - name: Install Dependencies
        run: ./bab.sh setup

      - name: Run Tests
        run: ./bab.sh test

      - name: Build
        run: ./bab.sh build
```

### Docker Example

```dockerfile
FROM node:18-alpine

WORKDIR /app

# Copy project files including compiled scripts
COPY . .

# Use compiled script - no bab installation needed
RUN ./bab.sh setup
RUN ./bab.sh build

CMD ["./bab.sh", "start"]
```

## Updating Compiled Scripts

When you modify your Babfile, recompile:

```bash
bab compile
```

Best practices:
1. Update Babfile
2. Recompile scripts
3. Test scripts locally
4. Commit all changes together

## Version Control

### What to Commit

**Option 1: Commit compiled scripts (Recommended)**

```bash
git add Babfile bab.sh bab.bat
git commit -m "Update tasks"
```

Benefits:
- Team doesn't need bab installed
- Immediate task availability
- CI/CD ready

**Option 2: Don't commit scripts**

Add to `.gitignore`:

```gitignore
bab.sh
bab.bat
```

When to use:
- Team has bab installed
- Prefer smaller repository
- Scripts regenerated in CI/CD

### .gitignore Example

If not committing scripts:

```gitignore
# Bab compiled scripts
bab.sh
bab.bat

# Or ignore in specific directory
dist/bab.sh
dist/bab.bat
```

## Advanced Usage

### Multiple Babfiles

Compile different task files:

```bash
# Compile development tasks
bab --file tasks/dev.yml compile -o dist/dev/

# Compile production tasks
bab --file tasks/prod.yml compile -o dist/prod/
```

Use the appropriate script:

```bash
dist/dev/bab.sh start
dist/prod/bab.sh deploy
```

### Script Customization

The generated scripts are standalone - you can:

1. Copy them to other projects
2. Modify them for specific needs
3. Distribute them independently

However, remember that manual changes will be overwritten when you recompile.

## Troubleshooting

### Permission Denied (Unix/Linux/macOS)

Make the script executable:

```bash
chmod +x bab.sh
```

### Script Not Found

Ensure you're in the correct directory:

```bash
ls -la bab.sh  # Check file exists
./bab.sh       # Run from current directory
```

### Colors Not Showing

Some terminals don't support colors. Compile without colors:

```bash
bab compile --no-color
```

### Task Not Working in Compiled Script

Verify the task works with bab first:

```bash
bab <task>              # Test with bab
bab compile             # Recompile
./bab.sh <task>         # Test compiled version
```

## Limitations

Compiled scripts:
- Are **static snapshots** of your Babfile
- Must be **regenerated** after Babfile changes
- Don't support future bab features automatically
- May be **larger** than the original Babfile

For these reasons, use compilation primarily for distribution, not as a replacement for bab during development.

## Next Steps

- Learn more about [Features](/features)
- Check the [Syntax Guide](/syntax)
- Read about [Installation](/installation) for your team
