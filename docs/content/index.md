---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Bab"
  text: "A modern task runner"
  tagline: from simple to scaled.
  actions:
    - theme: brand
      text: Get Started
      link: /get-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/bab-sh/bab

features:
  - title: Zero Dependencies
    details: No external dependencies required. Just download and run. When compiled to scripts, zero runtime dependencies needed.

  - title: Cross-Platform
    details: Works seamlessly on Windows, macOS, and Linux. Compile to platform-specific scripts (bab.sh for Unix, bab.bat for Windows).

  - title: Simple YAML Configuration
    details: Define your tasks in clean, readable YAML. Just like Makefile, but more intuitive and feature-rich.

  - title: Nested Task Support
    details: Organize related tasks into groups using colon notation (dev:start, test:unit). Keep your workflows organized and discoverable.

  - title: Script Compilation
    details: Compile your Babfile to standalone shell scripts. Distribute zero-dependency scripts to your team without requiring bab installation.

  - title: Developer-Friendly
    details: Dry-run mode to preview commands, verbose output for debugging, custom Babfile paths, and automatic task listing.
---

