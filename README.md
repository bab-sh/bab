# Bab - Custom Commands for your Projects

**Zero dependencies. Cross-platform by design.**

Write custom commands in YAML, execute them with bab. Just like Makefile but with more features and Zero dependencies. (soon)

ITS PRE PRE ALPHA DO NOT USE IT FRFR

### TODOS
- [x] Last commands executed history
- [ ] Basic babfile schema
- [ ] Advanced nesting and multi Babfile support
- [ ] Platform flags for commands so you can define commands specifically for OS
- [ ] Binary Compile (so its not just shell and batch scripts)
- [ ] Babfile configs and environment variable handeling
- [ ] Homepage
- [ ] Being a good markdown/taskfile replacement

### IDEAS
- [ ] Performance profiling
- [x] Interactive mode (fuzzy search or something)
- [ ] Modules (multi babfile support in module format)
- [ ] Generate babfiles/generate and add new commands to babfile via forms


## Installation

### Homebrew Cask (macOS)

```bash
brew tap bab-sh/tap
```
```bash
brew install --cask bab
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/bab-sh/bab.git
cd bab

# Build bab
go build -o bab

# Move to your PATH (optional)
sudo mv bab /usr/local/bin/
```

### 2. Create a Babfile

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

### 3. Run your tasks

```bash
bab                    # List all tasks
bab setup              # Run setup task
bab dev:start          # Run nested task
```

### 4. (Optional) Compile to standalone scripts

```bash
bab compile            # Generate bab.sh and bab.bat

# Now distribute zero-dependency scripts
./bab.sh setup         # Unix/Linux/macOS
bab.bat setup          # Windows
```

## 📋 Commands

```bash
# Basic usage
bab                    # List all tasks
bab <task>             # Run a task

# Example task usage
bab setup              # Run setup task
bab dev:start          # Run nested task

# Options
bab <task> --dry-run   # Show what would run
bab <task> --verbose   # Verbose output
bab --file custom.yaml <task>  # Use custom Babfile

# Compile to standalone scripts
bab compile            # Generate bab.sh and bab.bat
bab compile -o dist    # Output to custom directory
bab compile --no-color # Disable colors in scripts
```

## 📜 License

MIT License - Free for personal and commercial use.

---

**Made with ❤️ by aio for developers who value simplicity and reliability.**
