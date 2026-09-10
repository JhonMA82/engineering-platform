# Engineering Platform — Agent Bootstrap and `eng init`

## Purpose

Define the next usability layer for Engineering Platform after the v1.1 generated-foundations work.

The goal is to make Engineering Platform feel like a tool that is naturally used **from an AI coding agent**, instead of a deterministic engine that requires the user to manually understand and execute:

```text
resolve
plan
materialize
doctor
```

This document focuses on:

- global installation of the `eng` binary;
- project-local agent integration;
- a new `eng init` command;
- PI as the first supported bootstrap target;
- future OpenCode support using the same model;
- disposable bootstrap files;
- persistent minimal Engineering Platform provenance;
- the `/newproject <idea>` workflow;
- clean separation between Engineering Platform, the generated project, and the development agent.

---

## 1. Core principle

Use a hybrid installation model:

```text
GLOBAL
└── eng binary

LOCAL TO EACH PROJECT
├── agent bootstrap integration
├── Engineering Platform discovery prompt/skill
└── temporary initialization state
```

The Engineering Platform binary is a global development tool.

The PI/OpenCode integration used to create a specific project must be local to that project.

Do **not** install Engineering Platform's `/newproject`, discovery skills, or bootstrap instructions globally in PI or OpenCode.

> `eng` is global; agent integration is local; bootstrap is disposable; project provenance is persistent.

---

## 2. Why the integration should not be global

A global PI/OpenCode skill would make Engineering Platform-specific capabilities available inside every unrelated repository.

Examples:

```text
external forks
legacy repositories
small scripts
third-party projects
repositories not created by Engineering Platform
```

Those projects should not automatically expose:

```text
/newproject
project-discovery
Engineering Platform bootstrap instructions
foundation selection behavior
```

Project-local installation also lets the bootstrap resources correspond to the version of Engineering Platform that initialized the workspace.

---

## 3. Installation model

### Recommended general installation

Keep prebuilt release binaries as the recommended installation method for users who do not use Go.

### Go installation

Document a first-class global installation:

```bash
go install github.com/JhonMA82/engineering-platform/cmd/eng@latest
```

The binary will normally be installed into:

```text
$(go env GOPATH)/bin
```

or the path defined by `GOBIN`.

Recommended documentation hierarchy:

```text
Prebuilt binary
  → recommended for most users

go install
  → recommended for Go users/developers

git clone + go build
  → Engineering Platform contributors
```

---

## 4. Current UX problem

Engineering Platform already contains the deterministic core and PI integration resources, but there is no complete first-run experience connecting them.

The current conceptual flow is too manual:

```text
install eng
    ↓
manually locate integrations/pi
    ↓
manually configure PI
    ↓
open PI
    ↓
run project discovery
    ↓
understand Engineering Platform commands
```

The user should not need to know the internal integration layout.

The desired flow is:

```bash
mkdir inventory
cd inventory

eng init
pi
```

Then inside PI:

```text
/newproject Inventory and warehouse management system...
```

From that point onward, the agent integration should orchestrate Engineering Platform.

---

## 5. New command: `eng init`

Add:

```bash
eng init
```

Its responsibility is to initialize the **current directory** as an Engineering Platform agent bootstrap workspace.

It must not create another project directory.

Preferred workflow:

```bash
mkdir my-project
cd my-project
eng init
```

This makes the current directory the stable workspace from the beginning.

---

## 6. `eng init` responsibilities

`eng init` should perform the minimum required work to prepare the workspace.

Conceptual sequence:

```text
1. validate current directory
2. determine Engineering Platform version
3. initialize temporary Engineering Platform state
4. prepare local PI integration
5. install/register required PI support dependency
6. expose /newproject
7. expose project-discovery
8. verify agent prerequisites
9. print the next action
```

Example successful UX:

```text
Engineering Platform initialized.

Agent: PI
Workspace: /path/to/project

Next:
  pi

Then run:
  /newproject <describe your idea>
```

Avoid requiring the user to understand `resolve`, `plan`, or `materialize` during normal use.

Those commands remain available as advanced/debugging commands.

---

## 7. Local PI bootstrap

The PI integration should be installed inside the current project instead of requiring global configuration.

Conceptually:

```text
project/
├── .pi/
│   ├── prompts/
│   │   └── newproject.md
│   └── skills/
│       └── project-discovery/
│           └── SKILL.md
│
└── .engineering/
    └── bootstrap state
```

Use PI's supported project-local mechanisms rather than inventing a parallel plugin system.

If an external PI package is required, `eng init` should install/configure it automatically using the supported local PI mechanism.

