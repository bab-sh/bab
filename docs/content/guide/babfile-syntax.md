# Babfile Syntax

Learn how to write Babfiles to define your project's tasks and workflows.

## File Naming

Bab automatically searches for your task file in the current directory in this order:

1. `Babfile`
2. `Babfile.yaml`
3. `Babfile.yml`

You can use any of these names - Bab will find it automatically.

::: info
All three formats are equivalent. Use whichever you prefer.
:::

## Basic Structure

A Babfile is written in YAML and consists of task definitions. Each task requires:

- A **name** (the YAML key)
- A **run** field with the command(s) to execute
- An optional **desc** field for documentation

### Simple Task

```yaml
build:
  desc: Build the project
  run: go build -o myapp
```

Run with:
```bash
bab build
```

## Task Properties

### `desc` - Description

Provides human-readable documentation for the task. Displayed when listing tasks.

```yaml
test:
  desc: Run all tests with coverage
  run: go test ./... -cover
```

::: tip
Always add descriptions to make your Babfile self-documenting. Your future self (and your team) will thank you.
:::

### `run` - Commands

Defines the command(s) to execute. Can be a single command or multiple commands.

#### Single Command

```yaml
setup:
  desc: Install dependencies
  run: npm install
```

#### Multiple Commands

When you need to run several commands in sequence:

```yaml
deploy:
  desc: Build and deploy to production
  run:
    - npm run build
    - npm test
    - rsync -av dist/ user@server:/var/www/
```

::: warning
Commands execute in order. If any command fails (non-zero exit code), execution stops immediately.
:::

#### Multiline Commands

For complex commands, use YAML's multiline syntax:

```yaml
backup:
  desc: Backup database with timestamp
  run: |
    timestamp=$(date +%Y%m%d_%H%M%S)
    pg_dump mydb > backup_${timestamp}.sql
    echo "Backup created: backup_${timestamp}.sql"
```

## Nested Tasks

Organize related tasks into groups using hierarchical structure. Tasks are accessed with colon notation:

```yaml
dev:
  start:
    desc: Start development server
    run: npm run dev

  watch:
    desc: Watch for file changes
    run: npm run watch

  clean:
    desc: Clean development artifacts
    run: rm -rf .cache dist

test:
  unit:
    desc: Run unit tests
    run: npm run test:unit

  integration:
    desc: Run integration tests
    run: npm run test:integration

  all:
    desc: Run all tests
    run: npm test
```

Run nested tasks:

```bash
bab dev:start
bab dev:clean
bab test:unit
bab test:all
```

Listing shows the hierarchy:

```bash
$ bab list
dev
├── clean Clean development artifacts
├── start Start development server
└── watch Watch for file changes
test
├── all Run all tests
├── integration Run integration tests
└── unit Run unit tests
```

::: tip Naming Convention
Use clear, descriptive names for parent categories:
- `dev:*` for development tasks
- `test:*` for testing tasks
- `build:*` for build tasks
- `deploy:*` for deployment tasks
- `docker:*` for Docker-related tasks
:::

## Task Naming Rules

Task names must follow these rules:

- Use lowercase letters, numbers, hyphens, and underscores
- No spaces or special characters (except `:` for nesting)
- Start with a letter

::: code-group

```yaml [✅ Good]
build:
  desc: Build project
  run: go build

test-unit:
  desc: Run unit tests
  run: npm test

dev_start:
  desc: Start server
  run: npm run dev
```

```yaml [❌ Bad]
Build:  # Uppercase
  run: go build

test unit:  # Space
  run: npm test

@dev:  # Special character
  run: npm run dev
```

:::

## Shell Commands

Commands in the `run` field execute using the system shell:

| Platform | Shell |
|----------|-------|
| macOS/Linux | `sh -c` |
| Windows | `cmd /c` |

This means you can use shell features:

### Pipes and Redirects

```yaml
find-large:
  desc: Find large files
  run: find . -type f -size +10M | sort -h

export:
  desc: Export database
  run: pg_dump mydb > backup.sql
```

### Environment Variables

```yaml
check-env:
  desc: Check Node version
  run: echo "Node version: $NODE_VERSION"

build:
  desc: Build with environment
  run: NODE_ENV=production npm run build
```

### Conditional Execution

```yaml
check-server:
  desc: Check if server is running
  run: curl http://localhost:3000 && echo "Server is up" || echo "Server is down"
```

### Command Substitution

```yaml
backup:
  desc: Create timestamped backup
  run: cp data.db backup_$(date +%Y%m%d).db
```

## Complete Examples

### Node.js Project

