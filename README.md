<div align="center">

<img src="https://cdn.bab.sh/l/favicon" alt="Bab Logo" width="170"/>

<h1>Bab</h1>

Custom commands for every project

[![Go Version](https://img.shields.io/github/go-mod/go-version/bab-sh/bab)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/bab-sh/bab)](https://github.com/bab-sh/bab/releases)
[![Status](https://img.shields.io/badge/Status-Pre--Alpha-red.svg)](https://github.com/bab-sh/bab)
[![Discord](https://img.shields.io/discord/1320389080407609344?label=Discord&color=%235865F2)](https://discord.bab.sh)

[Website](https://bab.sh) • [Documentation](https://docs.bab.sh) • [Discord](https://discord.bab.sh)

</div>

---

**Bab** is a modern task runner that replaces the clunky syntax of Makefiles and the limitations of npm scripts with a universal, dependency-free solution that works across any language or project.

## Quick Start
```bash
# macOS / Linux
curl -sSfL https://bab.sh/install.sh | sh

# Windows (PowerShell)
iwr -useb https://bab.sh/install.ps1 | iex
```
For more installation options, see the [Installation Documentation](https://docs.bab.sh/guide/installation).

Create a `Babfile.yml` in your project root:
```yaml
tasks:
  setup:
    desc: Install dependencies
    run:
      - cmd: npm install

  dev:
    desc: Start development server
    deps: [setup]
    run:
      - cmd: npm run dev

  test:
    desc: Run test suite
    run:
      - cmd: npm test

  deploy:
    desc: Deploy to production
    run:
      - cmd: ./scripts/deploy.sh
        platforms: [linux, darwin]
      - cmd: powershell scripts/deploy.ps1
        platforms: [windows]
```

Run your tasks:
```bash
bab                  # Browse tasks interactively
bab --list           # List all available tasks
bab dev              # Start development server
```

## Support

- 💬 [Discord](https://discord.bab.sh) - Questions and discussions
- 📚 [Documentation](https://docs.bab.sh) - Comprehensive guides
- 🐛 [Issues](https://github.com/bab-sh/bab/issues) - Bug reports and feature requests

## Acknowledgments

Bab stands on the shoulders of giants. Special thanks to:

- [Task](https://taskfile.dev) - The modern task runner that inspired Bab's approach
- [Charm](https://charm.sh) - For their beautiful terminal UI libraries that make Bab a joy to use

---

<div align="center">

Built with ❤️ by AIO

</div>