The user should not have to manually copy files, run installs inside `integrations/pi`, or edit PI configuration.

---

## 8. Rename `/new-project` to `/newproject`

Prefer:

```text
/newproject
```

instead of:

```text
/new-project
```

The shorter command is easier to remember and restores the simplicity of the original Engineering Platform Python workflow.

The command should accept the initial project idea directly.

Example:

```text
/newproject Inventory system for three warehouses with mobile barcode scanning
```

The initial argument should be treated as the starting discovery context.

Do not ask the user again for the same project idea if it was already supplied in the command.

---

## 9. Target PI workflow

The desired end-to-end experience is:

```text
mkdir project
cd project
eng init
pi
```

Then:

```text
/newproject <idea>
```

PI should then handle:

```text
project discovery
      ↓
ProjectIntent
      ↓
eng resolve
      ↓
eng plan
      ↓
materialization
      ↓
eng doctor
      ↓
development handoff
```

The user should interact primarily with PI during discovery.

Engineering Platform remains responsible for deterministic architecture selection and materialization.

---

## 10. Current directory must become the project

Do not return to the old model where Engineering Platform generates into a separate output directory after discovery.

Preferred model:

```text
project directory chosen first
        ↓
eng init
        ↓
discovery occurs inside project
        ↓
materialization occurs into the same workspace
```

This keeps agent context, intent, plan, provenance, generated foundation, and project code associated with the same workspace from the beginning.

---

## 11. Bootstrap must be disposable

The local resources installed by `eng init` are **bootstrap infrastructure**.

They are not necessarily part of the final software project.

After successful materialization, Engineering Platform should automatically remove any resources that only existed to create the project.

Examples:

```text
/newproject prompt
project-discovery skill
temporary bootstrap configuration
temporary discovery artifacts
temporary initialization state
```

> The generated application is the product. Engineering Platform is the tool that created it.

---

## 12. Do not delete all `.engineering/`

Engineering Platform should keep only a minimal persistent project identity/provenance record.

This is useful for future commands such as:

```text
eng doctor
eng explain
eng update
eng surface add
eng extend
```

Persistent information should be limited to useful facts such as:

```text
Engineering Platform version
foundation(s) used
foundation version/tag/commit
profiles
arguments
adapter fingerprint
materialization metadata
surface relationships
project identity
```

Do not preserve discovery conversation noise or bootstrap-only resources.

Conceptual final state:

```text
project/
├── application code
├── package.json / go.mod / etc.
├── AGENTS.md
├── PROJECT.md
├── framework-specific agent context
└── .engineering/
    └── minimal persistent provenance/state
```

---

## 13. Cleanup lifecycle

Desired lifecycle:

```text
eng init
   ↓
.pi bootstrap
.engineering bootstrap
   ↓
/newproject ...
   ↓
discovery
   ↓
resolve
   ↓
plan
   ↓
materialize
   ↓
doctor
   ↓
bootstrap cleanup
   ↓
final project + minimal Engineering Platform provenance
```

Cleanup should only occur after successful materialization and validation.

If generation fails, retain enough state to diagnose or retry safely.

---

## 14. `eng init` must be idempotent

Running `eng init` multiple times must be safe.

First execution:

```text
Engineering Platform initialized.
PI integration installed.
```

Subsequent execution:

```text
Engineering Platform already initialized.
PI integration is up to date.
```

It must not:

```text
overwrite user files unexpectedly
duplicate prompts
duplicate skills
destroy existing agent configuration
reset project state
```

Engineering Platform should only update files it owns/manages.

---

## 15. Ownership boundaries

Files created by `eng init` should be clearly attributable to Engineering Platform.

Never recursively replace an entire:

```text
.pi/
.opencode/
```

directory.

Only manage the specific Engineering Platform integration resources.

Cleanup must be ownership-based, not pattern-based.

---

## 16. Agent detection

For the first implementation, PI can remain the primary/default supported agent.

Avoid overengineering.

Recommended initial behavior:

```bash
eng init
```

→ installs PI integration.

Future evolution may support:

```bash
eng init --agent pi
eng init --agent opencode
```

Automatic detection can be added later if it remains deterministic and understandable.

Do not block `eng init` architecture on multi-agent support.

---

## 17. Future OpenCode integration

Use the same architectural principle as PI.

OpenCode itself may be globally installed, but Engineering Platform-specific skills/instructions should remain local to the project.

Conceptually:

```text
project/
└── .opencode/
    └── skills/
        └── engineering-project-discovery/
            └── SKILL.md
```

Possible future command:

```bash
eng init --agent opencode
```

Do not install Engineering Platform's project-discovery skill globally into OpenCode.

