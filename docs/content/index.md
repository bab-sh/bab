---
layout: home

title: Bab
titleTemplate: Custom commands for every project

hero:
  name: Bab
  text: Custom commands for every project liam du kek
  tagline: Modern task runner. Zero dependencies, cross-platform, simple.
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
    details: Pure Go binary - just download and run.

  - icon: 🌍
    title: Cross-Platform
    details: Works on Windows, macOS, and Linux.

  - icon: 📝
    title: Simple YAML
    details: Define tasks in clean, readable YAML.

  - icon: 🔗
    title: Task Dependencies
    details: Automatic prerequisite execution with deps field.
---

## Installation

### Quick Install

Get started instantly with platform-specific install scripts:

::: code-group

```bash [macOS/Linux]
curl -sSfL https://bab.sh/install.sh | sh
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

:::

### Package Managers

Install using your preferred package manager:

::: code-group

```bash [Homebrew Cask]
brew tap bab-sh/tap
brew install --cask bab
```

```bash [Homebrew]
brew tap bab-sh/tap
brew install bab
```

```powershell [Chocolatey]
choco install bab
```

```bash [Scoop]
scoop bucket add bab-sh https://github.com/bab-sh/scoop-bucket
scoop install bab
```

```bash [Snap]
snap install bab-sh
```

```bash [yay]
yay -S bab-bin
```

```bash [paru]
paru -S bab-bin
```

```bash [Go]
go install github.com/bab-sh/bab@latest
```

:::

[See all installation methods →](/guide/installation)
