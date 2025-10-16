---
layout: home

title: Bab
titleTemplate: Custom commands for every project

hero:
  name: Bab
  text: Custom commands for every project
  tagline: |
    A modern task runner built for seamless development workflows.
    Zero dependencies, cross-platform, simple and powerful.
  image:
    src: https://cdn.bab.sh/l/icon
    alt: Bab Logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: brand
      text: Installation
      link: /guide/installation
    - theme: alt
      text: GitHub
      link: https://github.com/bab-sh/bab
    - theme: alt
      text: Join Discord
      link: https://discord.bab.sh

features:
  - icon: 🚀
    title: Zero Dependencies
    details: Pure Go binary with zero dependencies - just download and run.
    link: /guide/installation
    linkText: Installation Guide

  - icon: 🌍
    title: Cross-Platform
    details: Works seamlessly on Windows, macOS, and Linux with native shell execution.
    link: /guide/getting-started
    linkText: Getting Started

  - icon: 📝
    title: Simple YAML Configuration
    details: Define your tasks in clean, readable YAML - more intuitive than Makefiles.
    link: /guide/babfile-syntax
    linkText: Babfile Syntax

  - icon: 🎯
    title: Nested Task Support
    details: Organize tasks into groups with colon notation like dev:start and test:unit.
    link: /guide/babfile-syntax#nested-tasks
    linkText: Learn About Nested Tasks

  - icon: ⚡
    title: Fast & Lightweight
    details: Built with Go for instant startup and minimal overhead.
    link: /guide/getting-started
    linkText: Quick Start

  - icon: 🛠️
    title: Developer-Friendly
    details: Dry-run mode, verbose output, beautiful task listing, and intuitive CLI.
    link: /guide/cli-reference
    linkText: CLI Reference

  - icon: 📦
    title: Universal Task Runner
    details: Works with any language or project - Node.js, Go, Python, and more.
    link: /guide/babfile-syntax#complete-examples
    linkText: View Examples

  - icon: 🎨
    title: Beautiful CLI
    details: Colorized output, tree-structured task listing, and clear error messages.
    link: /guide/cli-reference
    linkText: Explore CLI Features
---

## Installation

Get started with Bab using your preferred installation method:

::: code-group

```bash [Quick Install]
curl -sSfL https://bab.sh/install.sh | sh
```

```bash [Homebrew Cask]
brew install --cask bab-sh/tap/bab
```

```bash [Homebrew]
brew install bab-sh/tap/bab
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

```powershell [Chocolatey]
choco install bab
```

```bash [Scoop]
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket
scoop install bab
```

```bash [Go]
go install github.com/bab-sh/bab@latest
```

:::

[See all installation methods →](/guide/installation)
