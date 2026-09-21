package app

import (
	"context"
	"os"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/materializer"
	"github.com/jhonma82/engineering-platform/internal/selfupdate"
)

// SelfUpdate resolves repo/token defaults (flags, then ENG_GITHUB_REPO /
// GITHUB_TOKEN environment) and runs the binary update through
// internal/materializer, the only package with network/filesystem
// effects. It is the thin orchestration behind `eng self-update`.
func SelfUpdate(ctx context.Context, repo, tag, token string, checkOnly bool) (*materializer.SelfUpdateReport, error) {
	if strings.TrimSpace(repo) == "" {
		repo = strings.TrimSpace(os.Getenv("ENG_GITHUB_REPO"))
	}
	if strings.TrimSpace(repo) == "" {
		repo = selfupdate.DefaultRepo
	}
	if strings.TrimSpace(token) == "" {
		token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}
	if strings.TrimSpace(token) == "" {
		token = strings.TrimSpace(os.Getenv("ENG_GITHUB_TOKEN"))
	}
	return materializer.RunSelfUpdate(ctx, materializer.SelfUpdateOptions{
		Repo:      repo,
		Tag:       tag,
		Current:   CoreVersion,
		CheckOnly: checkOnly,
		Token:     token,
	})
}
