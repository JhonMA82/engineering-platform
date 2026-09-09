package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// runCatalog dispatches the catalog read commands. Bare `eng catalog` lists
// the index (recipes + boilerplates + surfaces); `show` renders one entry by
// id across every catalog kind; `validate` keeps its historical behavior.
func runCatalog(args []string) int {
	if len(args) == 0 {
		return runCatalogList(nil)
	}
	switch args[0] {
	case "validate":
		return runCatalogValidate(args[1:])
	case "show":
		return runCatalogShow(args[1:])
	case "list":
		return runCatalogList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown catalog command %q\n", args[0])
		fmt.Fprintln(os.Stderr, "usage: eng catalog [list|show <id>|validate] [--catalog-dir DIR] [--json]")
		return 2
	}
}

func runCatalogValidate(args []string) int {
	fs := newFlagSet("catalog validate")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cat, err := catalog.Load(*catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catalog validate: %v\n", err)
		return 1
	}
	if err := catalog.Validate(cat); err != nil {
		fmt.Fprintf(os.Stderr, "catalog validate: %v\n", err)
		return 1
	}
	// H3 curation enforcement: evidence links resolve against the overlay
	// first (overlay entries may ship their own evidence) then the base
	// catalog — data-only evidence additions need no core changes.
	dirs := []string{catalog.BaseDir()}
	if strings.TrimSpace(*catalogDir) != "" {
		dirs = append([]string{*catalogDir}, dirs...)
	}
	if err := catalog.ValidateCuration(cat, dirs); err != nil {
		fmt.Fprintf(os.Stderr, "catalog validate: %v\n", err)
		return 1
	}
	fmt.Printf("catalog OK: version %s, %d recipes, %d boilerplates\n",
		cat.CatalogVersion, len(cat.Recipes), len(cat.Boilerplates))
	return 0
}

func runCatalogList(args []string) int {
	fs := newFlagSet("catalog list")
	catalogDir := fs.String("catalog-dir", "", "overlay catalog directory")
	asJSON := fs.Bool("json", false, "print the index as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cat, err := catalog.Load(*catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catalog list: %v\n", err)
		return 1
	}
	if *asJSON {
		out, _ := json.MarshalIndent(struct {
			Recipes      []domain.Recipe      `json:"recipes"`
			Boilerplates []domain.Boilerplate `json:"boilerplates"`
			Surfaces     []domain.Surface     `json:"surfaces"`
		}{Recipes: cat.Recipes, Boilerplates: cat.Boilerplates, Surfaces: cat.Surfaces}, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	fmt.Printf("recipes (%d):\n", len(cat.Recipes))
	for _, r := range cat.Recipes {
		fmt.Printf("  %-28s %-8s %s\n", r.ID, r.Version, r.Description)
	}
	fmt.Printf("boilerplates (%d):\n", len(cat.Boilerplates))
	for _, b := range cat.Boilerplates {
		fmt.Printf("  %-28s %-16s surfaces=%s\n", b.ID, b.Pin, surfaceList(b.Provides.Surfaces))
	}
	fmt.Printf("surfaces (%d):\n", len(cat.Surfaces))
	for _, s := range cat.Surfaces {
		fmt.Printf("  %-28s %s\n", string(s.ID), s.Description)
	}
	return 0
}

func runCatalogShow(args []string) int {
	// Accept flags on either side of the id (`show GP-06 --json` and
	// `show --json GP-06`): stdlib flag.Parse stops at the first positional,
	// so flags are pre-scanned out before parsing positionals.
	catalogDir, asJSON, rest := splitCatalogShowArgs(args)
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: eng catalog show <id> [--catalog-dir DIR] [--json]")
		return 2
	}
	id := rest[0]
	cat, err := catalog.Load(catalogDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catalog show: %v\n", err)
		return 1
	}
	kind, found := lookupCatalogEntry(cat, id)
	if found == nil {
		fmt.Fprintf(os.Stderr, "catalog show: unknown id %q\n%s", id, validIDHint(cat))
		return 1
	}
	if asJSON {
		out, _ := json.MarshalIndent(struct {
			Kind  string `json:"kind"`
			Entry any    `json:"entry"`
		}{Kind: kind, Entry: found}, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	fmt.Printf("%s %s\n", kind, id)
	printCatalogEntry(found)
	return 0
}

// splitCatalogShowArgs extracts --json and --catalog-dir (both `--flag value`
// and `--flag=value` forms) from any position, returning the remaining
// positional args.
func splitCatalogShowArgs(args []string) (string, bool, []string) {
	asJSON := false
	catalogDir := ""
	var rest []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			asJSON = true
		case a == "--catalog-dir" && i+1 < len(args):
			i++
			catalogDir = args[i]
		case strings.HasPrefix(a, "--catalog-dir="):
			catalogDir = strings.TrimPrefix(a, "--catalog-dir=")
		case a == "--json=true" || a == "--json=false":
			asJSON = strings.HasSuffix(a, "true")
		default:
			rest = append(rest, a)
		}
	}
	return catalogDir, asJSON, rest
}

