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

### `alias` / `aliases` - Short Names

Define short aliases for tasks. Run `bab t` instead of `bab test`:

```yaml
tasks:
  test:
    desc: Run tests
    alias: t
    run:
      - cmd: go test ./...

  build:
    desc: Build the application
    aliases: [b, bld]
    run:
      - cmd: go build -o app
```

Use `alias` for a single shortcut or `aliases` for multiple. Both can be combined:

```yaml
tasks:
  deploy:
    alias: d
    aliases: [dep, ship]
    run:
      - cmd: ./deploy.sh
```

Aliases appear in `bab --list` and work with tab completion.

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

<div v-pre>

## Conditional Execution

Skip tasks or run items based on conditions using the `when` key. Conditions support variable interpolation and comparison operators.

### Task-Level Conditions

Skip an entire task if the condition is false:

```yaml
tasks:
  deploy-prod:
    desc: Deploy to production
    when: ${{ environment }} == 'prod'
    run:
      - cmd: ./deploy.sh --env=prod
```

### Run Item Conditions

Skip individual commands, task references, logs, or prompts:

```yaml
tasks:
  build:
    run:
      - prompt: skip_tests
        type: confirm
        message: "Skip tests?"
        default: false
      - cmd: npm test
        when: ${{ skip_tests }} == 'false'
      - cmd: npm run build
```

### Condition Syntax

| Syntax | Description |
|--------|-------------|
| `${{ var }}` | Truthy check - runs if variable is non-empty and not "false" |
| `${{ var }} == 'value'` | Equality - runs if variable equals value |
| `${{ var }} != 'value'` | Inequality - runs if variable does not equal value |

Both single quotes (`'value'`) and double quotes (`"value"`) are supported.

### Truthy Values

Values are evaluated as follows:
- Empty string (`""`) → falsy (skip)
- `"false"` (case-insensitive) → falsy (skip)
- Whitespace-only → falsy (skip)
- Undefined variables → falsy (skip)
- Any other value → truthy (run)

### Using with Prompts

Conditions work with prompt results for dynamic workflows:

```yaml
tasks:
  deploy:
    run:
      - prompt: confirm
        type: confirm
        message: "Deploy to production?"
        default: false
      - cmd: ./deploy.sh
        when: ${{ confirm }}
      - log: "Deployment skipped"
        when: ${{ confirm }} == 'false'
```

### Complete Example

```yaml
tasks:
  release:
    desc: Build and optionally deploy
    run:
      - prompt: environment
        type: select
        message: "Select environment:"
        options: [dev, staging, prod]
        default: dev

      - prompt: run_tests
        type: confirm
        message: "Run tests first?"
        default: true

      - cmd: npm test
        when: ${{ run_tests }}

      - cmd: npm run build

      - cmd: ./deploy.sh --env=dev
        when: ${{ environment }} == 'dev'

      - cmd: ./deploy.sh --env=staging
        when: ${{ environment }} == 'staging'

      - cmd: ./deploy.sh --env=prod
        when: ${{ environment }} == 'prod'

      - log: "Deployed to ${{ environment }}"
```

</div>

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

## Working Directory

The `dir` option sets the working directory for command execution. Relative paths are resolved from the source Babfile's location.

### Global Dir

Set the default working directory for all tasks:

```yaml
dir: ./src

tasks:
  build:
    run:
      - cmd: npm run build  # runs in ./src
```

### Task Dir

Override the global directory for a specific task:

```yaml
dir: ./src

tasks:
  test:
    dir: ./tests
    run:
      - cmd: npm test  # runs in ./tests
```

### Command Dir

Set the directory for a specific command:

```yaml
tasks:
  deploy:
    run:
      - cmd: npm run build
        dir: ./frontend
      - cmd: go build
        dir: ./backend
```

### Dir Precedence

Command-level overrides task-level, which overrides global. The default is the Babfile's directory.

```yaml
dir: ./global

tasks:
  example:
    dir: ./task
    run:
      - cmd: pwd          # runs in ./task
      - cmd: pwd          # runs in ./cmd
        dir: ./cmd
```

