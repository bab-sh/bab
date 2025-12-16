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
List of commands or task references to execute.

#### Shell Commands
Use the `cmd` key to run shell commands:

```yaml
tasks:
  deploy:
    desc: Deploy application
    run:
      - cmd: npm test
      - cmd: npm run build
      - cmd: ./deploy.sh
```

#### Task References
Use the `task` key to run another task inline:

```yaml
tasks:
  setup:
    run:
      - cmd: npm install

  build:
    run:
      - task: setup
      - cmd: npm run build
```

Task references support the same options as commands (`silent`, `output`, etc.).

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

## Silent Mode

The `silent` option suppresses command prompt display (e.g., `$ echo hello`). Useful for reducing noise when running many commands.

### Global Silent

Apply to all tasks:

```yaml
silent: true

tasks:
  build:
    run:
      - cmd: npm run build
```

### Task Silent

Apply to a specific task:

```yaml
tasks:
  install:
    silent: true
    run:
      - cmd: npm install
```

### Command Silent

Apply to individual commands or task references:

```yaml
tasks:
  build:
    run:
      - cmd: echo "Installing..."
        silent: true
      - cmd: npm install
      - task: test
        silent: true
```

### Silent Precedence

Command-level overrides task-level, which overrides global. Default is `false` (show prompts).

## Output Control

The `output` option controls whether stdout/stderr from commands is displayed. Different from `silent` which only affects the command prompt line.

### Global Output

Apply to all tasks:

```yaml
output: false

tasks:
  install:
    run:
      - cmd: npm install
```

### Task Output

Apply to a specific task:

```yaml
tasks:
  install:
    output: false
    run:
      - cmd: npm install
```

### Command Output

Apply to individual commands or task references:

```yaml
tasks:
  build:
    run:
      - cmd: npm install
        output: false
      - cmd: npm run build
```

### Output Precedence

Command-level overrides task-level, which overrides global. Default is `true` (show output).

## Environment Variables

Define environment variables at three levels: global, task, or command. Variables cascade with lower levels overriding higher ones.

### Global Environment

Set variables for all tasks at the root level:

```yaml
env:
  NODE_ENV: production
  API_URL: https://api.example.com

tasks:
  build:
    run:
      - cmd: echo "Building for $NODE_ENV"
```

### Task Environment

Set variables for a specific task:

```yaml
tasks:
  dev:
    desc: Start development server
    env:
      PORT: "3000"
      DEBUG: "true"
    run:
      - cmd: npm run dev
```

### Command Environment

Set variables for a specific command:

```yaml
tasks:
  deploy:
    run:
      - cmd: ./deploy.sh
        env:
          DEPLOY_ENV: staging
      - cmd: ./notify.sh
        env:
          DEPLOY_ENV: production
```

### Precedence

When the same variable is defined at multiple levels, command-level overrides task-level, which overrides global:

```yaml
env:
  MODE: global

tasks:
  example:
    env:
      MODE: task
    run:
      - cmd: echo $MODE  # prints "command"
        env:
          MODE: command
```

<div v-pre>

## Variables

Define reusable values with `${{ }}` syntax. Variables are resolved by Bab before commands run.

### Global Variables

```yaml
vars:
  app_name: myapp
  version: "1.0.0"

tasks:
  build:
    run:
      - cmd: go build -o ${{ app_name }}
```

### Task Variables

Override global variables within a task:

```yaml
vars:
  mode: production

tasks:
  dev:
    vars:
      mode: development
    run:
      - cmd: echo "Running in ${{ mode }} mode"
```

### Environment Access

Read OS environment variables with `${{ env.VAR }}`:

```yaml
vars:
  home: ${{ env.HOME }}
  target: ${{ env.GOOS }}

tasks:
  build:
    run:
      - cmd: echo "Building for ${{ target }}"
```

### Variable References

Variables can reference other variables:

```yaml
vars:
  base: /app
  build_dir: ${{ base }}/build
  output: ${{ build_dir }}/bin

tasks:
  build:
    run:
      - cmd: mkdir -p ${{ output }}
```

### Export to Shell

Variables are not auto-exported. Use `env:` to pass to shell:

```yaml
vars:
  app: myapp

tasks:
  run:
    env:
      APP_NAME: ${{ app }}
    run:
      - cmd: echo $APP_NAME
```

### Escaping

Use `$${{` to output literal `${{`:

```yaml
tasks:
  help:
    run:
      - cmd: echo "Use $${{ var }} syntax"
```

</div>

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
env:
  NODE_ENV: production

tasks:
  setup:
    desc: Install dependencies
    silent: true
    output: false
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
    env:
      BUILD_MODE: release
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
        env:
          DEPLOY_ENV: production
      - cmd: powershell deploy.ps1
        platforms: [windows]
        env:
          DEPLOY_ENV: production
```