```yaml
# Babfile for Node.js project

setup:
  desc: Install dependencies and setup environment
  run:
    - npm install
    - cp .env.example .env

dev:
  start:
    desc: Start development server
    run: npm run dev

  watch:
    desc: Watch and rebuild
    run: npm run watch

test:
  unit:
    desc: Run unit tests
    run: npm run test:unit

  e2e:
    desc: Run E2E tests
    run: npm run test:e2e

  all:
    desc: Run all tests
    run: npm test

build:
  dev:
    desc: Build for development
    run: npm run build:dev

  prod:
    desc: Build for production
    run:
      - npm run lint
      - npm test
      - npm run build:prod

clean:
  desc: Remove build artifacts and dependencies
  run:
    - rm -rf dist
    - rm -rf node_modules
    - rm -rf .cache

lint:
  desc: Lint and format code
  run:
    - npm run eslint
    - npm run prettier --write .
```

### Go Project

```yaml
# Babfile for Go project

setup:
  desc: Download dependencies
  run: go mod download

dev:
  run:
    desc: Run with hot reload
    run: air

  build:
    desc: Build for development
    run: go build -o app

test:
  unit:
    desc: Run unit tests
    run: go test ./...

  coverage:
    desc: Run tests with coverage
    run: go test ./... -coverprofile=coverage.out

  bench:
    desc: Run benchmarks
    run: go test -bench=. ./...

build:
  desc: Build production binary
  run: go build -ldflags="-s -w" -o app

  all:
    desc: Build for all platforms
    run:
      - GOOS=linux GOARCH=amd64 go build -o app-linux
      - GOOS=darwin GOARCH=amd64 go build -o app-darwin
      - GOOS=windows GOARCH=amd64 go build -o app.exe

clean:
  desc: Clean build artifacts
  run: rm -f app app-* coverage.out

lint:
  desc: Run linters
  run:
    - go fmt ./...
    - go vet ./...
    - golangci-lint run
```

### Docker Project

```yaml
# Babfile for Docker project

docker:
  build:
    desc: Build Docker image
    run: docker build -t myapp:latest .

  run:
    desc: Run container
    run: docker run -p 8080:8080 myapp:latest

  stop:
    desc: Stop all containers
    run: docker stop $(docker ps -q)

  clean:
    desc: Remove all containers and images
    run:
      - docker rm $(docker ps -aq)
      - docker rmi $(docker images -q)

  compose:
    up:
      desc: Start all services
      run: docker-compose up -d

    down:
      desc: Stop all services
      run: docker-compose down

    logs:
      desc: View logs
      run: docker-compose logs -f
```

## Best Practices

### 1. Always Add Descriptions

```yaml
# ✅ Good - self-documenting
build:
  desc: Build the application for production
  run: npm run build

# ❌ Less helpful
build:
  run: npm run build
```

### 2. Use Nested Tasks for Organization

```yaml
# ✅ Good - organized
docker:
  build:
    desc: Build Docker image
    run: docker build -t myapp .

  run:
    desc: Run Docker container
    run: docker run myapp

# ❌ Less organized
docker-build:
  desc: Build Docker image
  run: docker build -t myapp .

docker-run:
  desc: Run Docker container
  run: docker run myapp
```

### 3. Break Complex Tasks into Steps

```yaml
# ✅ Good - clear steps
deploy:
  desc: Deploy to production
  run:
    - echo "Building application..."
    - npm run build
    - echo "Running tests..."
    - npm test
    - echo "Deploying..."
    - npm run deploy
    - echo "Deployment complete!"
```

### 4. Keep Commands Cross-Platform When Possible

```yaml
# ✅ More cross-platform
clean:
  desc: Clean build directory
  run: node scripts/clean.js

# ⚠️ Unix-specific
clean:
  desc: Clean build directory
  run: rm -rf dist/
```

### 5. Use Comments for Clarity

```yaml
# Build tasks
build:
  desc: Build for production
  run: npm run build

# Development tasks
dev:
  start:
    desc: Start dev server
    run: npm run dev
```

### 6. Group Related Tasks Logically

```yaml
# Installation and setup
setup:
  desc: Setup project
  run: npm install

# Development tasks
dev:
  # ... dev tasks

# Testing tasks
test:
  # ... test tasks

# Build tasks
build:
  # ... build tasks

# Deployment tasks
deploy:
  # ... deploy tasks
```

## Next Steps

- **[CLI Reference](/guide/cli-reference)** - Learn about all CLI commands and flags
- **[Getting Started](/guide/getting-started)** - Quick start guide
