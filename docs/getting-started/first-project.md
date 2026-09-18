# First project

Create your first project with Engineering Platform: install the tool,
prepare a workspace, describe your idea to Pi, and hand the generated
project to Gentle AI (or your team) for feature work.

## 1. Install `eng`

Go users:

```bash
go install github.com/jhonma82/engineering-platform/cmd/eng@latest
```

Or download a prebuilt binary from the
[releases page](https://github.com/JhonMA82/engineering-platform/releases)
(`eng-linux-amd64`, `eng-linux-arm64`, `eng-darwin-arm64`,
`eng-windows-amd64.exe`, plus `checksums.txt`) and place it on your
`PATH`. Then:

```bash
eng version
```

## 2. Create and prepare a directory

```bash
mkdir my-project
cd my-project
eng init
```

`eng init` installs the project-local agent integration (Pi by default;
`--agent opencode` for OpenCode) plus `.engineering/` state. It writes
only inside the current directory and is safe to re-run (reports
"already initialized" when there is nothing to do).

## 3. Describe your idea in Pi

```bash
pi
```

Then:

```text
/newproject Inventory system for warehouses with mobile scanning
```

Pi asks what it needs (problem, users, flows, surfaces, data,
offline/native needs, restrictions) and turns your answers into a
`ProjectIntent`. It resolves the architecture with `eng resolve`,
shows you the winning recipe and its reasons for confirmation, then
plans, materializes, and verifies with `eng doctor`.

## 4. What you get

```text
my-project/
├── AGENTS.md
├── ARCHITECTURE.md
├── GENTLE.md
├── apps/
├── services/
└── .engineering/
```

The concrete layout depends on the selected recipe. `.engineering/`
keeps the machine state (manifest, provenance, project map,
materialization plan) for later `eng doctor` and evolution commands.

## 5. Hand off to development

Point your development agent at `GENTLE.md`: architecture, surface
topology, and selected foundations are locked; product requirements are
listed as pending implementation. Engineering Platform selects the
foundation and prepares the project — product features are built from
here.

## Next steps

- [CLI reference](../reference/cli.md) — every command and flag.
- [Evolve a project](../guides/evolve-project.md) — grow the project
  later (`surface add`, `extend`, `add`, `update`).
- [Catalog](../concepts/catalog.md) — the curated foundations to choose
  from.
