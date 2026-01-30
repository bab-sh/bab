#!/usr/bin/env bash
# Conventional commit message validator
# Aligned with changelog generation in .goreleaser.yaml
#
# Format: type(scope): description
#
# Categories:
#   🚨 Breaking Changes - any type with ! (e.g., feat!:)
#   ✨ Features         - feat
#   🐛 Bug Fixes        - fix
#   ⚡ Performance       - perf
#   ♻️ Refactoring      - refactor
#   📚 Documentation    - docs
#   🧪 Tests            - test
#   🔧 Build & CI       - build, ci
#   📦 Other Changes    - style, revert
#   ❌ Excluded         - chore (ignored in changelog)

set -e

COMMIT_MSG_FILE="$1"
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

if [[ "$COMMIT_MSG" =~ ^Merge|^fixup!|^squash! ]]; then
    exit 0
fi

CONVENTIONAL_REGEX="^(feat|fix|perf|refactor|docs|test|build|ci|style|revert|chore)(\([a-zA-Z0-9_-]+\))?!?: .{1,}"

if ! echo "$COMMIT_MSG" | head -1 | grep -qE "$CONVENTIONAL_REGEX"; then
    echo ""
    echo "ERROR: Invalid commit message format"
    echo ""
    echo "Your message: $(echo "$COMMIT_MSG" | head -1)"
    echo ""
    echo "Expected format: type(scope): description"
    echo ""
    echo "Valid types (aligned with changelog categories):"
    echo "  feat     - ✨ Features"
    echo "  fix      - 🐛 Bug Fixes"
    echo "  perf     - ⚡ Performance"
    echo "  refactor - ♻️ Refactoring"
    echo "  docs     - 📚 Documentation"
    echo "  test     - 🧪 Tests"
    echo "  build    - 🔧 Build & CI"
    echo "  ci       - 🔧 Build & CI"
    echo "  style    - 📦 Other Changes"
    echo "  revert   - 📦 Other Changes"
    echo "  chore    - ❌ Excluded from changelog"
    echo ""
    echo "Examples:"
    echo "  feat: add new parser for YAML files"
    echo "  fix(parser): handle empty input correctly"
    echo "  docs: update README with installation steps"
    echo "  feat!: breaking change to API"
    echo ""
    exit 1
fi

SUBJECT_LINE=$(echo "$COMMIT_MSG" | head -1)
if [ ${#SUBJECT_LINE} -gt 72 ]; then
    echo "WARNING: Commit subject is ${#SUBJECT_LINE} chars (recommended max: 72)"
fi

exit 0
