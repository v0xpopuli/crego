# crego

`crego` is a TUI-first Go project generator inspired by Spring Initializr and Powerlevel10k.

It is designed to be **interactive by default**, **deterministic by recipe**, and **scriptable for CI**.

## Features

- **TUI Flow**: Interactive project setup with sensible defaults (Coming Soon).
- **Deterministic Recipes**: Define your project once, generate it many times with predictable results.
- **Component Registry**: Pick and choose components (server, database, logger, etc.) and let `crego` handle the wiring.
- **Dry-run & Explain**: See what will be generated before touching the disk.
- **Validation**: Strict recipe validation ensures compatibility and reduces errors.

## Installation

```bash
go install github.com/v0xpopuli/crego/cmd/crego@latest
```

Or build from source:

```bash
git clone https://github.com/v0xpopuli/crego.git
cd crego
make build
# Binary will be in ./build/app/crego
```

## Quick Start

### 1. Initialize a recipe

Create a starter recipe for a web project with PostgreSQL:

```bash
crego recipe init --preset web-postgres --module github.com/acme/orders --out crego.yaml
```

### 2. Validate the recipe

Ensure your recipe is correct:

```bash
crego recipe validate crego.yaml
```

### 3. Preview generation

See what files and components would be generated:

```bash
crego explain --recipe crego.yaml
```

### 4. Generate the project

Write the generated files to the target directory:

```bash
crego generate --recipe crego.yaml --out ./orders-api
```

## Core Concepts

### Recipes

Recipes are the deterministic contract for project generation. They are YAML files that describe the project metadata, layout, and selected components.

### Components

Components are the building blocks of your project. They can be servers, database drivers, loggers, or custom hooks. `crego` resolves dependencies between components and renders the necessary code.

## Command Reference

- `crego new`: Interactive project setup (Coming Soon).
- `crego generate`: Generate a project from a recipe.
- `crego explain`: Print a generation plan for a recipe without writing files.
- `crego recipe init`: Create a starter recipe file.
- `crego recipe validate`: Validate a recipe file.
- `crego components list`: List available components.
- `crego components show`: Show details for a specific component.

## License

MIT