# Core

The core is the pure spine `ProjectIntent → Catalog → Resolver →
ArchitectureDecision → Composer → Composition → Planner →
MaterializationPlan`, with side effects pushed to the edges.

- `internal/domain`: contracts with zero infrastructure imports (intent,
  recipe, boilerplate, surface, capability, decision, typed errors). The
  boilerplate adapter accepts a legacy string or a declarative object
  (`{operations, prune_paths, setup, checks, managed_files}`); adapter
  commands are argv arrays, never shell strings.
- `internal/catalog`: loader (base dir plus overlay merge by id), semantic
  validator (unique ids, dangling refs to surfaces/capabilities/
  boilerplates/profiles, pin plus adapter presence) and index (alias,
  provider and active-recipe lookups). Data lives in `catalog/` (recipes,
  boilerplates, surfaces, capabilities, database-profiles, vocabulary,
  curation evidence); the core never hardcodes those lists.
- `internal/resolver`: the pure R1–R10 pipeline (validate, normalize,
  classify, derive, candidates, hard constraints, gaps, score, confidence,
  explain). Deterministic: stable sorts, sha256 intent fingerprint, no
  time/rand/net/LLM/filesystem. See `docs/architecture/routing.md`.
- `internal/composer`: decision plus catalog → providers, destinations,
  database profile. Recovers required work from decision reasons only;
  product features never add components. See
  `docs/architecture/composition.md`.
- `internal/planner` plus `internal/app`: pure plan generation and the thin
  `ResolveProject`/`PlanProject` services.
- `internal/materializer`, `internal/project`, `internal/handoff`: the only
  filesystem/process side effects, project state plus doctor, and pure
  Gentle-handoff generation. `cmd/eng` plus `internal/cli` stay thin; Pi
  integration lives in `integrations/pi/` and the core never imports it.
