# Babfile Syntax

Define tasks in YAML format. Bab searches for `Babfile`, `Babfile.yaml`, or `Babfile.yml`.

## Basic Task

```yaml
build:
  desc: Build the application
  run: go build -o app
```

## Task Properties

### `desc` - Description
Optional documentation for the task.

### `run` - Commands
Single command or list of commands to execute.

```yaml
deploy:
  desc: Deploy application
  run:
    - npm test
    - npm run build
    - ./deploy.sh
```

### `deps` - Dependencies
Tasks to run before this task.

```yaml
setup:
  desc: Install dependencies
  run: npm install

build:
  desc: Build application
  deps: setup
  run: npm run build

deploy:
  desc: Deploy to production
  deps: [build, test]
  run: ./deploy.sh
```

## Nested Tasks

Organize tasks with colon notation:

```yaml
dev:
  start:
    desc: Start dev server
    run: npm run dev

  watch:
    desc: Watch files
    run: npm run watch

test:
  unit:
    desc: Unit tests
    run: npm run test:unit

  e2e:
    desc: E2E tests
    run: npm run test:e2e
```

Run with `bab dev:start` or `bab test:unit`.

## Example

```yaml
setup:
  desc: Install dependencies
  run: npm install

lint:
  desc: Run linter
  deps: setup
  run: npm run lint

build:
  desc: Build application
  deps: [setup, lint]
  run: npm run build

test:
  desc: Run tests
  deps: build
  run: npm test

deploy:
  desc: Deploy to production
  deps: [build, test]
  run: ./deploy.sh
```
