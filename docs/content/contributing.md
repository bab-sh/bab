# Contributing

Contributions are welcome! Here's how to get started.

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## Development Setup

```bash
# Clone your fork
git clone https://github.com/bab-sh/bab.git
cd bab

# Install dependencies
go mod download

# Build
go build -o bab

# Run
./bab list
```

## Testing

```bash
# Run tests
go test ./...

# Run linter
golangci-lint run
```

## Guidelines

- Write clear commit messages
- Add tests for new features
- Update documentation as needed
- Follow Go best practices

## Need Help?

- [Discord](https://discord.bab.sh) - Ask questions
- [GitHub Issues](https://github.com/bab-sh/bab/issues) - Report bugs or request features
