#!/usr/bin/env bash
# Conventional commit message validator
# Format: type(scope): description
#
# Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

set -e

COMMIT_MSG_FILE="$1"
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

if [[ "$COMMIT_MSG" =~ ^Merge|^fixup!|^squash! ]]; then
    exit 0
fi

CONVENTIONAL_REGEX="^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9_-]+\))?(!)?: .{1,}"

if ! echo "$COMMIT_MSG" | head -1 | grep -qE "$CONVENTIONAL_REGEX"; then
    echo ""
    echo "ERROR: Invalid commit message format"
    echo ""
    echo "Your message: $(echo "$COMMIT_MSG" | head -1)"
    echo ""
    echo "Expected format: type(scope): description"
    echo ""
    echo "Valid types:"
    echo "  feat     - A new feature"
    echo "  fix      - A bug fix"
    echo "  docs     - Documentation only changes"
    echo "  style    - Formatting, whitespace, etc"
    echo "  refactor - Code change that neither fixes a bug nor adds a feature"
    echo "  perf     - Performance improvement"
    echo "  test     - Adding or fixing tests"
    echo "  build    - Changes to build system or dependencies"
    echo "  ci       - Changes to CI configuration"
    echo "  chore    - Other changes that don't modify src or test files"
    echo "  revert   - Reverts a previous commit"
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
