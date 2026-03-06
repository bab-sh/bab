<div align="center">

<img src="https://raw.githubusercontent.com/bab-sh/bab/main/assets/svg/icon-rounded.svg" alt="Bab Logo" width="170"/>

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
# Global variables — supports references and env access
vars:
  app_name: myapp
  version: "1.0.0"
  base_dir: ./src
  build_dir: ${{ base_dir }}/build
  home: ${{ env.HOME }}

# Global environment variables — passed to all commands
env:
  APP_NAME: ${{ app_name }}
  NODE_ENV: production

# Global defaults
silent: false
output: true
dir: ./project

# Import tasks from other Babfiles (namespaced)
includes:
  utils:
    babfile: ./tools/Babfile.yml

tasks:
  setup:
    desc: Install dependencies
    silent: true
    output: false
    run:
      - cmd: npm install

  lint:
    desc: Run linter
    dir: ./frontend
    run:
      - cmd: npm run lint

  test:unit:
    desc: Run unit tests
    alias: t
    deps: [setup]
    env:
      CI: "true"
    run:
      - cmd: npm test

  test:all:
    desc: Run all checks in parallel
    run:
      - parallel:
          - task: lint
          - task: test:unit
        mode: interleaved

  ci:
    desc: Full CI pipeline
    run:
      - parallel:
          - task: lint
            label: lint
          - task: test:unit
            label: tests
          - cmd: npm run build
            label: build
        mode: grouped
        limit: 2

  configure:
    desc: Interactive project configuration
    run:
      # Confirm prompt
      - prompt: proceed
        type: confirm
        message: "Configure the project?"
        default: true

      # Input prompt with validation
      - prompt: project_name
        type: input
        message: "Project name:"
        default: myapp
        placeholder: "my-project"
        validate: "^[a-z][a-z0-9-]+$"
        when: ${{ proceed }}

      # Select prompt
      - prompt: environment
        type: select
        message: "Select environment:"
        options: [dev, staging, prod]
        default: dev
        when: ${{ proceed }}

      # Multiselect prompt with min/max
      - prompt: features
        type: multiselect
        message: "Select features:"
        options: [auth, api, ui, analytics]
        defaults: [auth, api]
        min: 1
        max: 3
        when: ${{ proceed }}

      # Number prompt with range
      - prompt: replicas
        type: number
        message: "Number of replicas:"
        default: 3
        min: 1
        max: 10
        when: ${{ proceed }}

      # Password prompt with confirmation
      - prompt: api_key
        type: password
        message: "Enter API key:"
        confirm: true
        when: ${{ proceed }}
        platforms: [linux, darwin]

      - log: "Configured ${{ project_name }} for ${{ environment }}"
        level: info
        when: ${{ proceed }}

  build:
    desc: Build ${{ app_name }} v${{ version }}
    alias: b
    aliases: [bld, compile]
    deps: [lint, test:unit]
    vars:
      output_path: ${{ build_dir }}/${{ app_name }}
    run:
      - log: Building ${{ app_name }}...
        level: debug
      - cmd: npm run build
        dir: ./frontend
        env:
          BUILD_OUTPUT: ${{ output_path }}
        silent: true
        output: true
      - cmd: cp -r dist ${{ output_path }}
        platforms: [linux, darwin]
      - cmd: xcopy dist ${{ output_path }} /E
        platforms: [windows]
      - log: Build complete!
        level: info

  deploy:
    desc: Deploy to ${{ env.DEPLOY_ENV }}
    when: ${{ environment }} != 'dev'
    deps: [build]
    run:
      - log: Deploying ${{ app_name }} to ${{ env.DEPLOY_ENV }}...
        level: warn
      - cmd: ./scripts/deploy.sh
        env:
          VERSION: ${{ version }}
        when: ${{ environment }} == 'staging'
      - cmd: ./scripts/deploy-prod.sh
        env:
          VERSION: ${{ version }}
        when: ${{ environment }} == 'prod'
      - task: utils:notify
        when: ${{ environment }} == 'prod'
      - log: "Use $${{ var }} for literal syntax"
        level: debug
```

Run your tasks:
```bash
bab                  # Browse tasks interactively
bab --list           # List all available tasks
bab build            # Build the application
bab b                # Same as above (using alias)
bab bld              # Also works (multiple aliases)
bab t                # Run tests (using alias)
bab utils:setup      # Run included task
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