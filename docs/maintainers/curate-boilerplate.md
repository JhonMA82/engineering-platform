# Curate a boilerplate

Curating a boilerplate follows the same seven steps. Work from the upstream
repository and any prior review as reference; never copy code, only extract
knowledge.

## 1. Read the upstream

Record the candidate: repository URL, default branch and current commit,
license, maintenance signals (recent activity, releases), and the
`decision_status` / `delivery_status` you will claim (downgrade on doubt —
see the enforcement rules below). The upstream repository is authoritative;
the historic 0.x catalog no longer lives in this tree.

## 1b. Identity rule (never derive one field from another)

Catalog id, display name, repository URL, upstream pin, adapter id and
Surface are independent concepts: never derive one from another by textual
similarity. Copy `repository` and the reviewed commit verbatim from the
upstream source — an id like `hono-api` may point at `JhonMA82/api-starter`,
and `tanstack-admin` at `arhamkhnz/tanstack-shadcn-admin-dashboard`.
Cross-check every repo/pin against the upstream source before writing; when
in doubt, the upstream wins and the discrepancy goes in the entry curation
stub (`catalog/curation/<id>.md`).
Record the mapping in the entry `provenance` object: the regression suite
(`internal/catalog/legacy_migration_test.go` over
`testdata/catalog/legacy-boilerplate-baseline.json`) guards the restored
identity.
## 2. Read how it installs

Determine the materializer type: a git-copy tree (clone a pinned commit and
copy/prune it) or a command-generator (a CLI that scaffolds its own output
directory, e.g. `ignite-cli new`). Note destinations, setup/check commands,
prune paths and requirements.
Map to the v1 adapter object (`{name, operations, prune_paths, setup,
checks, managed_files}`); the v1 operation vocabulary is
fetch/copy/prune/template/compose/generate. A `command-generator` legacy
adapter (a CLI that scaffolds its own output directory, e.g. `ignite-cli
new`) maps to the generic `generate` operation (`{run: argv, output:
relative-path}` with `{name}` substitution — see
`docs/architecture/materialization.md` and
`docs/architecture/generated-foundations.md`); only use a fetch+copy placeholder
for a generator if the generator command cannot be expressed as argv, and
then document the gap in the curation stub.

## 3. Record the evidence

Write the curation stub `catalog/curation/<id>.md` first: reviewed date,
license, maintenance signal, pin and verification method, gaps and pilot
commands. Carry every gap and caveat forward into the stub — curation must not launder away known
limitations (see `tauri-ui.md` for the pattern).

## 4. Transform to the v1 contract

Write `catalog/boilerplates/<id>.json`:

- `repo`: the upstream URL unchanged.
- `pin`: the reviewed commit. Verify it exists before writing: advertised
  HEAD via `git ls-remote <repo> HEAD`, or the commits API for a frozen
  non-HEAD pin (record which method in the stub). Never float a pin.

### Pin policy (immutable pins)

Maximum reproducibility means commit SHAs, not mutable refs:

- Prefer a full commit SHA as `pin` (kind `sha`).
- A tag may be used only with an `expected_sha` recorded alongside when the
  SHA is verifiable offline from existing evidence (pilot log, prior
  `git ls-remote` output quoted in the curation stub). Never invent a
  SHA: an unverifiable tag stays kind `tag` and is recorded in the entry
  curation stub (`catalog/curation/<id>.md`) as **re-verify-on-pilot** — the
  next real pilot resolves the tag once, quotes the SHA, and the entry is
  re-pinned to it.
- Per-entry pin kinds live with the entry (`pin` plus the verification
  method quoted in the curation stub). All current catalog pins are full
  SHAs; the historical tag-pin table was retired with the migration notes
  (git history).
- CI (`pilots.yml`) materializes through these pins; a moved tag without a
  recorded SHA is a release blocker, not a silent upgrade.
- `decision_status` / `delivery_status`: choose values inside the eligible sets (decision:
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
  upstream commands.
- `curation`: formal evidence link — `{"status": "<delivery_status>",
  "evidence": "curation/<id>.md"}`. The status reuses the delivery
  vocabulary (single axis: a set `curation.status` must equal
  `delivery_status`); the evidence path must be a clean relative slash
  path confined to the catalog (no `..`, no absolute paths, no symlink
  escapes) pointing at a real file.

## 4b. Curation enforcement rules (`eng catalog validate` enforces)

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
success record is at most `pilot-ready`, and the reason goes in the entry
curation stub. Never fake evidence to keep a
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
genuinely an improvement, document it in the commit message instead of
silently updating expectations.

## Generated foundations

Curating a factory (as opposed to a copyable tree) needs the evidence
listed in `docs/architecture/generated-foundations.md` (Curation):
validated command, version/pin, profiles and default, output path,
non-interactive pilot, factory-leak check and `AGENTS.md` behavior.
