# Architecture

Crego is organized around a simple path:

```text
Cobra command -> recipe model -> component resolver -> generation plan -> templates -> files
```

The TUI and non-interactive CLI share that same path. Interactive choices become a recipe; recipes drive generation.

## Cobra CLI Shell

The CLI is built with Cobra. Command construction is kept injectable and testable: commands accept writers where practical, return errors instead of exiting, and keep command wiring separate from generation logic.

Primary command groups:

```text
crego
  new
  configure
  generate
  recipe
    init
    validate
    print
  components
    list
    show
  explain
  version
```

`new` is the end-to-end entrypoint. `configure` creates recipes. `generate` writes projects from recipes. `explain` prints the resolved plan without mutation.

## Recipe Model

The recipe package owns:

- YAML loading and saving
- snake_case field names
- defaults
- normalization
- validation
- presets

Validation happens before filesystem writes. Recipe validation checks value enums and cross-field compatibility, especially database driver, framework, and migration combinations.

The recipe should contain scaffolding choices, not generated application runtime concerns. For example, Crego stores `task_scheduler: gocron`, while the generated application stores task cron, immediate start, batch size, and retention settings.

## Component Registry And Resolver

The component registry is the catalog of available generation pieces. Each component has:

- ID
- category
- name and description
- required components
- conflicting components
- template files
- Go modules
- hooks

The resolver maps normalized recipe values to component IDs, expands dependencies, checks conflicts, and builds a deterministic plan. The plan contains the resolved component list, files to render, Go modules to add, and hooks to report or run.

## Generator And Templates

The generator takes a validated recipe and resolved plan, then renders embedded templates. It is responsible for:

- rendering target paths
- rendering template contents
- dry-run planning
- safe writes
- overwrite protection
- output directory checks
- path validation

Generation should be idempotent where practical and explainable before mutation.

## TUI Layer

The TUI layer turns interactive choices into the same recipe model used by the CLI. It owns wizard screens, state transitions, previews, save behavior, and generation progress views.

The architectural rule is that TUI screens do not become a separate generation path. They collect input, preview the recipe and plan, and call the same recipe, resolver, and generator packages.

## Application Composition In Generated Projects

Generated web projects compose runtime dependencies in the application package:

- config loader
- logger
- database clients
- HTTP server
- readiness checks
- optional scheduler

Layered layout separates these concerns into packages like `internal/config`, `internal/database`, `internal/logging`, `internal/server`, and `internal/scheduler`. Minimal layout keeps more code in `internal/app` for smaller projects.

The generated layout follows common Go project conventions inspired by `golang-standards/project-layout`, but Crego does not treat that repository as an official standard.
