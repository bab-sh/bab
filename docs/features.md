# Features

Bab comes packed with features to make task management simple and efficient.

## Task Execution

Run any defined task by name:

```bash
bab <task-name>
```

Example:

```yaml
# Babfile
build:
  desc: Build the project
  run: go build -o myapp
```

```bash
bab build
```

### Sequential Command Execution

Tasks with multiple commands execute in sequence. If any command fails, execution stops:

```yaml
deploy:
  desc: Deploy to production
  run:
    - npm run build      # Runs first
    - npm test           # Runs if build succeeds
    - npm run deploy     # Runs if tests pass
```

## Task Listing

Run `bab` without arguments to see all available tasks:

```bash
bab
```

Output:

```
Available tasks:
  setup         Setup development environment
  dev:start     Start development server
  dev:watch     Watch for file changes
  test          Run all tests
  build         Build for production
```

Tasks are automatically organized and formatted for readability.

## Nested and Grouped Tasks

Organize related tasks using colon notation:

```yaml
docker:
  build:
    desc: Build Docker image
    run: docker build -t myapp .

  run:
    desc: Run container
    run: docker run -p 8080:8080 myapp

  stop:
    desc: Stop container
    run: docker stop myapp
```

Execute with:

```bash
bab docker:build
bab docker:run
bab docker:stop
```

Benefits:
- **Organization**: Group related tasks together
- **Discoverability**: Clear task hierarchy
- **Namespace**: Avoid naming conflicts

## Dry Run Mode

Preview commands without executing them:

```bash
bab <task> --dry-run
# or
bab <task> -n
```

Example:

```bash
$ bab deploy --dry-run
INFO  ▶ Running task name=deploy
DEBUG Command step=[1/3] cmd=npm run build
DEBUG Command step=[2/3] cmd=npm test
DEBUG Command step=[3/3] cmd=npm run deploy:prod
```

Perfect for:
- Verifying task commands before execution
- Debugging complex workflows
- Understanding what a task does

## Verbose Output

Get detailed execution information:

```bash
bab <task> --verbose
# or
bab <task> -v
```

Shows:
- Task descriptions
- Command execution details
- Debug information
- Step-by-step progress

Example:

```bash
$ bab build --verbose
INFO  ▶ Running task name=build
DEBUG Task description desc=Build the project for production
DEBUG Command step=[1/2] cmd=npm run clean
  Cleaning dist directory...
DEBUG Command step=[2/2] cmd=npm run build:prod
  Building for production...
  Build complete!
INFO  Task completed name=build
```

## Custom Babfile Path

Use a different file instead of the default `Babfile`:

```bash
bab --file path/to/custom.yaml <task>
# or
bab -f custom-tasks.yml <task>
```

Useful for:
- Multiple task files in one project
- Environment-specific tasks
- Shared task libraries

Example:

```bash
# Use production-specific tasks
bab --file tasks/production.yml deploy

# Use development tasks
bab --file tasks/dev.yml start
```

## Script Compilation

Compile your Babfile to standalone shell scripts:

```bash
bab compile
```

Generates:
- `bab.sh` - Unix/Linux/macOS script
- `bab.bat` - Windows batch file

Options:

```bash
# Compile to custom directory
bab compile -o dist/

# Disable colors in generated scripts
bab compile --no-color
```

Benefits:
- **Zero Dependencies**: No bab installation required
- **Distribution**: Share tasks with your team
- **CI/CD**: Use in build pipelines
- **Portability**: Works anywhere shell/batch runs

See the [Compile Guide](/compile) for more details.

## Cross-Platform Support

Bab works seamlessly across platforms:

| Platform | Shell | Support |
|----------|-------|---------|
| macOS | sh | ✅ Full |
| Linux | sh | ✅ Full |
| Windows | cmd | ✅ Full |

Commands are executed using the appropriate shell for each platform:
- **Unix/Linux/macOS**: `sh -c "command"`
- **Windows**: `cmd /c "command"`

## Version Information

Check your bab version:

```bash
bab --version
```

## Help and Documentation

Access help information:

```bash
# General help
bab --help

# Command-specific help
bab compile --help
```

## Command-Line Flags

### Global Flags

Available for all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--file <path>` | `-f` | Path to Babfile |
| `--dry-run` | `-n` | Show commands without executing |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help information |
| `--version` |  | Show version information |

### Compile Command Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output <dir>` | `-o` | Output directory for scripts |
| `--no-color` |  | Disable colors in scripts |

## Error Handling

Bab provides clear error messages:

### Task Not Found

```bash
$ bab invalid-task
ERROR Task not found task=invalid-task
INFO  Run 'bab' to see available tasks
```

### No Babfile Found

```bash
$ bab build
ERROR no Babfile found
```

### Command Failure

```bash
$ bab test
INFO  ▶ Running task name=test
  Running tests...
ERROR command failed: exit status 1
```

## Output Streaming

Command output is streamed in real-time:

```yaml
watch:
  desc: Watch for changes
  run: npm run watch
```

```bash
$ bab watch
INFO  ▶ Running task name=watch
  Watching for file changes...
  File changed: src/index.js
  Rebuilding...
  Build complete!
```

Standard output and errors are properly handled:
- **stdout**: Displayed normally
- **stderr**: Shown as errors

## Interactive Commands

Bab supports interactive commands that require user input:

```yaml
init:
  desc: Initialize new project
  run: npm init
```

```bash
$ bab init
INFO  ▶ Running task name=init
  package name: (my-project)
  version: (1.0.0)
  # ... interactive prompts work normally
```

## Best Practices

### Use Dry Run for Testing

Before running potentially dangerous commands:

```bash
bab cleanup --dry-run  # Check what will be deleted
bab cleanup            # Actually delete
```

### Combine Flags

Flags can be combined for powerful debugging:

```bash
bab deploy --verbose --dry-run
```

### Use Verbose in CI/CD

Enable verbose output in automated environments:

```bash
bab build --verbose  # Better logs in CI
```

## Next Steps

- Learn about [Script Compilation](/compile) in detail
- Check the [Syntax Guide](/syntax) for advanced patterns
- See [Getting Started](/get-started) for quick setup
