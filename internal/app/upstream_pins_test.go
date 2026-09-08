package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jhonma82/engineering-platform/internal/catalog"
)

// upstreamPinCases are the five materializable foundations whose restored
// pins this pilot verifies: the repository must be reachable and the pinned
// commit must exist upstream. fastapi and goship are excluded by design:
// they declare no pin (catalog-only), so there is nothing to verify.
var upstreamPinCases = []struct {
	id   string
	repo string
	pin  string
}{
	{"stardrive", "https://github.com/peltmonger/stardrive", "5c449810b763140ac72133ff4ae63d8497cce77a"},
	{"tanstack-admin", "https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard", "e6e5d3bdb7974d4a2283df763ea2dd222d82e1f0"},
	{"hono-api", "https://github.com/JhonMA82/api-starter", "360eb274cc5936fee5aab88eb8bd94977e95dfc9"},
	{"tanstack-transactional-pwa", "https://github.com/JhonMA82/tanstack-transactional-pwa", "f2571ea8efb2e5a2ceaaafa7ff38c523dca1ac0c"},
	{"next-admin", "https://github.com/arhamkhnz/next-shadcn-admin-dashboard", "15e0a081bc1acad2b47adc638471b6e67fa36f10"},
}

// TestUpstreamCommitPins is the network-gated pilot (§13): repository
// reachable plus pinned commit exists, per migrated foundation. Normal
// suites never run it: besides -short it requires ENG_UPSTREAM_PILOTS=1,
// so `go test ./...` stays offline and deterministic. Explicit run:
//
//	ENG_UPSTREAM_PILOTS=1 go test ./internal/app/ -run TestUpstreamCommitPins -v
//
// A fully unreachable upstream skips (network probe, like the ignite
// pilot); a reachable repository with a missing commit fails.
func TestUpstreamCommitPins(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode: upstream pilot skipped")
	}
	if os.Getenv("ENG_UPSTREAM_PILOTS") != "1" {
		t.Skip("ENG_UPSTREAM_PILOTS != 1: upstream pilot skipped")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH: upstream pilot skipped")
	}
	cat, err := catalog.Load("")
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	idx := catalog.NewIndex(cat)
	client := &http.Client{Timeout: 30 * time.Second}
	for _, tc := range upstreamPinCases {
		t.Run(tc.id, func(t *testing.T) {
			bp, ok := idx.Boilerplate(tc.id)
			if !ok {
				t.Fatalf("%s missing from catalog", tc.id)
			}
			if bp.EffectiveRepo() != tc.repo || bp.Pin != tc.pin {
				t.Fatalf("%s catalog drift: repo=%q pin=%q", tc.id, bp.EffectiveRepo(), bp.Pin)
			}
			if out, err := exec.Command("git", "ls-remote", tc.repo, "HEAD").CombinedOutput(); err != nil {
				t.Skipf("upstream unreachable: pilot skipped (%v: %s)", err, string(out))
			}
			req, err := http.NewRequest("GET",
				fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", shortRepo(tc.repo), tc.pin), nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("commits API unreachable: pilot skipped (%v)", err)
			}
			defer resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusOK:
				// Commit exists upstream.
			case http.StatusNotFound:
				t.Fatalf("pin %s does not exist in %s", tc.pin, tc.repo)
			default:
				t.Skipf("commits API status %d: pilot inconclusive, skipped", resp.StatusCode)
			}
		})
	}
}

// shortRepo strips the https://github.com/ prefix for API paths.
func shortRepo(repo string) string {
	const prefix = "https://github.com/"
	if len(repo) > len(prefix) && repo[:len(prefix)] == prefix {
		return repo[len(prefix):]
	}
	return repo
}
