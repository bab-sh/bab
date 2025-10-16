---
layout: home

hero:
  name: Bab
  text: Custom commands for every project
  tagline: A modern task runner built for seamless development workflows. Simple when you want it, powerful when you need it.
  image:
    src: https://cdn.bab.sh/l/icon
    alt: Bab Logo
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/bab-sh/bab

features:
  - icon: 🚀
    title: Zero Dependencies
    details: No npm, Python, or other runtimes required. Just download and run. Pure Go binary with no external dependencies.

  - icon: 🌍
    title: Cross-Platform
    details: Works seamlessly on Windows, macOS, and Linux. Native shell execution on every platform with consistent behavior.

  - icon: 📝
    title: Simple YAML Configuration
    details: Define your tasks in clean, readable YAML. More intuitive than Makefiles, more powerful than npm scripts.

  - icon: 🎯
    title: Nested Task Support
    details: Organize related tasks into groups using colon notation (dev:start, test:unit). Keep your workflows organized and discoverable.

  - icon: ⚡
    title: Fast & Lightweight
    details: Built with Go for speed and efficiency. Instant startup, minimal overhead. From zero to running your tasks in milliseconds.

  - icon: 🛠️
    title: Developer-Friendly
    details: Dry-run mode to preview commands, verbose output for debugging, beautiful task listing, and intuitive CLI.

  - icon: 📦
    title: Universal Task Runner
    details: Works with any language or project. Whether you're building Node.js, Go, Python, or anything else, Bab has you covered.

  - icon: 🎨
    title: Beautiful CLI
    details: Colorized output, tree-structured task listing, and clear error messages. A CLI that's actually pleasant to use.
---

## Quick Example

Create a `Babfile` in your project root:

```yaml
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

Run your tasks:

::: code-group

```bash [macOS/Linux]
# List all tasks
bab list

# Run a task
bab dev

# Preview without executing
bab build --dry-run
```

```powershell [Windows]
# List all tasks
bab list

# Run a task
bab dev

# Preview without executing
bab build --dry-run
```

:::

## Why Bab?

**Bab** replaces the clunky syntax of Makefiles and the limitations of npm scripts with a universal, dependency-free solution that works across any language or project. Designed with developer experience at its core, Bab scales effortlessly from small scripts to hundreds of tasks.

### Comparison

| Feature | Bab | Make | npm scripts | Task/Taskfile |
|---------|-----|------|-------------|---------------|
| Zero dependencies | ✅ | ✅ | ❌ (requires npm) | ❌ (requires Task) |
| Cross-platform | ✅ | ⚠️ | ✅ | ✅ |
| Simple syntax | ✅ | ❌ | ⚠️ | ✅ |
| Nested tasks | ✅ | ❌ | ⚠️ | ⚠️ |
| Dry-run mode | ✅ | ⚠️ | ❌ | ✅ |
| Any language | ✅ | ✅ | ❌ | ✅ |

## Installation

::: code-group

```bash [macOS/Linux (curl)]
curl -sSfL https://bab.sh/install.sh | sh
```

```bash [Homebrew]
brew tap bab-sh/tap
brew install --cask bab
```

```powershell [Windows]
iwr -useb https://bab.sh/install.ps1 | iex
```

```bash [Go]
go install github.com/bab-sh/bab@latest
```

:::

[See all installation methods →](/guide/installation)

## Connect With Us

Stay connected with the Bab community across all platforms:

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1rem; margin-top: 1.5rem;">

<div>

**🌐 [bab.sh](https://bab.sh)**
Visit our official website

</div>

<div>

**💬 [Discord](https://discord.bab.sh)**
Get help, discuss features, and connect with the community

</div>

<div>

**🐙 [GitHub](https://github.com/bab-sh/bab)**
Star the repo, contribute code, and report issues

</div>

<div>

**𝕏 [X/Twitter](https://x.com/babshdev)**
Follow @babshdev for updates and announcements

</div>

<div>

**📷 [Instagram](https://instagram.com/babshdev)**
Follow @babshdev for visual updates

</div>

<div>

**🤖 [Reddit](https://reddit.com/r/babsh)**
Join the r/babsh community

</div>

<div>

**🧵 [Threads](https://threads.net/@babshdev)**
Follow @babshdev on Threads

</div>

</div>
