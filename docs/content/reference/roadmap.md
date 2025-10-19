# Roadmap

Track Bab's development progress and upcoming features.

::: warning Pre-Alpha
Bab is in **pre-alpha** (v0.x.x). The API and features may change between releases.
:::

## What's Available Now

Current features in the latest release:

### Core Features
- [x] Task execution from Babfile
- [x] Nested tasks with colon notation (`dev:start`)
- [x] Multi-command tasks
- [x] Cross-platform support (macOS, Linux, Windows)
- [x] Multiple Babfile formats (Babfile, Babfile.yaml, Babfile.yml)

### CLI Features
- [x] Task listing (`bab list`)
- [x] Dry-run mode (`--dry-run`)
- [x] Verbose output (`--verbose`)
- [x] Shell completion (bash, zsh, fish, powershell)
- [x] Graceful shutdown handling

### Developer Experience
- [x] Colorized CLI output
- [x] Tree-structured task display
- [x] Clear error messages
- [x] Task descriptions
- [x] Task dependencies

## What's Coming Next

High-priority features for upcoming releases:

::: info Priority Features
The most requested features by the community.
:::

### Custom File Paths
Specify alternative Babfile locations with `--file` flag.

### Interactive Mode
Fuzzy search interface for browsing and selecting tasks.

### Platform-Specific Tasks
Define tasks that run only on specific operating systems.

## Future Plans

Additional features being considered:

### Task Management
- [ ] Task history tracking
- [ ] Performance profiling
- [ ] Task watchers (re-run on file changes)
- [ ] Parallel task execution

### Configuration
- [ ] Babfile schema validation
- [ ] Multi-Babfile imports
- [ ] Environment variable management
- [ ] Module system for reusable tasks

### Distribution
- [ ] Script compilation (standalone shell scripts)
- [ ] Template system
- [ ] Plugin architecture

### Advanced
- [ ] Remote task execution
- [ ] TUI/GUI interfaces
- [ ] CI/CD integrations

## Get Involved

Help shape Bab's future:

- **Vote** - Star features on [GitHub Issues](https://github.com/bab-sh/bab/issues)
- **Suggest** - Open an issue with your idea
- **Discuss** - Join [Discord](https://discord.bab.sh) to chat about the roadmap
- **Build** - Contribute code via [pull requests](/contributing)

---

**Latest Release**: See [GitHub Releases](https://github.com/bab-sh/bab/releases) for detailed version history.
