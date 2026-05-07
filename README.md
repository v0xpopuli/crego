# crego

**Generate Go services you would actually deploy.**

`crego` is a TUI-first Go project generator inspired by Spring Initializr and Powerlevel10k. It helps you create Go web services and CLI projects through an interactive terminal flow, deterministic recipe files, or scriptable commands.

Crego-generated projects use common Go project layout patterns inspired by the widely used [`golang-standards/project-layout`](https://github.com/golang-standards/project-layout) repository. That repository is not an official Go standard; Crego treats it as a practical source of familiar conventions.

## Features

- TUI-first project setup with recipe preview.
- Recipe-first generation with strict YAML validation.
- Component registry for project type, layout, HTTP server, config, logging, database, migrations, scheduler, observability, deployment, and CI.
- `explain` and `--dry-run` flows for reviewing generated output before writing files.
- Safe generation defaults: no silent overwrites, validation before writes, injectable command output, and no `os.Exit` inside internal packages.
- gocron-backed task scheduler generation for web services.

## Installation From Source

```sh
git clone https://github.com/v0xpopuli/crego.git
cd crego
make build
./build/app/crego version
```

For local development without installing:

```sh
go run ./cmd/crego version
go run ./cmd/crego components list
```

## Basic Usage

```sh
crego new
crego new github.com/acme/orders-web
crego configure --recipe web-postgres.yaml
crego generate --recipe crego.yaml
crego recipe validate crego.yaml
crego explain --recipe crego.yaml
crego components list
```

## TUI Usage

Run the default wizard:

```sh
crego new
```

Or create a recipe interactively without generating immediately:

```sh
crego configure --recipe web-postgres.yaml
```

The TUI collects project choices, shows the normalized recipe, resolves the generation plan, and can either save the recipe or generate the project.

## Recipe Usage

Recipes are the deterministic contract for generation. They use YAML with snake_case field names.

```sh
crego recipe init --preset web-postgres --module github.com/acme/orders-web --out crego.yaml
crego recipe validate crego.yaml
crego explain --recipe crego.yaml
crego generate --recipe crego.yaml --out ./orders-web
```

### Example `crego.yaml`

This is the public recipe shape for a layered Chi service with PostgreSQL, pgx, goose migrations, gocron, Docker, Compose, and both GitHub Actions and GitLab CI.

```yaml
version: v1

project:
  name: orders-web
  module: github.com/acme/orders-web
  type: web

go:
  version: "1.25"

layout:
  style: layered

server:
  framework: chi
  port: 8080
  graceful_shutdown: true

configuration:
  format: yaml

logging:
  framework: zap
  format: json
  request_logging: true

database:
  driver: postgres
  framework: pgx
  migrations: goose

task_scheduler: gocron

observability:
  health: true
  readiness: true

deployment:
  docker: true
  compose: true

ci:
  github_actions: true
  gitlab_ci: true
```

## Non-Interactive Usage

Generate from a recipe without prompts:

```sh
crego generate --recipe crego.yaml --out ./orders-web --non-interactive
```

Or generate directly from flags:

```sh
crego new github.com/acme/orders-web \
  --type web \
  --layout layered \
  --server chi \
  --configuration yaml \
  --logging zap \
  --database postgres \
  --framework pgx \
  --migrations goose \
  --docker \
  --compose \
  --github-actions \
  --gitlab-ci \
  --health \
  --readiness \
  --non-interactive
```

## Generated Project Structure

A layered web service is generated with a familiar Go layout:

```text
orders-web/
├── cmd/orders-web/main.go
├── configs/config.yaml
├── deployments/Dockerfile
├── deployments/docker-compose.yml
├── internal/app/app.go
├── internal/config/config.go
├── internal/database/postgres.go
├── internal/database/migrations.go
├── internal/logging/logger.go
├── internal/scheduler/scheduler.go
├── internal/scheduler/tasks/example_cleanup.go
├── internal/server/server.go
├── internal/server/routes.go
├── internal/server/handler/health.go
├── internal/server/handler/ready.go
├── scripts/migrations/000001_init.sql
├── .github/workflows/test.yml
├── .gitlab-ci.yml
├── go.mod
├── Makefile
└── README.md
```

## Supported Component Matrix

| Category | Values |
| --- | --- |
| `project` | `web`, `cli` |
| `layout` | `minimal`, `layered` |
| `server` | `nethttp`, `chi`, `gin`, `echo`, `fiber` |
| `configuration` | `env`, `yaml`, `json`, `toml` |
| `logging` | `slog`, `zap`, `zerolog`, `logrus` |
| `database` | `none`, `postgres`, `mysql`, `sqlite`, `redis`, `mongodb` |
| `framework` | `pgx`, `sql`, `gorm` |
| `migrations` | `none`, `goose`, `migrate` |
| `task_scheduler` | `none`, `gocron` |
| `observability` | `health`, `readiness` |
| `deployment` | `docker`, `compose` |
| `ci` | `github-actions`, `gitlab-ci` |

## Development Commands

```sh
make build
make tests
go run ./cmd/crego components list
go run ./cmd/crego recipe validate examples/crego.yaml
```

Per project policy, agents do not run tests automatically. Developers should run the commands locally and provide failures when fixes are needed.

## License

MIT. See [LICENSE](LICENSE).
