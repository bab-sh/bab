# Babfile Syntax

Define tasks in YAML format. Bab searches for `Babfile`, `Babfile.yaml`, or `Babfile.yml`.

## Basic Task

```yaml
tasks:
  build:
    desc: Build the application
    run:
      - cmd: go build -o app
```

## Task Properties

### `desc` - Description
Optional documentation for the task.

### `run` - Commands
List of commands to execute. Each command uses the `cmd` key.

```yaml
tasks:
  deploy:
    desc: Deploy application
    run:
      - cmd: npm test
      - cmd: npm run build
      - cmd: ./deploy.sh
```

### `deps` - Dependencies
Tasks to run before this task.

```yaml
tasks:
  setup:
    desc: Install dependencies
    run:
      - cmd: npm install

  build:
    desc: Build application
    deps: [setup]
    run:
      - cmd: npm run build

  deploy:
    desc: Deploy to production
    deps: [build, test]
    run:
      - cmd: ./deploy.sh
```

## Namespaced Tasks

Use colon notation for task namespaces (flat structure, not nested YAML):

```yaml
tasks:
  dev:start:
    desc: Start dev server
    run:
      - cmd: npm run dev

  dev:watch:
    desc: Watch files
    run:
      - cmd: npm run watch

  test:unit:
    desc: Unit tests
    run:
      - cmd: npm run test:unit

  test:e2e:
    desc: E2E tests
    run:
      - cmd: npm run test:e2e
```

Run with `bab dev:start` or `bab test:unit`.

## Platform-Specific Commands

Run different commands based on the operating system using the `platforms` array:

```yaml
tasks:
  deploy:
    desc: Deploy to production
    run:
      - cmd: ./scripts/deploy.sh
        platforms: [linux, darwin]
      - cmd: powershell scripts/deploy.ps1
        platforms: [windows]
```

Available platforms: `linux`, `darwin`, `windows`. Commands without a `platforms` array run on all platforms.

## Includes

Import tasks from other Babfiles with namespace prefixes:

```yaml
includes:
  utils:
    babfile: ./tools/Babfile.yml

tasks:
  build:
    desc: Build everything
    deps: [utils:setup]
    run:
      - cmd: go build -o app
```

Tasks from the included file are prefixed with the namespace (e.g., `utils:setup`, `utils:lint`).

## Complete Example

```yaml
tasks:
  setup:
    desc: Install dependencies
    run:
      - cmd: npm install

  lint:
    desc: Run linter
    deps: [setup]
    run:
      - cmd: npm run lint

  build:
    desc: Build application
    deps: [setup, lint]
    run:
      - cmd: npm run build

  test:
    desc: Run tests
    deps: [build]
    run:
      - cmd: npm test

  deploy:
    desc: Deploy to production
    deps: [build, test]
    run:
      - cmd: ./deploy.sh
        platforms: [linux, darwin]
      - cmd: powershell deploy.ps1
        platforms: [windows]
```