### Included Babfiles

Tasks from included Babfiles run in their source Babfile's directory by default. Relative paths in included tasks are resolved from the included Babfile's location, not the main Babfile.

```yaml
# main/Babfile.yml
includes:
  api:
    babfile: ./api/Babfile.yml  # api tasks run in ./api by default

tasks:
  build:
    run:
      - cmd: pwd  # runs in ./main
```

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

## Interactive Prompts

Collect user input during task execution with the `prompt` run item. Prompt results are stored in variables for use in subsequent commands.

### Prompt Types

#### Confirm (Yes/No)

```yaml
tasks:
  deploy:
    run:
      - prompt: continue_deploy
        type: confirm
        message: "Deploy to production?"
        default: false
      - cmd: echo "Deploying..."
```

Result stored: `"true"` or `"false"`

#### Input (Free Text)

```yaml
tasks:
  setup:
    run:
      - prompt: project_name
        type: input
        message: "Enter project name:"
        default: "my-app"
        placeholder: "name"
        validate: "^[a-z-]+$"  # optional regex
      - cmd: echo "Creating ${{ project_name }}"
```

#### Select (Single Choice)

```yaml
tasks:
  configure:
    run:
      - prompt: environment
        type: select
        message: "Select environment:"
        options:
          - dev
          - staging
          - prod
        default: dev
      - cmd: echo "Selected: ${{ environment }}"
```

#### Multiselect (Multiple Choices)

```yaml
tasks:
  install:
    run:
      - prompt: features
        type: multiselect
        message: "Select features to install:"
        options:
          - auth
          - api
          - ui
          - analytics
        defaults: [auth, api]
        min: 1
        max: 3
      - cmd: echo "Installing: ${{ features }}"
```

Result stored as comma-separated string: `"auth,api,ui"`

#### Password (Hidden Input)

```yaml
tasks:
  login:
    run:
      - prompt: api_key
        type: password
        message: "Enter API key:"
        confirm: true  # require re-entry
      - cmd: ./auth.sh
        env:
          API_KEY: ${{ api_key }}
```

#### Number (Numeric Input)

```yaml
tasks:
  scale:
    run:
      - prompt: replicas
        type: number
        message: "Number of replicas:"
        default: 3
        min: 1
        max: 10
      - cmd: kubectl scale --replicas=${{ replicas }}
```

### Prompt Properties

| Property | Type | Description |
|----------|------|-------------|
| `prompt` | string | Variable name to store result (required) |
| `type` | string | Prompt type: confirm, input, select, multiselect, password, number (required) |
| `message` | string | Message to display (required) |
| `default` | string | Default value for confirm/input/select/number |
| `defaults` | array | Default selections for multiselect |
| `options` | array | Available choices for select/multiselect |
| `placeholder` | string | Placeholder text for input |
| `validate` | string | Regex pattern to validate input |
| `min` | integer | Minimum value (number) or selections (multiselect) |
| `max` | integer | Maximum value (number) or selections (multiselect) |
| `confirm` | boolean | Require password re-entry for confirmation |
| `platforms` | array | Run only on specified platforms |
| `when` | string | Condition to evaluate before running |

### Non-Interactive Mode (CI)

When running in non-interactive environments (no TTY), prompts use default values. If no default is set, the task fails with an error. Password prompts always fail in non-interactive mode.

### Complete Example

```yaml
vars:
  app_name: myapp

tasks:
  deploy:
    desc: Interactive deployment
    run:
      - prompt: confirm_deploy
        type: confirm
        message: "Deploy ${{ app_name }} to production?"
        default: false

      - prompt: environment
        type: select
        message: "Select target environment:"
        options: [staging, production]
        default: staging

      - prompt: replica_count
        type: number
        message: "Number of replicas:"
        default: 3
        min: 1
        max: 10

      - log: "Deploying to ${{ environment }} with ${{ replica_count }} replicas"
      - cmd: ./deploy.sh --env=${{ environment }} --replicas=${{ replica_count }}
```

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
