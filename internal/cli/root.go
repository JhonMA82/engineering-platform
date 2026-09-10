// Package cli exposes the eng command surface with stdlib flags only.
package cli

import (
	"flag"
	"fmt"
	"os"
)

// Main dispatches eng subcommands and returns the process exit code.
func Main(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "resolve":
		return runResolve(args[1:])
	case "explain":
		return runExplain(args[1:])
	case "catalog":
		return runCatalog(args[1:])
	case "start":
		return runStart(args[1:])
	case "plan":
		return runPlan(args[1:])
	case "materialize":
		return runMaterialize(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "surface":
		return runSurface(args[1:])
	case "add":
		return runAdd(args[1:])
	case "extend":
		return runExtend(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "version":
		return runVersion(args[1:])
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		usage()
		return 2
	}
}

func runSurface(args []string) int {
	if len(args) == 0 || args[0] != "add" {
		fmt.Fprintln(os.Stderr, "usage: eng surface add --project <dir> --surface <id> [--provider <boilerplate>]")
		return 2
	}
	return runSurfaceAdd(args[1:])
}

func usage() {
	fmt.Println("usage: eng <init|resolve|explain|catalog|plan|materialize|start|doctor|surface|add|extend|update|version> [flags]")
	fmt.Println("  init [--agent pi|opencode] [--force] [--no-agent]")
	fmt.Println("  resolve --input intent.json [--json] [--verbose] [--catalog-dir DIR]")
	fmt.Println("  plan --input intent.json [--json] [--catalog-dir DIR]")
	fmt.Println("  materialize --plan plan.json --output <dir> [--intent intent.json] [--decision decision.json] [--catalog-dir DIR]")
	fmt.Println("  start --intent intent.json --output <dir> [--catalog-dir DIR] [--dry-run] [--json]")
	fmt.Println("  doctor [--project <dir>] [--json]")
	fmt.Println("  surface add --project <dir> --surface <id> [--provider <boilerplate>] [--catalog-dir DIR]")
	fmt.Println("  extend --project <dir> --surface <id> [--catalog-dir DIR]")
	fmt.Println("  add --project <dir> --requirement \"<text>\" [--scope required_now|planned_later] [--catalog-dir DIR]")
	fmt.Println("  update --project <dir> [--json] [--catalog-dir DIR]")
	fmt.Println("  explain --input decision.json")
	fmt.Println("  catalog [list] [--catalog-dir DIR] [--json]")
	fmt.Println("  catalog show <id> [--catalog-dir DIR] [--json]")
	fmt.Println("  catalog validate [--catalog-dir DIR]")
	fmt.Println("  version")
}

func newFlagSet(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ContinueOnError)
}
