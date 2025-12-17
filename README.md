<div align="center">

<img src="https://raw.githubusercontent.com/bab-sh/bab/main/assets/icon-256.png" alt="Bab Logo" width="170"/>

<h1>Bab</h1>

Clean commands for any project.

[![Go Version](https://img.shields.io/github/go-mod/go-version/bab-sh/bab)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/bab-sh/bab)](https://github.com/bab-sh/bab/releases)
[![Status](https://img.shields.io/badge/Status-Alpha-red.svg)](https://github.com/bab-sh/bab)
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
vars:
  app_name: myapp
  version: "1.0.0"
  build_dir: ./build

env:
  APP_NAME: ${{ app_name }}
  NODE_ENV: production

tasks:
  setup:
    desc: Install dependencies
    run:
      - cmd: npm install

  lint:
    desc: Run linter
    run:
      - cmd: npm run lint

  test:unit:
    desc: Run unit tests
    deps: [setup]
    run:
      - cmd: npm test

  test:all:
    desc: Run all tests
    run:
      - task: test:unit
      - log: All tests passed!
        level: info

  build:
    desc: Build ${{ app_name }} v${{ version }}
    deps: [lint, test:unit]
    vars:
      output: ${{ build_dir }}/${{ app_name }}
    run:
      - log: Building ${{ app_name }}...
      - cmd: npm run build
      - cmd: cp -r dist ${{ output }}
        platforms: [linux, darwin]
      - cmd: xcopy dist ${{ output }} /E
        platforms: [windows]
      - log: Build complete!
        level: info

  deploy:
    desc: Deploy to ${{ env.DEPLOY_ENV }}
    deps: [build]
    run:
      - log: Deploying ${{ app_name }} to ${{ env.DEPLOY_ENV }}...
        level: warn
      - cmd: ./scripts/deploy.sh
        env:
          VERSION: ${{ version }}
```

Run your tasks:
```bash
bab                  # Browse tasks interactively
bab --list           # List all available tasks
bab build            # Build the application
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