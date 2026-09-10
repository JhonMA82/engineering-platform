package catalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	catalogdata "github.com/jhonma82/engineering-platform/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
	"github.com/jhonma82/engineering-platform/internal/version"
)

// defaultDirCandidates resolves the in-repo catalog dir via relative path so
// both the CLI (repo root) and package tests find the same base data.
var defaultDirCandidates = []string{"catalog", "../catalog", "../../catalog"}

// SupportedSchemaVersion is the only catalog schema the core understands.
// Load refuses any catalog tree declaring another revision so catalog
// evolution never silently feeds unknown contracts into the engine (§30).
const SupportedSchemaVersion = 1

// Load returns the base catalog, optionally merged with an overlay
// directory: load(base) then overlay merge. The base comes from the
// in-repo ./catalog tree when one resolves from the working directory
// (dev checkouts and tests); otherwise the catalog embedded in the binary
// is the normal base source, so a globally installed eng works from any
// directory. --catalog-dir stays an optional overlay (or additional
// external catalog): it merges over the base and is never a mandatory
// replacement. The running binary (version.CoreVersion) must satisfy the
// merged metadata bounds; incompatible catalogs fail fast instead of
// feeding unknown contracts into the engine (§1.2).
func Load(overlayDir string) (Catalog, error) {
	return LoadWithCore(overlayDir, version.CoreVersion)
}

// checkCoreCompatible rejects a loaded catalog whose core bounds exclude
// the running binary. The "dev" bypass lives in version.CheckCompatibility
// and is documented there.
func checkCoreCompatible(c Catalog, core string) error {
	if err := version.CheckCompatibility(core, c.MinCoreVersion, c.MaxCoreVersion); err != nil {
		return domain.Catalog(err.Error())
	}
	return nil
}

// LoadWithCore is Load against an explicit core line, so tests can prove
// the compatibility gate without restamping the binary.
func LoadWithCore(overlayDir, core string) (Catalog, error) {
	base, err := loadBaseWithCore(core)
	if err != nil {
		return Catalog{}, err
	}
	if strings.TrimSpace(overlayDir) == "" {
		return base, nil
	}
	overlay, err := LoadDirWithCore(overlayDir, core)
	if err != nil {
		return Catalog{}, err
	}
	merged := MergeOverlay(base, overlay)
	if err := checkCoreCompatible(merged, core); err != nil {
		return Catalog{}, err
	}
	return merged, nil
}

func resolveBaseDir() string {
	dir, ok := findBaseDir()
	if !ok {
		return "catalog"
	}
	return dir
}

// findBaseDir reports the in-repo catalog dir when one resolves from the
// working directory, mirroring the historic probe order.
func findBaseDir() (string, bool) {
	for _, dir := range defaultDirCandidates {
		meta := filepath.Join(dir, "metadata.json")
		if st, err := os.Stat(meta); err == nil && !st.IsDir() {
			return dir, true
		}
	}
	return "", false
}

// loadBaseWithCore loads the base catalog from the resolved in-repo tree
// when one exists, else from the binary-embedded catalog.
func loadBaseWithCore(core string) (Catalog, error) {
	if dir, ok := findBaseDir(); ok {
		return LoadDirWithCore(dir, core)
	}
	return loadEmbeddedWithCore(core)
}

// loadEmbeddedWithCore loads the base catalog embedded in the binary.
func loadEmbeddedWithCore(core string) (Catalog, error) {
	return loadFromFS(catalogdata.FS, "embedded catalog", core)
}

// BaseDir reports the resolved in-repo catalog directory so CLI commands
// can resolve catalog-relative references (curation evidence) against the
// same tree Load reads.
func BaseDir() string {
	return resolveBaseDir()
}

type metadataFile struct {
	CatalogVersion string `json:"catalog_version"`
	MinCoreVersion string `json:"min_core_version"`
	MaxCoreVersion string `json:"max_core_version,omitempty"`
	SchemaVersion  int    `json:"schema_version"`
}

type aliasesFile struct {
	Aliases map[string]string `json:"aliases"`
	Notes   map[string]string `json:"notes"`
}

