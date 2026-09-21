# Engineering Platform

Turn a product idea into a ready-to-develop project foundation.

Describe your idea in Pi; it selects the stack, composes the architecture, and generates the project for you.

## Install

Remote installer (Linux/macOS, installs or updates to the latest release,
sha256-verified against the release `checksums.txt`):

```bash
curl -fsSL https://github.com/JhonMA82/engineering-platform/releases/latest/download/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://github.com/JhonMA82/engineering-platform/releases/latest/download/install.ps1 | iex
```

Keep it updated and remove it:

```bash
eng self-update --check   # what would change, writes nothing
eng self-update --yes     # download, verify, atomically replace
curl -fsSL https://github.com/JhonMA82/engineering-platform/releases/latest/download/uninstall.sh | sh
```

Go users:

```bash
go install github.com/jhonma82/engineering-platform/cmd/eng@latest
```

Prebuilt binaries (`eng-linux-amd64`, `eng-linux-arm64`, `eng-darwin-arm64`, `eng-windows-amd64.exe`, plus `checksums.txt`) ride every `v*` tag on the [releases page](https://github.com/JhonMA82/engineering-platform/releases).

## Quickstart

```bash
mkdir demo
cd demo
eng init
opencode
```

Then inside OpenCode:

```text
/newproject <describe your idea>
```

```text
idea
→ OpenCode
→ Engineering Platform
→ selected/composed foundation
→ generated project ready for development
```

Full walkthrough: [First project](docs/getting-started/first-project.md).

## Commands

| Command | Description |
| --- | --- |
| `eng init` | Prepare the current directory as a workspace |
| `eng resolve` | Resolve an intent to an architecture decision |
| `eng plan` | Print the deterministic materialization plan |
| `eng materialize` | Execute a plan into a new or empty directory |
| `eng start` | Chain resolve, plan, materialize, doctor |
| `eng doctor` | Re-validate a materialized project |
| `eng surface add` | Add one required surface (delta only) |
| `eng extend` | Promote a `planned_later` surface to now |
| `eng add` | Record a product requirement |
| `eng update` | Report pin drift against the catalog |
| `eng explain` | Print why a saved decision resolved as it did |
| `eng catalog` | List, show, and validate the catalog |
| `eng version` | Print core, catalog, commit, and build date |

Flags and exit behavior: [CLI reference](docs/reference/cli.md).

## What you get

```text
my-project/
├── AGENTS.md
├── ARCHITECTURE.md
├── GENTLE.md
├── apps/
├── services/
└── .engineering/
```

Scaffolding plus agent context: root and per-surface `AGENTS.md`, `GENTLE.md` takeover instructions, an implementation brief, and `.engineering/` machine state for `eng doctor` and later evolution. Architecture stays locked; product features are built from here.

## Curated foundations

| Recipe | What it covers |
| --- | --- |
| GP-01 | Public website |
| GP-02 | Authenticated backoffice |
| GP-03 | Python and data app |
| GP-04 | Native mobile client plus shared API |
| GP-05 | Installed desktop tool, local database |
| GP-06 | Multi-app (web plus admin plus API plus PWA) |
| GP-07 | Commercial SaaS |

Every foundation ships with a known origin and an immutable pin, never whatever trended this week. If no curated foundation fits, the resolver returns a `CatalogGap` instead of forcing a wrong answer.

## License

No license file is declared in this repository yet: check with the maintainers before reuse beyond evaluation.

---

- Docs: [index](docs/index.md) · [first project](docs/getting-started/first-project.md) · [evolve](docs/guides/evolve-project.md) · [concepts](docs/concepts/catalog.md) · [architecture](docs/architecture/core.md)
- Maintain: [curate](docs/maintainers/curate-boilerplate.md) · [release](docs/maintainers/release.md) · [contribute](CONTRIBUTING.md)
