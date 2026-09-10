package app

import (
	"github.com/jhonma82/engineering-platform/internal/materializer"
)

// InitWorkspace prepares dir as an Engineering Platform agent bootstrap
// workspace. Filesystem effects live in internal/materializer; this is the
// thin orchestration that stamps the binary release line.
func InitWorkspace(dir, agent string, force bool) (*materializer.InitReport, error) {
	return materializer.InitWorkspace(dir, materializer.InitOptions{
		Agent:      agent,
		Force:      force,
		EngVersion: CoreVersion,
	})
}

// CleanupBootstrap removes eng-owned bootstrap resources from an init
// workspace after successful materialization and validation. It is a
// no-op for directories eng init never prepared.
func CleanupBootstrap(dir string) (*materializer.CleanupReport, error) {
	return materializer.CleanupBootstrap(dir)
}
