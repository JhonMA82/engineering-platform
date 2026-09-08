# Catalog

The catalog (`catalog/`) is the declarative knowledge base the core reasons
over. It is data, not code: adding a foundation, surface or capability never
requires recompiling or changing the Go core.

## Layout

```text
catalog/
  metadata.json          # catalog_version, min/max_core_version, schema_version
  recipes/               # GP-01..GP-07: provides, composition, database policy
  boilerplates/          # id, repo, pin, adapter, provides, technology, curation
  surfaces/              # open vocabulary (tui exists with no provider yet)
  capabilities/          # architectural consequences, never product features
  database-profiles/     # engine/provider/supports + recipe allow-lists
  vocabulary/aliases.json# user terms → canonical ids
  curation/              # evidence stubs linked by curation.status/evidence
```

## Contracts

- **Versioning is independent from core releases.** `catalog_version` moves
  on curation changes; `min_core_version` is a real semver floor enforced at
  load (`catalog requires core >= X`). See ADR-0007.
- **Surfaces and capabilities are an open vocabulary**, validated by the
  catalog — never a closed enum in Go (§7.4 regression test resolves a
  synthetic overlay surface with zero core changes).
- **Pins are immutable references.** Commit SHA preferred; a tag is allowed
  only with a recorded SHA, otherwise it is `re-verify-on-pilot`. Pin kinds
  are recorded per entry in `docs/decisions/migration-notes.md`; the policy
  lives in `docs/guides/curate-boilerplate.md` §4.
- **Curation gates selection** (`eng catalog validate`): `catalog-only`
  entries exist without full evidence but are not selectable;
  `pilot-ready` needs license/repo/pin/adapter/basic evidence; `curated` and
  `released` additionally need a recorded successful pilot. See ADR-0009.
- **Evidence paths stay inside the catalog** (no `..`, absolute paths or
  symlink escapes).

## Evolving the catalog

Additive by default: new entries, new surfaces, new evidence links. Overlay
dirs (`catalog.LoadDir` + `MergeOverlay`) let tests and pilots extend the
base catalog without touching it. Removing or renaming an id is a breaking
change — record it in the migration notes with the affected recipes.