The same cleanup rules should apply after materialization.

---

## 18. Persistent boilerplate agent tooling is different

Do not confuse the temporary Engineering Platform bootstrap with the AI-friendly tooling intentionally provided by generated boilerplates.

A generated Next/TanStack foundation may intentionally preserve:

```text
AGENTS.md
PROJECT.md
docs/ai/
generate:feature
generate:dashboard
generate:crud
ai:context
validators
```

Those are **project development capabilities** and should remain.

By contrast:

```text
/newproject
Engineering Platform project-discovery skill
temporary Engineering Platform bootstrap config
```

exist only to create the project and can be removed.

```text
Engineering Platform bootstrap
→ temporary

Generated boilerplate development tooling
→ persistent
```

---

## 19. Multi-foundation projects

The cleanup model must also work when Engineering Platform composes multiple foundations/surfaces.

Example:

```text
web
api
mobile
```

`eng init` still occurs once at project root.

After successful composition:

```text
temporary agent bootstrap disappears
minimal root provenance remains
surface-specific AI/tooling remains where appropriate
```

Do not copy the root `/newproject` integration into every generated surface.

---

## 20. Existing `eng start`

Do not try to restore the old Python behavior inside `eng start`.

The current Go `eng start` can remain a deterministic convenience command that runs the existing pipeline.

New user-facing project creation should use:

```text
eng init
```

followed by:

```text
/newproject
```

Responsibility split:

```text
eng init
→ prepare workspace + agent integration

/newproject
→ conversational discovery and orchestration

eng start
→ optional deterministic CLI pipeline / advanced usage
```

---

## 21. Normal path vs advanced path

### Normal AI-assisted path

```bash
mkdir project
cd project
eng init
pi
```

Then:

```text
/newproject <idea>
```

### Advanced/manual path

Keep direct commands available:

```text
eng resolve
eng plan
eng materialize
eng doctor
eng explain
```

This preserves scripting, debugging, CI, and agent-independent use.

---

## 22. Documentation changes

Update the root README so the first-run experience is obvious.

Recommended order:

### Install

```text
Prebuilt binary
Go install
Build from source
```

### Create a project with PI

```bash
mkdir my-project
cd my-project
eng init
pi
```

Then:

```text
/newproject <idea>
```

### Advanced CLI

Only after the primary workflow explain:

```text
resolve
plan
materialize
doctor
start
```

Do not make users learn the engine internals before creating their first project.

---

## 23. Suggested CLI UX

Potential future options:

```text
--agent pi
--agent opencode
--force
--no-agent
```

Do not add options until there is a real use case.

Keep v1 of `eng init` small.

---

## 24. Safety requirements

`eng init` should:

- never delete arbitrary user files;
- never overwrite unknown `.pi` or `.opencode` resources;
- use atomic writes where practical;
- detect managed resource version/state;
- refuse dangerous initialization contexts if necessary;
- never recursively initialize inside a materialized surface unless explicitly supported;
- only clean files known to belong to the Engineering Platform bootstrap.

---

## 25. Suggested internal model

Avoid coupling the command directly to PI-specific filesystem operations.

Conceptually:

```text
InitWorkspace
    ↓
AgentBootstrap
    ├── PiBootstrap
    └── OpenCodeBootstrap (future)
```

The core lifecycle remains generic:

```text
validate workspace
create Engineering state
install agent adapter
verify
report
```

PI-specific paths and commands belong to the PI adapter.

Do not let multi-agent abstraction become a prerequisite for shipping PI support.

---

## 26. Versioning bootstrap resources

Bootstrap resources embedded in `eng` should have a version or fingerprint.

This allows `eng init` to determine:

```text
not installed
installed and current
installed but outdated
```

Possible metadata:

```text
agent: pi
bootstrapVersion
engVersion
resourceFingerprint
initializedAt
```

Temporary bootstrap metadata should be removed or reduced after materialization.

---

## 27. Package resources inside the Go binary

Prefer embedding the PI bootstrap resources into the `eng` binary rather than requiring users to clone the Engineering Platform repository.

Desired experience:

```text
download/install one eng binary
        ↓
eng init
        ↓
all required Engineering Platform PI files available
```

The release artifact must be sufficient to initialize a project.

Do not require:

```text
git clone engineering-platform
cd integrations/pi
manual copy
```

---

## 28. Acceptance criteria

### Installation

- [ ] README documents prebuilt binary installation.
- [ ] README documents `go install github.com/JhonMA82/engineering-platform/cmd/eng@latest`.
- [ ] installed binary works globally from `PATH`.

### Workspace initialization

