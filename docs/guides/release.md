# Release runbook

How `eng` goes from a commit to a tagged release. CI gates live in
`.github/workflows/ci.yml` (fast deterministic: gofmt, vet, tests, build,
explicit `eng catalog validate`, cross-compilation matrix); upstream pilots
live in `pilots.yml` and never block CI; tagged releases live in
`release.yml`.

## Everyday CI

Every push/PR runs:

```bash
out=$(gofmt -l .); [ -n "$out" ] && exit 1  # format gate: any output fails
go vet ./...
go test ./...
go build ./cmd/eng
go run ./cmd/eng catalog validate
```

plus a cross-compilation check (`linux/amd64`, `linux/arm64`,
`darwin/arm64`, `windows/amd64`) — built, not executed.

Keep the tree green locally first:

```bash
gofmt -l .
go vet ./...
go test -count=1 ./...
go build -o /tmp/eng ./cmd/eng
/tmp/eng catalog validate
/tmp/eng version
```

## Cutting a release

1. Full validation green (above) plus the relevant pilots:
   `go test ./internal/app/ -run 'Pilot' -count=1 -v` (offline) and, when
   touching generation, the network-gated
   `ENG_UPSTREAM_PILOTS=1 go test ./internal/app/ -run TestPilotIgniteGeneration`.
2. Readiness mapped in `docs/decisions/readiness-v1.md` (§11 checkboxes).
3. Tag and push:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. `release.yml` builds the matrix with ldflags stamps
   (`CoreVersion=<tag>`, commit, build date), writes `checksums.txt`
   (sha256) and publishes all files as release assets.

No goreleaser, no Homebrew/AUR — plain `go build` + `sha256sum` in the
workflow, auditable in one file. Verify after release:

```bash
sha256sum -c checksums.txt
./eng-linux-amd64 version   # Core: 1.0.0, stamped commit/date
```

## Versioning rules

- `internal/version.CoreVersion` is the single canonical source; stamp it
  only via ldflags (Makefile `release` target or the workflow).
- Catalog versions move independently (`catalog/metadata.json`); the load
  gate enforces `min_core_version`/`schema_version` compatibility.
- Re-verify-on-pilot tag pins (migration-notes pin table) must be resolved
  to SHAs before they ride a stable release.
