# Curate a boilerplate (§38 runbook)

Each migration follows the same seven steps. Work from the legacy 0.x catalog
as reference; never copy code, only extract knowledge.

## 1. Read the old entry

`platform/boilerplates.json` → `entries[]`: id, repository, `upstream`
(branch/commit/license/observed_at), `decision_status`, `delivery_status`,
`maintenance_tier`, `use_when`/`avoid_when`, `integration` (mode,
update_strategy, adapter/evidence paths).

## 2. Read the adapter

`curation/<id>/adapter.json`: integration mode (overlay / direct /
seed-fork / reference-only), materializer type (git-copy vs
command-generator), destination, setup/checks, prune paths, requirements.
Map to the v1 adapter object (`{name, operations, prune_paths, setup,
checks, managed_files}`); the v1 operation vocabulary is
fetch/copy/prune/template/compose. A `command-generator` legacy adapter has
no v1 equivalent yet — record a fetch+copy placeholder and document the gap
in the curation stub and `docs/decisions/migration-notes.md` (as done for
ignite), do not invent semantics.

## 3. Read the evidence

`curation/<id>/evidence.json`: reviewed_at, ai_friendly, evidence list,
gaps, pilot commands. Carry every gap and caveat forward into
`catalog/curation/<id>.md` — a migration must not launder away known
limitations (see `tauri-ui.md` for the pattern).

## 4. Transform to the v1 contract

Write `catalog/boilerplates/<id>.json`:

- `repo`: the upstream URL unchanged.
- `pin`: the reviewed commit. Verify it exists before writing: advertised
  HEAD via `git ls-remote <repo> HEAD`, or the commits API for a frozen
  non-HEAD pin (record which method in the stub). Never float a pin.
- `decision_status` / `delivery_status`: keep the legacy values when they
  are inside the eligible sets (decision:
  curated/default/alternative/specialized/reference; delivery:
  stable/curated/pilot-ready/released). Downgrade on doubt, never upgrade.
- `provides.surfaces`: only surfaces the foundation actually serves (check
  `declared_vs_verified`-style notes — tauri-ui does NOT provide
  `local-filesystem` because the template ships no fs plugin).
- `provides.capabilities`: only architectural consequences (worker topology,
  offline operation, shared backend). Product features go to
  `included_features` (tie-break only).
- `tech_tags`: stack signals used by must-use matching.
- `adapter`: object form with argv `setup`/`checks` mirroring the piloted
  legacy commands.

Add any genuinely new surface or capability first (`catalog/surfaces/`,
`catalog/capabilities/`), with aliases in `catalog/vocabulary/aliases.json`
when a user term maps onto an existing surface (e.g. `kiosk →
public-intake`).

## 5. Validate

```bash
go build -o /tmp/eng ./cmd/eng
/tmp/eng catalog validate
/tmp/eng resolve --input <probe-intent>.json   # spot-check the intended winner
```

## 6. Contract test

Extend the table in `internal/catalog/phase9_test.go`: the entry must assert
repo, verified pin, adapter, provided surfaces, tech tags, composer
eligibility, the `catalog/curation/<id>.md` stub, and participation in one
routing plus one composition fixture.

## 7. Routing/composition scenario

Add at least one fixture under `testdata/routing/` that resolves through the
new foundation and one under `testdata/composition/` that places it as a
provider. Prefer discriminating pairs (native vs offline mobile, SaaS
foundation vs composed multi-app) over lone happy paths. Keep every
pre-existing fixture green — expand additively; if a ranking change is
genuinely an improvement, document it in the migration notes instead of
silently updating expectations.
