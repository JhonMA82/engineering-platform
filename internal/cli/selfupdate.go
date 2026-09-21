package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jhonma82/engineering-platform/internal/app"
	"github.com/jhonma82/engineering-platform/internal/selfupdate"
)

// runSelfUpdate fetches a newer eng binary from the GitHub release line
// and atomically replaces the running executable after sha256
// verification against the release checksums.txt.
//
// Safety rules: --check resolves and verifies without writing; without
// --yes the command prints the plan and writes nothing (piped/CI runs
// stay side-effect free); the binary is never touched before its
// checksum verifies.
func runSelfUpdate(args []string) int {
	flags := newFlagSet("self-update")
	var (
		repoFlag    = flags.String("repo", "", "GitHub repo OWNER/NAME (default "+selfupdate.DefaultRepo+", or ENG_GITHUB_REPO)")
		versionFlag = flags.String("version", "latest", "target release (X.Y.Z, vX.Y.Z or latest)")
		checkFlag   = flags.Bool("check", false, "resolve and verify only, change nothing")
		yesFlag     = flags.Bool("yes", false, "apply the update without prompting")
	)
	flags.Usage = func() {
		fmt.Println("usage: eng self-update [--version X.Y.Z|latest] [--repo OWNER/NAME] [--check] [--yes]")
		fmt.Println("")
		fmt.Println("Update eng from the GitHub release line with sha256 verification.")
		fmt.Println("Without --yes (or with --check) nothing is written.")
		fmt.Println("")
		fmt.Println("Flags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: eng self-update [--version X.Y.Z|latest] [--repo OWNER/NAME] [--check] [--yes]")
		return 2
	}
	checkOnly := *checkFlag
	if !checkOnly && !*yesFlag {
		checkOnly = true // dry plan first; explicit --yes applies it
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	report, err := app.SelfUpdate(ctx, *repoFlag, *versionFlag, "", checkOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "self-update: %v\n", err)
		return 1
	}
	if checkOnly && !*checkFlag {
		// Implicit dry run: the user did not ask for --check, show the
		// plan and require --yes to apply.
		fmt.Printf("current: %s\ntarget:  %s (%s)\n", report.From, report.To, report.Asset)
		if !report.UpdateAvailable {
			fmt.Println("eng is already up to date.")
			return 0
		}
		fmt.Println("re-run with --yes to download, verify and install this release.")
		return 0
	}
	if *checkFlag {
		fmt.Printf("current: %s\ntarget:  %s (%s)\n", report.From, report.To, report.Asset)
		if !report.UpdateAvailable {
			fmt.Println("eng is already up to date.")
			return 0
		}
		fmt.Println("an update is available: run `eng self-update --yes` to install it.")
		return 0
	}
	fmt.Printf("updated eng %s -> %s (%s)\n", report.From, report.To, report.Dest)
	fmt.Println("run `eng version` to confirm.")
	return 0
}
