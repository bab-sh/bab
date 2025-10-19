# CLI Reference

Complete reference for all Bab command-line interface commands and options.

## Overview

```bash
bab [command] [flags]
bab <task-name> [flags]
```

Bab can be used in two ways:
1. **Execute built-in commands** - Like `list`, `help`, or `completion`
2. **Execute tasks** - Run tasks defined in your Babfile

## Built-in Commands

### `bab list`

List all available tasks from your Babfile in a tree structure.

**Aliases:** `ls`, `tasks`

```bash
bab list
```

**Example output:**
```
build Build for production
dev
├── clean Clean development artifacts
├── start Start development server
└── watch Watch for file changes
test
├── all Run all tests
├── integration Run integration tests
└── unit Run unit tests
```

**Options:**
- `-v, --verbose` - Show verbose output including debug information
- `-h, --help` - Show help for the list command

::: tip
Use `bab list` to quickly see all available tasks and their descriptions.
:::

### `bab help`

Display help information about Bab or a specific command.

```bash
# General help
bab help

# Help for a specific command
bab help list
```

**Shortcut:**
```bash
bab --help
bab list --help
```

### `bab completion`

Generate shell completion scripts for various shells.

```bash
bab completion [shell]
```

**Supported shells:**
- `bash`
- `zsh`
- `fish`
- `powershell`

**Examples:**

::: code-group

```bash [Bash]
# Load completion in current session
source <(bab completion bash)

# Add to ~/.bashrc for permanent completion
echo 'source <(bab completion bash)' >> ~/.bashrc
```

```bash [Zsh]
# Load completion in current session
source <(bab completion zsh)

# Add to ~/.zshrc for permanent completion
echo 'source <(bab completion zsh)' >> ~/.zshrc
```

```bash [Fish]
# Load completion in current session
bab completion fish | source

# Add to fish config
bab completion fish > ~/.config/fish/completions/bab.fish
```

```powershell [PowerShell]
# Load completion
bab completion powershell | Out-String | Invoke-Expression
```

:::

## Task Execution

Run any task defined in your Babfile by name:

```bash
bab <task-name>
```

**Examples:**

```bash
# Run a simple task
bab build

# Run a nested task
bab dev:start
bab test:unit

# Run with flags
bab build --verbose
bab deploy --dry-run
```

::: info
If no Babfile is found in the current directory, Bab will show an error message.
:::

::: tip Automatic Dependency Execution
When a task has dependencies defined with `deps`, Bab automatically executes them in the correct order before running the main task.

```bash
# If deploy has deps: [build, test]
bab deploy
# Executes: build → test → deploy
```

Use `--verbose` to see dependency execution details:

```bash
bab deploy --verbose
# Shows: "Executing dependency..." for each dep
```
:::

## Global Flags

These flags work with all commands and task executions:

### `-n, --dry-run`

Preview commands without executing them. Shows what would run without actually running it.

```bash
bab build --dry-run
bab deploy -n
```

**Example output:**
```
INFO  ▶ Running task name=build dry-run=true
DEBUG Command step=[1/2] cmd=npm run lint
DEBUG Command step=[2/2] cmd=npm run build
```

::: tip
Always use `--dry-run` before running potentially destructive tasks like `clean` or `deploy`.
:::

### `-v, --verbose`

Enable verbose output for detailed execution information.

```bash
bab build --verbose
bab test -v
```

**Shows:**
- Task descriptions
- Command execution details
- Debug information
- Step-by-step progress

**Example output:**
```
INFO  ▶ Running task name=build
DEBUG Task description desc=Build the project for production
DEBUG Command step=[1/2] cmd=npm run clean
  Cleaning dist directory...
DEBUG Command step=[2/2] cmd=npm run build:prod
  Building for production...
  Build complete!
INFO  Task completed name=build
```

### `-h, --help`

Display help information.

```bash
# General help
bab --help
bab -h

# Command-specific help
bab list --help
bab completion --help
```

### `--version`

Display the Bab version.

```bash
bab --version
```

**Example output:**
```
bab version 1.0.0
```

