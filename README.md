<div align="center">

<img src="https://cdn.bab.sh/l/favicon" alt="Bab Logo" width="170"/>

<h1>Bab</h1>

Custom commands for every project

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bab-sh/bab)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/bab-sh/bab)](https://github.com/bab-sh/bab/releases)
[![Status](https://img.shields.io/badge/Status-Pre--Alpha-red.svg)](https://github.com/bab-sh/bab)

[Website](https://bab.sh) • [Documentation](https://docs.bab.sh) • [Discord](https://discord.bab.sh)

</div>

---

## Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Roadmap](#roadmap)
- [Support](#support)
- [License](#license)

---

## Overview

**Bab** is a modern task runner built for seamless development workflows. It replaces the clunky syntax of Makefiles and the limitations of npm scripts with a universal, dependency-free solution that works across any language or project. Whether you're running small scripts or orchestrating hundreds of tasks, Bab scales effortlessly, keeping your workflow smooth, organized, and maintainable.

---

## Quick Start

### Install Bab

```bash
# macOS / Linux
curl -sSfL https://bab.sh/install.sh | sh

# Windows (PowerShell)
iwr -useb https://bab.sh/install.ps1 | iex
```

For more installation options, see the [Installation Documentation](https://docs.bab.sh/guide/installation).

### Create a Babfile

Create a `Babfile` in your project root:

```yaml
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

For complete Babfile syntax and advanced features, see the [Babfile Documentation](https://docs.bab.sh/guide/babfile-syntax).

### Run Your Tasks

```bash
bab                  # Browse tasks interactively
bab list             # List all available tasks
bab setup            # Run the setup task
bab dev              # Start development server
bab build --dry-run  # Preview build task without executing
```

For more CLI commands and options, see the [CLI Guide](https://docs.bab.sh/guide/cli-reference).

---

## Roadmap

Bab is under active development. Some completed features may be refined or reimplemented as the project matures.

- [x] **Basic Task Execution** - Execute tasks defined in Babfile with shell commands
- [x] **Nested Task Support** - Organize tasks hierarchically with colon notation (e.g., `dev:start`)
- [x] **Cross-Platform Execution** - Native shell execution on Linux, macOS, and Windows
- [x] **Flexible Babfile Formats** - Support for Babfile, Babfile.yaml, and Babfile.yml
- [x] **Dry-Run Mode** - Preview commands without executing them (`--dry-run`)
- [x] **Verbose Logging** - Detailed execution logs for debugging (`--verbose`)
- [x] **Task Listing** - Display all available tasks in a colorized tree structure (`bab list`)
- [x] **Graceful Shutdown** - Proper handling of interrupts and termination signals
- [x] **Multi-Command Tasks** - Support for tasks with single or multiple commands
- [x] **Task Dependencies** - Automatic execution of prerequisite tasks before running a task
- [x] **Interactive Mode** - Fuzzy search interface for browsing and selecting tasks
- [ ] **Task History Tracking** - Per-project execution history with status and duration
- [ ] **Custom File Paths** - Specify alternative Babfile locations with `--file` flag
- [ ] **Babfile Schema Validation** - Structured schema for validating Babfile syntax and configuration
- [ ] **Advanced Multi-Babfile Support** - Import and compose multiple Babfiles for complex project structures
- [ ] **Platform-Specific Tasks** - Define tasks that run only on specific operating systems (Windows, Linux, macOS)
- [ ] **Compiled Binary Execution** - Support for executing compiled binaries beyond shell scripts
- [ ] **Environment Variable Management** - Built-in handling, interpolation, and validation of environment variables
- [ ] **Module System** - Share and reuse Babfiles across projects as importable modules
- [ ] **Performance Profiling** - Built-in performance monitoring and profiling for task execution
- [ ] **Interactive Babfile Generator** - Create and modify Babfiles through interactive command-line forms
- [ ] **Enhanced Output Formatting** - Rich, context-aware colored output with better visual hierarchy

See our [GitHub Issues](https://github.com/bab-sh/bab/issues) for the complete list of planned features and to suggest new ideas.

---

## Support

- 💬 **Discord** - Join our [Discord community](https://discord.bab.sh) for questions and discussions
- 📚 **Documentation** - Visit [docs.bab.sh](https://docs.bab.sh) for comprehensive guides
- 🐛 **Bug Reports** - [GitHub Issues](https://github.com/bab-sh/bab/issues) for bug reports and feature requests
- 🤝 **Contributing** - See the [Contributing Guide](https://docs.bab.sh/contributing) to get started
- 🌐 **Website** - Visit [bab.sh](https://bab.sh) for more information

---

## License

Bab is released under the [MIT License](LICENSE). This means you're free to use, modify, and distribute this software for personal or commercial purposes.

---

<div align="center">

Built with ❤️ by AIO

</div>