// lookupCatalogEntry resolves id across every catalog kind in a fixed order:
// recipe, boilerplate, surface, capability, database-profile.
func lookupCatalogEntry(cat catalog.Catalog, id string) (string, any) {
	for _, r := range cat.Recipes {
		if r.ID == id {
			r := r
			return "recipe", &r
		}
	}
	for _, b := range cat.Boilerplates {
		if b.ID == id {
			b := b
			return "boilerplate", &b
		}
	}
	for _, s := range cat.Surfaces {
		if string(s.ID) == id {
			s := s
			return "surface", &s
		}
	}
	for _, c := range cat.Capabilities {
		if string(c.ID) == id {
			c := c
			return "capability", &c
		}
	}
	for _, p := range cat.DatabaseProfiles {
		if p.ID == id {
			p := p
			return "database-profile", &p
		}
	}
	return "", nil
}

func printCatalogEntry(entry any) {
	switch v := entry.(type) {
	case *domain.Recipe:
		fmt.Printf("  version: %s  status: %s\n", v.Version, v.Status)
		fmt.Printf("  description: %s\n", v.Description)
		fmt.Printf("  surfaces: %s\n", surfaceList(v.Provides.Surfaces))
		fmt.Printf("  capabilities: %s\n", capabilityList(v.Provides.Capabilities))
		fmt.Printf("  primary_boilerplates: %s\n", strings.Join(v.PrimaryBoilerplates, ", "))
		fmt.Printf("  database: default=%s allowed=%s shared_backend=%v\n",
			v.DatabasePolicy.DefaultProfile, strings.Join(v.DatabasePolicy.AllowedProfiles, ", "),
			v.DatabasePolicy.SharedBackend)
	case *domain.Boilerplate:
		fmt.Printf("  repo: %s  pin: %s\n", v.Repo, v.Pin)
		fmt.Printf("  adapter: %s  delivery: %s  decision: %s\n", v.Adapter, v.DeliveryStatus, v.DecisionStatus)
		spec := v.EffectiveSpec()
		fmt.Printf("  materialization: %s\n", spec.Strategy())
		if gen := spec.Generate; gen != nil {
			ids := make([]string, 0, len(gen.Profiles))
			for _, pr := range gen.Profiles {
				ids = append(ids, pr.ID)
			}
			fmt.Printf("  profiles: default=%s available=%s\n", gen.DefaultProfile, strings.Join(ids, ", "))
		}
		if strings.TrimSpace(v.Curation.Evidence) != "" {
			fmt.Printf("  curation: %s  evidence: %s\n", v.Curation.Status, v.Curation.Evidence)
		}
		fmt.Printf("  surfaces: %s\n", surfaceList(v.Provides.Surfaces))
		fmt.Printf("  capabilities: %s\n", capabilityList(v.Provides.Capabilities))
		fmt.Printf("  tech_tags: %s\n", strings.Join(v.TechTags, ", "))
	case *domain.Surface:
		fmt.Printf("  description: %s\n", v.Description)
	case *domain.Capability:
		fmt.Printf("  description: %s\n", v.Description)
	case *domain.DatabaseProfile:
		fmt.Printf("  engine: %s  managed: %v\n", v.Engine, v.Managed)
		fmt.Printf("  description: %s\n", v.Description)
	}
}

// validIDHint lists every known id grouped by kind so a miss is actionable.
func validIDHint(cat catalog.Catalog) string {
	var b strings.Builder
	b.WriteString("valid ids:\n")
	groups := []struct {
		kind string
		ids  []string
	}{
		{"recipes", idsOf(cat.Recipes, func(r domain.Recipe) string { return r.ID })},
		{"boilerplates", idsOf(cat.Boilerplates, func(x domain.Boilerplate) string { return x.ID })},
		{"surfaces", idsOf(cat.Surfaces, func(s domain.Surface) string { return string(s.ID) })},
		{"capabilities", idsOf(cat.Capabilities, func(c domain.Capability) string { return string(c.ID) })},
		{"database-profiles", idsOf(cat.DatabaseProfiles, func(p domain.DatabaseProfile) string { return p.ID })},
	}
	for _, g := range groups {
		sort.Strings(g.ids)
		fmt.Fprintf(&b, "  %s: %s\n", g.kind, strings.Join(g.ids, ", "))
	}
	return b.String()
}

func idsOf[T any](items []T, id func(T) string) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, id(it))
	}
	return out
}

func surfaceList(surfaces []domain.SurfaceID) string {
	out := make([]string, 0, len(surfaces))
	for _, s := range surfaces {
		out = append(out, string(s))
	}
	return strings.Join(out, ", ")
}

func capabilityList(caps []domain.CapabilityID) string {
	out := make([]string, 0, len(caps))
	for _, c := range caps {
		out = append(out, string(c))
	}
	return strings.Join(out, ", ")
}