## Combining Flags

Flags can be combined for powerful workflows:

```bash
# Preview with verbose output
bab deploy --dry-run --verbose
bab deploy -n -v

# Verbose task execution
bab build --verbose
bab test -v
```

## Exit Codes

Bab uses standard exit codes:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success - Task completed successfully |
| `1` | Error - Task failed, command not found, or execution error |

**Examples:**

```bash
# Check exit code
bab build
echo $?  # 0 if success, 1 if failed

# Use in scripts
if bab test; then
  echo "Tests passed"
  bab deploy
else
  echo "Tests failed"
  exit 1
fi
```

## Environment Variables

Bab respects standard environment variables:

### Shell Execution

Commands are executed using the system shell, which respects your environment:

```bash
# Environment variables are available in tasks
export NODE_ENV=production
bab build  # Will use NODE_ENV=production
```

### Debug Output

For Bab's internal debugging (not the same as `--verbose`):

```bash
# Enable Go debug output (for development)
GODEBUG=http2debug=1 bab build
```

## Common Workflows

### Development

```bash
# List all tasks
bab list

# Start development server with verbose output
bab dev:start --verbose

# Run tests
bab test
```

### CI/CD

```bash
# Preview deployment steps
bab deploy --dry-run

# Run with verbose output for better logs
bab build --verbose
bab test --verbose
bab deploy --verbose
```

### Debugging

```bash
# Check what commands will run
bab deploy --dry-run --verbose

# Run with detailed output
bab build --verbose
```

## Error Messages

Bab provides clear, actionable error messages:

### Task Not Found

```bash
$ bab invalid-task
ERROR Task not found task=invalid-task
INFO  Run 'bab list' to see available tasks
```

**Solution:** Run `bab list` to see available tasks.

### No Babfile Found

```bash
$ bab build
ERROR no Babfile found
```

**Solution:** Create a `Babfile`, `Babfile.yaml`, or `Babfile.yml` in the current directory.

### Command Failed

```bash
$ bab test
INFO  ▶ Running task name=test
  Running tests...
ERROR command failed: exit status 1
```

**Solution:** Check the task's command output. The command returned a non-zero exit code.

### Parse Error

```bash
$ bab build
ERROR Failed to parse Babfile error=yaml: line 5: did not find expected key
```

**Solution:** Check your Babfile for YAML syntax errors at the indicated line.

## Shell Completion

Enable shell completion for a better experience:

### Bash

```bash
# Install completion
echo 'source <(bab completion bash)' >> ~/.bashrc
source ~/.bashrc
```

### Zsh

```bash
# Install completion
echo 'source <(bab completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

### Fish

```bash
# Install completion
bab completion fish > ~/.config/fish/completions/bab.fish
```

After enabling completion, you can:
- Tab-complete commands: `bab l<TAB>` → `bab list`
- Tab-complete tasks: `bab dev:<TAB>` → shows `dev:start`, `dev:watch`, etc.
- Tab-complete flags: `bab --d<TAB>` → `bab --dry-run`

## Tips & Tricks

### Quick Task Listing

```bash
# Quick way to see all tasks
bab list

# Even shorter
bab ls
```

### Safe Execution

```bash
# Always preview destructive commands
bab clean --dry-run
# If it looks good, run it
bab clean
```

### Verbose CI Builds

```bash
# Better logs in CI/CD pipelines
bab build --verbose
bab test --verbose
bab deploy --verbose
```

### Combining with Other Tools

```bash
# Chain commands with &&
bab build && bab test && bab deploy

# Use in scripts
#!/bin/bash
bab build --verbose || exit 1
bab test --verbose || exit 1
bab deploy --verbose
```

## Next Steps

- **[Babfile Syntax](/guide/babfile-syntax)** - Learn how to write Babfiles
- **[Getting Started](/guide/getting-started)** - Quick start guide

## Need Help?

- Join our [Discord community](https://discord.bab.sh)
- Check [GitHub Issues](https://github.com/bab-sh/bab/issues)
- Read the [documentation](https://docs.bab.sh)