// LoadDir loads a catalog tree from a directory on disk, gated on the
// running binary line (version.CoreVersion).
func LoadDir(dir string) (Catalog, error) {
	return LoadDirWithCore(dir, version.CoreVersion)
}

// LoadDirWithCore is LoadDir against an explicit core line, so tests can
// prove the compatibility gate without restamping the binary.
func LoadDirWithCore(dir, core string) (Catalog, error) {
	return loadFromFS(os.DirFS(dir), dir, core)
}

// loadFromFS reads a catalog tree from fsys. show names the tree in error
// messages; for on-disk loads it is the directory, preserving the historic
// messages byte for byte. fsys paths always use slashes (path.Join) so the
// embedded tree resolves on every platform; show uses filepath.Join.
func loadFromFS(fsys fs.FS, show, core string) (Catalog, error) {
	var c Catalog
	raw, err := fs.ReadFile(fsys, "metadata.json")
	if err != nil {
		return Catalog{}, domain.Catalog(fmt.Sprintf("read metadata: %v", err))
	}
	var meta metadataFile
	if err := json.Unmarshal(raw, &meta); err != nil {
		return Catalog{}, domain.Catalog(fmt.Sprintf("parse metadata: %v", err))
	}
	c.CatalogVersion = meta.CatalogVersion
	c.MinCoreVersion = meta.MinCoreVersion
	c.MaxCoreVersion = meta.MaxCoreVersion
	c.SchemaVersion = meta.SchemaVersion
	if meta.SchemaVersion != SupportedSchemaVersion {
		return Catalog{}, domain.Catalog(fmt.Sprintf(
			"unsupported catalog schema version: %d", meta.SchemaVersion))
	}
	if err := checkCoreCompatible(c, core); err != nil {
		return Catalog{}, err
	}
	load := func(sub string, add func(raw json.RawMessage) error) error {
		entries, err := fs.ReadDir(fsys, sub)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		var names []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			f := filepath.Join(show, sub, name)
			raw, err := fs.ReadFile(fsys, path.Join(sub, name))
			if err != nil {
				return fmt.Errorf("read %s: %w", f, err)
			}
			if err := add(raw); err != nil {
				return fmt.Errorf("%s: %w", f, err)
			}
		}
		return nil
	}
	// Each file unmarshals into a fresh value: reusing one struct across
	// files lets encoding/json reuse backing arrays and corrupt earlier
	// snapshots.
	if err := load("surfaces", func(raw json.RawMessage) error {
		var v domain.Surface
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.Surfaces = append(c.Surfaces, v)
		return nil
	}); err != nil {
		return Catalog{}, domain.Catalog(err.Error())
	}
	if err := load("capabilities", func(raw json.RawMessage) error {
		var v domain.Capability
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.Capabilities = append(c.Capabilities, v)
		return nil
	}); err != nil {
		return Catalog{}, domain.Catalog(err.Error())
	}
	if err := load("recipes", func(raw json.RawMessage) error {
		var v domain.Recipe
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.Recipes = append(c.Recipes, v)
		return nil
	}); err != nil {
		return Catalog{}, domain.Catalog(err.Error())
	}
	if err := load("boilerplates", func(raw json.RawMessage) error {
		var v domain.Boilerplate
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.Boilerplates = append(c.Boilerplates, v)
		return nil
	}); err != nil {
		return Catalog{}, domain.Catalog(err.Error())
	}
	if err := load("database-profiles", func(raw json.RawMessage) error {
		var v domain.DatabaseProfile
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}
		c.DatabaseProfiles = append(c.DatabaseProfiles, v)
		return nil
	}); err != nil {
		return Catalog{}, domain.Catalog(err.Error())
	}
	raw, err = fs.ReadFile(fsys, path.Join("vocabulary", "aliases.json"))
	if err == nil {
		var af aliasesFile
		if err := json.Unmarshal(raw, &af); err != nil {
			return Catalog{}, domain.Catalog(fmt.Sprintf("parse aliases: %v", err))
		}
		c.Aliases = af.Aliases
		c.AliasNotes = af.Notes
	}
	if c.Aliases == nil {
		c.Aliases = map[string]string{}
	}
	return c, nil
}

