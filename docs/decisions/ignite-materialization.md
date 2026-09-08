# Ignite materialization (H4)

Date: 2026-09-08 · Method: upstream README + pinned CLI source
(`src/commands/new.ts` @ `e829d2f`) + npm registry metadata. No core
special-casing was added (`TestNoBoilerplateSpecialCases` guards this).

## How Ignite generates a project

Ignite is **not a starter repo**: cloning `infinitered/ignite` yields the
CLI plus its bundled boilerplate template, not a runnable app. A project is
scaffolded by the generator CLI:

```bash
npx ignite-cli@<version> new <AppName> [--yes] [--bundle-id ...]
```

- **Inputs**: app name (positional, becomes the output directory);
  `--yes` accepts every default non-interactively (bundle id
  `com.<app>`, git init+commit, packager install, default workflow);
  optional overrides for bundle id, git, install-deps and packager.
  Verified against the pinned source: `yes`, `bundle`, `git`,
  `installDeps`, `packager` are all `parameter.option` inputs — no
  interactive prompt is needed for a `--yes` run.
- **Outputs**: a fresh `<AppName>/` directory with the generated
  React Native / Expo project (React Native 0.81, React 19, TypeScript 5,
  Expo SDK 55 per the pinned README). The CLI caps creation at 10 minutes
  internally.
- **Commands**: `npx ignite-cli@11.5.0 new {name} --yes`. Version 11.5.0
  exists on the npm registry (metadata + integrity hash verified
  2026-09-08); the v11.5.0 git tag exists upstream. The repo pin
  `e829d2f` (advertised HEAD) records the reviewed CLI source snapshot.

## Verdict

The flow is **genuinely unrepresentable** with fetch/copy/prune/template:
those primitives copy a checked-in tree, while the generator produces its
own output directory (running setup afterwards would nest the app inside a
copy of the CLI source). Per §4.3 the engine gains ONE generic operation:

- `generate`: `{run: argv, output: relative-path}` with `{name}`
  substitution from the destination basename. Argv-only (no shell),
  controlled cwd (fresh work dir), validated output confined to that dir,
  work-dir cleanup on failure, exit capture via the existing command
  runner. See `docs/architecture/materialization.md`.
- The ignite adapter declares exactly this (no `if ignite` in core);
  future generator foundations reuse the same shape.

## Honest setup/check position

Post-generation install is part of the generator itself (`installDeps`
defaults to true under `--yes`); no additional setup/check commands are
declared because none have been piloted. The network-gated pilot
(`TestPilotIgniteGeneration`, needs npx + network + minutes, skipped in
`-short`/offline CI) runs the real adapter end to end; confirming or
extending setup/checks is its explicit follow-up, not guessed here.
