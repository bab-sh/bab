# CLI Reference

## Commands

### `bab`
Browse and select tasks using an interactive fuzzy finder.

```bash
bab
```

Type to search for tasks, use arrow keys or `Ctrl+P`/`Ctrl+N` to navigate, and press `Enter` to execute. Press `Esc` or `Ctrl+C` to exit without running a task.

**Keyboard shortcuts:**
- `Enter` - Execute selected task
- `Up/Down` or `Ctrl+P/Ctrl+N` - Navigate through tasks
- `Ctrl+U` - Clear search input
- `Esc` or `Ctrl+C` - Exit without executing

### `bab <task>`
Execute a task.

```bash
bab build
bab dev:start
```

Tasks with dependencies run them first automatically.

## Flags

### `-l, --list`
List all available tasks.

```bash
bab --list
bab -l
```

### `-c, --completion <shell>`
Generate shell completion script.

```bash
# Bash
source <(bab --completion bash)

# Zsh
source <(bab --completion zsh)

# Fish
bab --completion fish > ~/.config/fish/completions/bab.fish
```

Supported shells: `bash`, `zsh`, `fish`, `powershell`

### `-n, --dry-run`
Preview commands without executing.

```bash
bab build --dry-run
```

### `-v, --verbose`
Show detailed execution logs.

```bash
bab build --verbose
```

### `--version`
Show version information.

```bash
bab --version
```

### `-h, --help`
Show help.

```bash
bab --help
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error |