// MergeOverlay returns base with overlay entries replacing by id; overlay
// aliases win over base aliases.
func MergeOverlay(base, overlay Catalog) Catalog {
	merged := base
	if overlay.CatalogVersion != "" {
		merged.CatalogVersion = overlay.CatalogVersion
	}
	if overlay.MinCoreVersion != "" {
		merged.MinCoreVersion = overlay.MinCoreVersion
	}
	if overlay.MaxCoreVersion != "" {
		merged.MaxCoreVersion = overlay.MaxCoreVersion
	}
	if overlay.SchemaVersion != 0 {
		merged.SchemaVersion = overlay.SchemaVersion
	}
	merged.Recipes = mergeRecipes(base.Recipes, overlay.Recipes)
	merged.Boilerplates = mergeBoilerplates(base.Boilerplates, overlay.Boilerplates)
	merged.Surfaces = mergeSurfaces(base.Surfaces, overlay.Surfaces)
	merged.Capabilities = mergeCapabilities(base.Capabilities, overlay.Capabilities)
	merged.DatabaseProfiles = mergeProfiles(base.DatabaseProfiles, overlay.DatabaseProfiles)
	if merged.Aliases == nil {
		merged.Aliases = map[string]string{}
	}
	for k, v := range overlay.Aliases {
		merged.Aliases[k] = v
	}
	return merged
}

func mergeRecipes(base, over []domain.Recipe) []domain.Recipe {
	byID := map[string]domain.Recipe{}
	var order []string
	for _, r := range base {
		if _, ok := byID[r.ID]; !ok {
			order = append(order, r.ID)
		}
		byID[r.ID] = r
	}
	for _, r := range over {
		if _, ok := byID[r.ID]; !ok {
			order = append(order, r.ID)
		}
		byID[r.ID] = r
	}
	out := make([]domain.Recipe, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

func mergeBoilerplates(base, over []domain.Boilerplate) []domain.Boilerplate {
	byID := map[string]domain.Boilerplate{}
	var order []string
	for _, b := range base {
		if _, ok := byID[b.ID]; !ok {
			order = append(order, b.ID)
		}
		byID[b.ID] = b
	}
	for _, b := range over {
		if _, ok := byID[b.ID]; !ok {
			order = append(order, b.ID)
		}
		byID[b.ID] = b
	}
	out := make([]domain.Boilerplate, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

func mergeSurfaces(base, over []domain.Surface) []domain.Surface {
	byID := map[domain.SurfaceID]domain.Surface{}
	var order []domain.SurfaceID
	put := func(s domain.Surface) {
		if _, ok := byID[s.ID]; !ok {
			order = append(order, s.ID)
		}
		byID[s.ID] = s
	}
	for _, s := range base {
		put(s)
	}
	for _, s := range over {
		put(s)
	}
	out := make([]domain.Surface, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

func mergeCapabilities(base, over []domain.Capability) []domain.Capability {
	byID := map[domain.CapabilityID]domain.Capability{}
	var order []domain.CapabilityID
	put := func(s domain.Capability) {
		if _, ok := byID[s.ID]; !ok {
			order = append(order, s.ID)
		}
		byID[s.ID] = s
	}
	for _, s := range base {
		put(s)
	}
	for _, s := range over {
		put(s)
	}
	out := make([]domain.Capability, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

func mergeProfiles(base, over []domain.DatabaseProfile) []domain.DatabaseProfile {
	byID := map[string]domain.DatabaseProfile{}
	var order []string
	for _, p := range base {
		if _, ok := byID[p.ID]; !ok {
			order = append(order, p.ID)
		}
		byID[p.ID] = p
	}
	for _, p := range over {
		if _, ok := byID[p.ID]; !ok {
			order = append(order, p.ID)
		}
		byID[p.ID] = p
	}
	out := make([]domain.DatabaseProfile, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}
