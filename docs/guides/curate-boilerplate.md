# Curate a boilerplate (§38 runbook)

Each migration follows the same seven steps. Work from the legacy 0.x catalog
as reference; never copy code, only extract knowledge.

## 1. Read the old entry

`platform/boilerplates.json` → `entries[]`: id, repository, `upstream`
(branch/commit/license/observed_at), `decision_status`, `delivery_status`,
`maintenance_tier`, `use_when`/`avoid_when`, `integration` (mode,
update_strategy, adapter/evidence paths).

## 1b. Identity rule (never derive one field from another)

Catalog id, display name, repository URL, upstream pin, adapter id and
Surface are independent concepts: never derive one from another by textual
similarity. Copy `repository` and `upstream.commit` verbatim from the
legacy entry — an id like `hono-api` may point at `JhonMA82/api-starter`,
and `tanstack-admin` at `arhamkhnz/tanstack-shadcn-admin-dashboard`.
Cross-check every repo/pin against the legacy source before writing; when
in doubt, legacy wins and the discrepancy goes in the migration notes.
Record the mapping in the entry `provenance` object: the regression suite
(`internal/catalog/legacy_migration_test.go` over
`testdata/catalog/legacy-boilerplate-baseline.json`) guards the restored
identity.

## 2. Read the adapter

`curation/<id>/adapter.json`: integration mode (overlay / direct /
seed-fork / reference-only), materializer type (git-copy vs
command-generator), destination, setup/checks, prune paths, requirements.
Map to the v1 adapter object (`{name, operations, prune_paths, setup,
checks, managed_files}`); the v1 operation vocabulary is
fetch/copy/prune/template/compose/generate. A `command-generator` legacy
adapter (a CLI that scaffolds its own output directory, e.g. `ignite-cli
new`) maps to the generic `generate` operation (`{run: argv, output:
relative-path}` with `{name}` substitution — see
`docs/decisions/ignite-materialization.md` and
`docs/architecture/materialization.md`); only use a fetch+copy placeholder
for a generator if the generator command cannot be expressed as argv, and
then document the gap in the curation stub and
`docs/decisions/migration-notes.md`.

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

### Pin policy (H6 — immutable pins)

Maximum reproducibility means commit SHAs, not mutable refs:

- Prefer a full commit SHA as `pin` (kind `sha`).
- A tag may be used only with an `expected_sha` recorded alongside when the
  SHA is verifiable offline from existing evidence (pilot log, prior
  `git ls-remote` output quoted in the migration notes). Never invent a
  SHA: an unverifiable tag stays kind `tag` and is recorded in
  `docs/decisions/migration-notes.md` as **re-verify-on-pilot** — the next
  real pilot resolves the tag once, quotes the SHA, and the entry is
  re-pinned to it.
- Per-entry pin kinds live in the migration-notes pin table (§H6 pins).
- CI (`pilots.yml`) materializes through these pins; a moved tag without a
  recorded SHA is a release blocker, not a silent upgrade.
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
- `curation`: formal evidence link — `{"status": "<delivery_status>",
  "evidence": "curation/<id>.md"}`. The status reuses the delivery
  vocabulary (single axis: a set `curation.status` must equal
  `delivery_status`); the evidence path must be a clean relative slash
  path confined to the catalog (no `..`, no absolute paths, no symlink
  escapes) pointing at a real file.

## 4b. Curation enforcement rules (H3 — `eng catalog validate` enforces)

| `delivery_status` | Evidence requirement |
| --- | --- |
| `catalog-only` | Evidence optional (a declared link must still resolve). Not
  selectable for normal materialization. |
| `pilot-ready` | Link required; the file must exist, be non-empty, and address
  the license (even if only to record "unverified" as an explicit gap).
  License checked, repository checked, pin defined, adapter present. |
| `curated` / `released` (`stable` enforces the same bar) | All of the
  above, plus a `Pilot:` line in the stub recording the successful pilot.
  `released` without pilot evidence is invalid. |

Rules: downgrade on doubt, never upgrade — an entry without a Pilot
success record is at most `pilot-ready`, and the reason goes in
`docs/decisions/migration-notes.md`. Never fake evidence to keep a
status. Adding evidence is a data-only operation (new stub +
`curation` link in the entry JSON); it needs no core changes.

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
