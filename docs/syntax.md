# Babfile Syntax

Learn how to write Babfiles to define your project's tasks and workflows.

## File Naming

Bab automatically searches for your task file in the following order:

1. `Babfile`
2. `Babfile.yaml`
3. `Babfile.yml`

You can also specify a custom file path:

```bash
bab --file custom.yaml <task>
```

## Basic Structure

A Babfile is written in YAML and consists of task definitions. Each task must have:

- A **name** (the key)
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

Provides documentation for the task. Shown when listing tasks with `bab`:

```yaml
test:
  desc: Run all tests
  run: go test ./...
```

### `run` - Commands

Defines the command(s) to execute. Can be a single string or a list of commands.

#### Single Command

```yaml
setup:
  desc: Install dependencies
  run: npm install
```

#### Multiple Commands

When you need to run multiple commands in sequence:

```yaml
deploy:
  desc: Build and deploy to production
  run:
    - npm run build
    - npm run test
    - rsync -av dist/ user@server:/var/www/
```

Commands execute in order. If any command fails, execution stops.

## Nested Tasks

Organize related tasks into groups using the colon (`:`) notation:

```yaml
dev:
  start:
    desc: Start development server
    run: npm run dev

  watch:
    desc: Watch for file changes
    run: npm run watch

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

Run nested tasks with:

```bash
bab dev:start
bab test:unit
bab test:all
```

List tasks shows the hierarchy:

```
Available tasks:
  dev:start         Start development server
  dev:watch         Watch for file changes
  test:unit         Run unit tests
  test:integration  Run integration tests
  test:all          Run all tests
```

## Complete Example

Here's a comprehensive Babfile showcasing different features:

```yaml
# Babfile - Project automation

# Setup and initialization
setup:
  desc: Setup development environment
  run:
    - npm install
    - cp .env.example .env
    - npm run db:migrate

# Development tasks
dev:
  start:
    desc: Start development server
    run: npm run dev

  clean:
    desc: Clean build artifacts
    run: rm -rf dist/ .cache/

  reset:
    desc: Reset development environment
    run:
      - npm run dev:clean
      - npm run setup

# Testing
test:
  unit:
    desc: Run unit tests
    run: npm run test:unit

  integration:
    desc: Run integration tests
    run: npm run test:integration

  e2e:
    desc: Run end-to-end tests
    run: npm run test:e2e

  all:
    desc: Run all tests
    run:
      - npm run test:unit
      - npm run test:integration
      - npm run test:e2e

# Building
build:
  dev:
    desc: Build for development
    run: npm run build:dev

  prod:
    desc: Build for production
    run:
      - npm run test
      - npm run build:prod
      - npm run build:verify

# Deployment
deploy:
  staging:
    desc: Deploy to staging
    run:
      - npm run build:prod
      - npm run deploy:staging

  production:
    desc: Deploy to production
    run:
      - npm run build:prod
      - npm run test:all
      - npm run deploy:production

# Maintenance
clean:
  desc: Remove all generated files
  run:
    - rm -rf dist/
    - rm -rf node_modules/
    - rm -rf .cache/

format:
  desc: Format code
  run: npm run prettier --write .

lint:
  desc: Lint code
  run: npm run eslint .
```

## Shell Commands

Commands in the `run` field are executed using the system shell:

- **Unix/Linux/macOS**: Commands run with `sh -c`
- **Windows**: Commands run with `cmd /c`

This means you can use shell features like:

```yaml
check:
  desc: Check if server is running
  run: curl http://localhost:3000 && echo "Server is up"

backup:
  desc: Backup database
  run: pg_dump mydb > backup_$(date +%Y%m%d).sql
```

## Best Practices

### 1. Always Add Descriptions

Make your Babfile self-documenting:

```yaml
# Good
build:
  desc: Build the application for production
  run: npm run build

# Less helpful
build:
  run: npm run build
```

### 2. Use Nested Tasks for Organization

Group related tasks together:

```yaml
# Good - organized
docker:
  build:
    desc: Build Docker image
    run: docker build -t myapp .

  run:
    desc: Run Docker container
    run: docker run -p 8080:8080 myapp

# Less organized - flat structure
docker-build:
  desc: Build Docker image
  run: docker build -t myapp .

docker-run:
  desc: Run Docker container
  run: docker run -p 8080:8080 myapp
```

### 3. Break Complex Tasks into Steps

Use multiple commands for clarity:

```yaml
deploy:
  desc: Deploy to production
  run:
    - echo "Building application..."
    - npm run build
    - echo "Running tests..."
    - npm test
    - echo "Deploying..."
    - npm run deploy:prod
    - echo "Deployment complete!"
```

### 4. Keep Commands Cross-Platform When Possible

Prefer Node.js scripts or cross-platform tools over shell-specific commands:

```yaml
# More cross-platform
clean:
  desc: Clean build directory
  run: node scripts/clean.js

# Less cross-platform (Unix-specific)
clean:
  desc: Clean build directory
  run: rm -rf dist/
```

## Next Steps

- Learn about all available [Features](/features)
- See how to [Compile to Scripts](/compile)
- Check out the [Getting Started](/get-started) guide