- [ ] `eng init` exists.
- [ ] it initializes the current directory.
- [ ] it does not create an extra project directory.
- [ ] it is idempotent.
- [ ] it installs PI integration locally.
- [ ] it does not overwrite unrelated PI configuration.
- [ ] it reports clear next steps.

### PI workflow

- [ ] `/newproject` is available after `eng init`.
- [ ] `/newproject <idea>` accepts the initial idea as an argument.
- [ ] discovery does not repeat information already supplied.
- [ ] PI can execute the Engineering Platform pipeline without manual user CLI orchestration.

### Materialization

- [ ] project is materialized into the current workspace.
- [ ] one or multiple foundations can be composed.
- [ ] normal `doctor` validation still runs.
- [ ] bootstrap cleanup runs only after successful materialization.

### Cleanup

- [ ] `/newproject` bootstrap prompt is removed when no longer needed.
- [ ] project-discovery bootstrap skill is removed when no longer needed.
- [ ] temporary Engineering Platform bootstrap files are removed.
- [ ] generated boilerplate AI tooling remains.
- [ ] minimal `.engineering/` provenance remains.
- [ ] failed materialization preserves enough diagnostic/retry state.

### Architecture

- [ ] no changes are required to resolver/composer/planner architecture.
- [ ] deterministic core remains agent-independent.
- [ ] PI is an adapter, not part of core decision logic.
- [ ] future OpenCode support can reuse the same initialization lifecycle.
- [ ] Engineering Platform-specific OpenCode/PI bootstrap remains local, not global.

---

## 29. Suggested validation scenarios

### Scenario A — fresh PI project

```bash
mkdir inventory
cd inventory
eng init
pi
```

Then:

```text
/newproject Inventory application with web admin and API
```

Verify successful materialization and cleanup.

### Scenario B — run `eng init` twice

```bash
eng init
eng init
```

Expected:

```text
no duplicated resources
no destroyed files
no duplicate dependency installation
clear already-initialized/up-to-date result
```

### Scenario C — existing `.pi` directory

Verify:

```text
existing user files preserved
Engineering Platform files added only in owned paths
```

### Scenario D — failed materialization

Verify:

```text
no destructive cleanup
bootstrap remains usable
diagnostic state remains
retry is possible
```

### Scenario E — multiple foundations

Verify:

```text
single root bootstrap
successful composition
root provenance retained
bootstrap removed after success
surface development tooling retained
```

---

## 30. Recommended implementation order

### Phase 1 — Installation documentation

1. document prebuilt binary clearly;
2. add `go install`;
3. clarify `PATH`;
4. separate user installation from contributor build instructions.

### Phase 2 — `eng init`

1. create command;
2. initialize current directory;
3. add minimal `.engineering` bootstrap metadata;
4. embed PI resources in the binary;
5. materialize local PI resources;
6. verify prerequisites;
7. make initialization idempotent.

### Phase 3 — `/newproject`

1. rename/migrate `new-project` → `newproject`;
2. accept initial idea argument;
3. reuse existing project-discovery skill;
4. orchestrate resolve/plan/materialize/doctor;
5. target current workspace.

### Phase 4 — Bootstrap cleanup

1. mark Engineering Platform-owned bootstrap resources;
2. remove them after successful materialization;
3. preserve minimal provenance;
4. preserve generated boilerplate development tooling;
5. add retry-safe failure behavior.

### Phase 5 — OpenCode

Only after PI workflow is proven.

1. implement local OpenCode bootstrap adapter;
2. support `eng init --agent opencode`;
3. reuse the same ownership/cleanup lifecycle.

---

## 31. Non-goals

This work should **not**:

- redesign Resolver;
- redesign Composer;
- redesign Planner;
- replace Generated Foundations;
- make PI required for direct CLI use;
- make OpenCode required;
- globally install Engineering Platform discovery skills;
- add a daemon;
- add a central project registry;
- add cloud state;
- preserve temporary bootstrap artifacts forever;
- remove the existing manual/advanced CLI pipeline.

---

## 32. Final target UX

```bash
go install github.com/JhonMA82/engineering-platform/cmd/eng@latest

mkdir inventory
cd inventory

eng init
pi
```

Then:

```text
/newproject Inventory system for multiple warehouses with user roles and API
```

Engineering Platform and PI handle:

```text
discovery
→ stack/foundation selection
→ planning
→ generated foundation(s)
→ validation
→ clean handoff
```

The resulting repository contains the software project and only the Engineering Platform metadata that remains useful for future lifecycle operations.

> Engineering Platform should be visible while creating and evolving the architecture, but it should not pollute the final application with bootstrap infrastructure that has already served its purpose.
