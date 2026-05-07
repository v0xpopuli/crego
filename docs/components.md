# Components

Crego components are the building blocks selected by recipes and resolved into a generation plan. A component can contribute template files, Go modules, dependencies, conflicts, or hooks.

Component IDs use `category.value` form, such as `server.chi`, `database.postgres`, or `task_scheduler.gocron`.

## Full Matrix

| Recipe category | Component IDs |
| --- | --- |
| `project` | `project.web`, `project.cli` |
| `layout` | `layout.minimal`, `layout.layered` |
| `server` | `server.nethttp`, `server.chi`, `server.gin`, `server.echo`, `server.fiber` |
| `configuration` | `configuration.env`, `configuration.yaml`, `configuration.json`, `configuration.toml` |
| `logging` | `logging.slog`, `logging.zap`, `logging.zerolog`, `logging.logrus` |
| `database` | `database.none`, `database.postgres`, `database.mysql`, `database.sqlite`, `database.redis`, `database.mongodb` |
| `database.framework` | `database.framework.pgx`, `database.framework.sql`, `database.framework.gorm` |
| `migrations` | `migrations.none`, `migrations.goose`, `migrations.migrate` |
| `task_scheduler` | `task_scheduler.none`, `task_scheduler.gocron` |
| `observability` | `observability.health`, `observability.readiness` |
| `deployment` | `deployment.docker`, `deployment.compose` |
| `ci` | `ci.github_actions`, `ci.gitlab_ci`, `ci.azure_pipelines` |

Run:

```sh
crego components list
crego components show task_scheduler.gocron
```

## Dependencies

Dependencies are resolved before files are rendered. Examples:

- `server.chi`, `server.gin`, `server.echo`, and `server.fiber` require `project.web`.
- `deployment.compose` requires `deployment.docker`.
- `task_scheduler.gocron` requires `project.web`.
- `database.framework.pgx` requires PostgreSQL.
- migration components require a SQL database.

Dependency expansion is deterministic. A recipe selects high-level values; the resolver expands the required components and de-duplicates files and modules.

## Conflicts

Conflicts prevent incompatible choices from being generated together:

- server frameworks are mutually exclusive
- configuration formats are mutually exclusive
- logging providers are mutually exclusive
- `database.none` conflicts with real database drivers
- `migrations.none`, `migrations.goose`, and `migrations.migrate` are mutually exclusive
- `task_scheduler.none` conflicts with `task_scheduler.gocron`

## Database Compatibility

| Driver | Frameworks | Migrations |
| --- | --- | --- |
| `none` | none | none |
| `postgres` | `pgx`, `sql`, `gorm` | `none`, `goose`, `migrate` |
| `mysql` | `sql`, `gorm` | `none`, `goose`, `migrate` |
| `sqlite` | `sql`, `gorm` | `none`, `goose`, `migrate` |
| `redis` | none | none |
| `mongodb` | none | none |
