# CLI Reference

## Commands

### `bab list`
List all available tasks.

```bash
bab list
```

Aliases: `ls`, `tasks`

### `bab <task>`
Execute a task.

```bash
bab build
bab dev:start
```

Tasks with dependencies run them first automatically.

### `bab completion <shell>`
Generate shell completion script.

```bash
# Bash
source <(bab completion bash)

# Zsh
source <(bab completion zsh)

# Fish
bab completion fish > ~/.config/fish/completions/bab.fish
```

Supported shells: `bash`, `zsh`, `fish`, `powershell`

## Flags

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
bab list --help
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error |
