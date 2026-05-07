# Recipes

Recipes are Crego's deterministic generation contract. A recipe describes the project metadata and selected components; Crego validates it, resolves dependencies, renders templates, and writes the generated project.

Recipe files are YAML and use snake_case field names.

## Purpose

Use recipes when you want generation to be repeatable across local development, CI, and code review. A recipe should answer "what should be generated" without hiding network calls, filesystem mutations, or runtime application settings.

The normal workflow is:

```sh
crego configure --recipe crego.yaml
crego recipe validate crego.yaml
crego explain --recipe crego.yaml
crego generate --recipe crego.yaml
```

## Schema Overview

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
  azure_pipelines: true
```

## Presets

Presets are starter recipes that can be edited before generation:

| Preset | Purpose |
| --- | --- |
| `web-basic` | Basic web service |
| `web-postgres` | Web service with PostgreSQL |
| `web-mysql` | Web service with MySQL |
| `web-sqlite` | Web service with SQLite |
| `web-redis` | Web service with Redis |
| `web-mongodb` | Web service with MongoDB |
| `cli-basic` | Basic CLI application |

Create a preset recipe:

```sh
crego recipe init --preset web-postgres --module github.com/acme/orders-web --out crego.yaml
```

## Validation

Validation runs before component resolution and before filesystem writes. It checks:

- required project fields
- safe project names and Go module-like paths
- enum values such as `project.type`, `layout.style`, `server.framework`, and `task_scheduler`
- database driver, framework, and migration compatibility
- logging and configuration formats

Examples:

```sh
crego recipe validate crego.yaml
crego explain --recipe crego.yaml
```

`explain` is useful for CI and review workflows because it prints the resolved plan without writing files.

## Component Fields

| Field | Values |
| --- | --- |
| `project.type` | `web`, `cli` |
| `layout.style` | `minimal`, `layered` |
| `server.framework` | `nethttp`, `chi`, `gin`, `echo`, `fiber` |
| `configuration.format` | `env`, `yaml`, `json`, `toml` |
| `logging.framework` | `slog`, `zap`, `zerolog`, `logrus` |
| `database.driver` | `none`, `postgres`, `mysql`, `sqlite`, `redis`, `mongodb` |
| `database.framework` | `pgx`, `sql`, `gorm` |
| `database.migrations` | `none`, `goose`, `migrate` |
| `task_scheduler` | `none`, `gocron` |
| `ci.github_actions` | `true`, `false` |
| `ci.gitlab_ci` | `true`, `false` |
| `ci.azure_pipelines` | `true`, `false` |

Compatibility rules:

- `pgx` is PostgreSQL-only.
- `sql` and `gorm` apply to SQL databases: PostgreSQL, MySQL, SQLite.
- `goose` and `migrate` apply to SQL databases.
- Redis and MongoDB do not use SQL frameworks or migrations.
- `task_scheduler: gocron` is generated for web projects.

## Full Web Example

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
  azure_pipelines: true
```
