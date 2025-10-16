# Contributing

We welcome contributions from the community! Whether it's bug reports, feature requests, documentation improvements, or code contributions, we appreciate your help in making Bab better.

## Ways to Contribute

### Report Bugs

Found a bug? Help us fix it:

1. **Search existing issues** on [GitHub](https://github.com/bab-sh/bab/issues) to avoid duplicates
2. **Create a new issue** with:
   - Clear, descriptive title
   - Steps to reproduce the bug
   - Expected vs actual behavior
   - Your environment (OS, Bab version, Go version)
   - Relevant Babfile content (if applicable)

**Example bug report:**

```markdown
**Bug:** Task with multiple commands stops on first failure

**Steps to reproduce:**
1. Create Babfile with task containing multiple commands
2. Make second command fail
3. Run the task

**Expected:** All commands should run
**Actual:** Execution stops after first command fails

**Environment:**
- OS: macOS 14.0
- Bab version: v0.1.0
- Go version: 1.25.1
```

### Request Features

Have an idea for a new feature?

1. **Check the [roadmap](/reference/roadmap)** to see if it's already planned
2. **Search existing issues** to avoid duplicates
3. **Create a new issue** with:
   - Clear description of the feature
   - Use cases and examples
   - Why it would be valuable
   - (Optional) Implementation ideas

### Improve Documentation

Documentation improvements are always welcome:

- Fix typos or unclear explanations
- Add examples or use cases
- Improve code snippets
- Translate documentation

To contribute to docs:

1. Fork the repository
2. Edit files in `docs/content/`
3. Submit a pull request

### Contribute Code

Ready to write some code?

## Development Setup

### Prerequisites

- **Go 1.25 or later** - [Download here](https://golang.org/dl/)
- **Git** - [Download here](https://git-scm.com/)
- **Make** (optional) - For using the Makefile

### Setup Steps

1. **Fork and clone the repository:**

   ```bash
   # Fork on GitHub first, then:
   git clone https://github.com/YOUR-USERNAME/bab.git
   cd bab
   ```

2. **Download dependencies:**

   ```bash
   go mod download
   ```

3. **Build the project:**

   ```bash
   go build -o bab
   ```

4. **Run the binary:**

   ```bash
   ./bab --version
   ```

5. **Run tests:**

   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Create a Branch

Create a feature branch for your changes:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

**Branch naming conventions:**
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Your Changes

Write clean, well-documented code:

- Follow Go best practices and idioms
- Add comments for complex logic
- Keep functions focused and small
- Use descriptive variable names

### 3. Test Your Changes

Ensure all tests pass:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests verbosely
go test ./... -v
```

Add new tests for your changes:

```bash
# Example test file structure
internal/parser/parser_test.go
```

### 4. Format Your Code

Format code before committing:

```bash
# Format all files
go fmt ./...

# Run linter
go vet ./...

# (Optional) Use golangci-lint
golangci-lint run
```

### 5. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add feature: task dependencies support"
```

**Good commit messages:**
```
✅ Add dry-run mode for task execution
✅ Fix: Handle missing Babfile gracefully
✅ Docs: Add examples for nested tasks
✅ Refactor: Simplify task parser logic
```

**Less helpful commit messages:**
```
❌ Update code
❌ Fix bug
❌ Changes
```

### 6. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name
```

Then create a pull request on GitHub:

1. Go to the [Bab repository](https://github.com/bab-sh/bab)
2. Click "New Pull Request"
3. Select your branch
4. Fill in the PR template
5. Submit!

## Pull Request Guidelines

### PR Title

Use clear, descriptive titles:

- `feat: Add task dependency support`
- `fix: Handle missing Babfile gracefully`
- `docs: Add installation guide for Windows`
- `refactor: Simplify task parser`

### PR Description

Include in your PR:

- **What** - What does this PR do?
- **Why** - Why is this change needed?
- **How** - How does it work?
- **Testing** - How did you test it?
- **Related Issues** - Link to related issues

**Example:**

```markdown
## What
Adds support for task dependencies using the `deps` field.

## Why
Users need to run prerequisite tasks automatically (e.g., build before deploy).

## How
- Added `deps` field to task struct
- Modified executor to run dependencies first
- Added validation for circular dependencies

## Testing
- Added unit tests for dependency resolution
- Tested with sample Babfiles
- Verified circular dependency detection

## Related Issues
Closes #42
```

### Code Review

Once submitted:

1. **CI checks** will run automatically
2. **Maintainers** will review your code
3. **Address feedback** by pushing new commits
4. **Merge** once approved

::: tip
Be patient! Reviews may take a few days. Feel free to ping after a week if no response.
:::

## Coding Standards

### Go Style Guide

Follow the [official Go style guide](https://go.dev/doc/effective_go):

- Use `gofmt` for formatting
- Follow naming conventions (camelCase for private, PascalCase for public)
- Write descriptive comments for exported functions
- Keep functions small and focused

### Error Handling

Handle errors properly:

```go
// ✅ Good
if err != nil {
    log.Error("Failed to parse Babfile", "error", err)
    return fmt.Errorf("failed to parse Babfile: %w", err)
}

// ❌ Avoid
if err != nil {
    panic(err)  // Don't panic in library code
}
```

### Testing

Write tests for new functionality:

```go
func TestTaskParser(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Task
        wantErr bool
    }{
        {
            name:  "simple task",
            input: "build:\n  desc: Build\n  run: go build",
            want:  Task{Description: "Build", Commands: []string{"go build"}},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseTask(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseTask() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

Document exported functions:

```go
// ParseBabfile parses a Babfile and returns a map of tasks.
// It supports YAML format and validates task structure.
//
// Example:
//   tasks, err := ParseBabfile("Babfile")
//   if err != nil {
//       return err
//   }
func ParseBabfile(path string) (map[string]Task, error) {
    // Implementation...
}
```

## Project Structure

Understanding the codebase:

```
bab/
├── cmd/              # CLI commands (root, list, etc.)
├── internal/         # Internal packages
│   ├── executor/     # Task execution logic
│   ├── finder/       # Babfile discovery
│   ├── parser/       # YAML parsing and validation
│   └── version/      # Version information
├── docs/             # Documentation (VitePress)
├── scripts/          # Build and release scripts
├── main.go           # Entry point
├── go.mod            # Go module definition
└── README.md         # Project readme
```

## Release Process

Maintainers handle releases, but here's how it works:

1. **Version bump** in `internal/version/version.go`
2. **Update CHANGELOG.md** with changes
3. **Create git tag** (e.g., `v1.0.0`)
4. **Push tag** to trigger automated release
5. **GoReleaser** builds binaries for all platforms
6. **GitHub release** created automatically

## Community

### Discord

Join our [Discord community](https://discord.bab.sh) for:
- Questions and support
- Feature discussions
- Development chat
- Community showcase

### Code of Conduct

We follow a code of conduct to ensure a welcoming environment:

- **Be respectful** - Treat everyone with respect
- **Be constructive** - Provide helpful feedback
- **Be inclusive** - Welcome newcomers
- **No harassment** - Zero tolerance for harassment or discrimination

Report violations to the maintainers.

## Recognition

Contributors are recognized in:
- **README.md** - Listed in contributors section
- **Release notes** - Mentioned in relevant releases
- **Discord** - Special contributor role

## Getting Help

Need help contributing?

- **Discord** - Ask in the #development channel on [Discord](https://discord.bab.sh)
- **GitHub Discussions** - Post in [Discussions](https://github.com/bab-sh/bab/discussions)
- **Issues** - Comment on relevant issues

## Next Steps

Ready to contribute?

1. **Find an issue** - Look for ["good first issue"](https://github.com/bab-sh/bab/labels/good%20first%20issue) labels
2. **Set up your environment** - Follow the development setup above
3. **Make your changes** - Write code and tests
4. **Submit a PR** - Follow the PR guidelines
5. **Celebrate** - You're now a Bab contributor!

Thank you for contributing to Bab! 🎉
