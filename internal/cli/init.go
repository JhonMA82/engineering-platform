package cli

import (
	"fmt"
	"os"

	"github.com/jhonma82/engineering-platform/internal/app"
)

// runInit prepares the current directory as an Engineering Platform
// workspace. It is a thin adapter: flags in, app.InitWorkspace, report out.
// Filesystem effects live in internal/materializer.
func runInit(args []string) int {
	flags := newFlagSet("init")
	var (
		agentFlag   = flags.String("agent", "pi", "Agent to bootstrap (pi, opencode)")
		forceFlag   = flags.Bool("force", false, "Repair eng-owned files even when customized")
		noAgentFlag = flags.Bool("no-agent", false, "Initialize only Engineering Platform state, skip agent integration")
	)
	flags.Usage = func() {
		fmt.Println("usage: eng init [--agent pi|opencode] [--force] [--no-agent]")
		fmt.Println("")
		fmt.Println("Initialize the current directory as an Engineering Platform workspace.")
		fmt.Println("")
		fmt.Println("Flags:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}

	agent := *agentFlag
	if *noAgentFlag {
		agent = "none"
	}
	if agent != "pi" && agent != "opencode" && agent != "none" {
		fmt.Fprintf(os.Stderr, "error: unknown agent %q (want pi, opencode or none)\n", agent)
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get current directory: %v\n", err)
		return 1
	}

	report, err := app.InitWorkspace(cwd, agent, *forceFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if report.AlreadyUpToDate {
		fmt.Println("Engineering Platform already initialized.")
		fmt.Printf("Agent: %s\n", report.Agent)
		fmt.Println("PI integration is up to date.")
		return 0
	}

	fmt.Println("Engineering Platform initialized.")
	fmt.Printf("\nAgent: %s\n", report.Agent)
	fmt.Printf("Workspace: %s\n", report.Workspace)
	for _, p := range report.Created {
		fmt.Printf("  created %s\n", p)
	}
	for _, p := range report.Updated {
		fmt.Printf("  updated %s\n", p)
	}
	for _, p := range report.Kept {
		fmt.Printf("  kept %s (customized; use --force to repair)\n", p)
	}
	fmt.Println("\nNext:")
	if agent == "pi" {
		fmt.Println("  pi")
		fmt.Println("\nThen run:")
		fmt.Println("  /newproject <describe your idea>")
	} else if agent == "opencode" {
		fmt.Println("  opencode")
		fmt.Println("\nThen run:")
		fmt.Println("  /newproject <describe your idea>")
	} else {
		fmt.Println("  eng start --intent intent.json --output .")
	}
	return 0
}
