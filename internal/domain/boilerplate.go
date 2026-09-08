package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Catalog-driven operation vocabulary (§71). An adapter combines these
// generic operations; adding a boilerplate that uses them needs no core
// release.
var validAdapterOperations = map[string]bool{
	"fetch": true, "copy": true, "prune": true, "template": true, "compose": true,
}

// AdapterCommand is one curated executable declaration. It is always argv,
// never a shell string: the materializer executes it without a shell, so
// shell metacharacters are rejected instead of interpreted.
type AdapterCommand struct {
	Run []string `json:"run"`
}

// UnmarshalJSON accepts an argv array (["npm", "run", "build"]) or an
// object with a run/command shape ({"run": [...]} or
// {"command": "npm", "args": [...]}).
func (c *AdapterCommand) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		return fmt.Errorf("adapter command must not be null")
	}
	var argv []string
	if err := json.Unmarshal(raw, &argv); err == nil {
		c.Run = argv
		return nil
	}
	var obj struct {
		Run     []string `json:"run"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("adapter command must be an argv array or {run|command+args} object: %s", trimJSON(raw))
	}
	if obj.Run != nil {
		c.Run = obj.Run
		return nil
	}
	if obj.Command != "" {
		c.Run = append([]string{obj.Command}, obj.Args...)
		return nil
	}
	return fmt.Errorf("adapter command object declares no run or command: %s", trimJSON(raw))
}

// MarshalJSON emits the canonical {run: [...]} object form.
func (c AdapterCommand) MarshalJSON() ([]byte, error) {
	type wire struct {
		Run []string `json:"run"`
	}
	return json.Marshal(wire{Run: c.Run})
}

// Validate checks argv shape. Shell/metacharacter rejection lives in the
// materializer process boundary (fail-safe at execution); this stays
// structural so the catalog validator catches malformed adapters early.
func (c AdapterCommand) Validate() error {
	if len(c.Run) == 0 {
		return Validation("adapter command run must not be empty")
	}
	for _, arg := range c.Run {
		if strings.TrimSpace(arg) == "" {
			return Validation("adapter command argument must not be empty")
		}
	}
	return nil
}

// AdapterSpec is the declarative object form of a boilerplate adapter. The
// legacy plain-string form (adapter name only) remains valid and yields a
// nil spec with default fetch+copy semantics.
type AdapterSpec struct {
	Name         string           `json:"name,omitempty"`
	Operations   []string         `json:"operations,omitempty"`
	PrunePaths   []string         `json:"prune_paths,omitempty"`
	Setup        []AdapterCommand `json:"setup,omitempty"`
	Checks       []AdapterCommand `json:"checks,omitempty"`
	ManagedFiles []string         `json:"managed_files,omitempty"`
}

// Validate enforces the operation vocabulary, relative safe paths and
// well-formed commands.
func (s AdapterSpec) Validate() error {
	seen := map[string]bool{}
	for _, op := range s.Operations {
		if !validAdapterOperations[op] {
			return Validation(fmt.Sprintf("unknown adapter operation %q (want fetch|copy|prune|template|compose)", op))
		}
		if seen[op] {
			return Validation(fmt.Sprintf("duplicate adapter operation %q", op))
		}
		seen[op] = true
	}
	for _, p := range s.PrunePaths {
		if err := checkRelativePath("prune_paths", p); err != nil {
			return err
		}
	}
	for _, p := range s.ManagedFiles {
		if err := checkRelativePath("managed_files", p); err != nil {
			return err
		}
	}
	for _, cmd := range s.Setup {
		if err := cmd.Validate(); err != nil {
			return err
		}
	}
	for _, cmd := range s.Checks {
		if err := cmd.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// checkRelativePath rejects absolute paths, ".." escapes and backslashes so
// adapter-declared paths can never address files outside a destination.
func checkRelativePath(field, p string) error {
	if strings.TrimSpace(p) == "" {
		return Validation(fmt.Sprintf("adapter %s entry must not be empty", field))
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return Validation(fmt.Sprintf("adapter %s entry %q must be a relative slash path", field, p))
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return Validation(fmt.Sprintf("adapter %s entry %q escapes its destination", field, p))
		}
		if strings.TrimSpace(seg) == "" {
			return Validation(fmt.Sprintf("adapter %s entry %q is not clean", field, p))
		}
	}
	return nil
}

// SourceSpec declares where a boilerplate is fetched from. Empty type keeps
// the legacy behavior: a git source derived from the boilerplate repo field.
// Local sources exist for fixtures and tests and verify their pin against a
// PIN file at the source root (offline analog of the git tag match).
type SourceSpec struct {
	Type string `json:"type,omitempty"`
	Path string `json:"path,omitempty"`
	Repo string `json:"repo,omitempty"`
}

// Kind resolves the effective source kind: "local" or "git".
func (s SourceSpec) Kind() string {
	if s.Type == "local" {
		return "local"
	}
	return "git"
}

// Validate checks the source shape against the legacy repo fallback.
func (s SourceSpec) Validate(repo string) error {
	switch s.Type {
	case "", "git":
		eff := s.Repo
		if strings.TrimSpace(eff) == "" {
			eff = repo
		}
		if strings.TrimSpace(eff) == "" {
			return Validation("boilerplate repo is required for git sources")
		}
	case "local":
		if strings.TrimSpace(s.Path) == "" {
			return Validation("boilerplate local source path is required")
		}
	default:
		return Validation(fmt.Sprintf("unknown source type %q (want local or git)", s.Type))
	}
	return nil
}

// Boilerplate is a foundation, not a feature bundle. Feature coverage only
// feeds a small scoring tie-breaker, never eligibility.
type Boilerplate struct {
	ID               string       `json:"-"`
	Repo             string       `json:"-"`
	Pin              string       `json:"-"`
	Adapter          string       `json:"-"`
	AdapterSpec      *AdapterSpec `json:"-"`
	Source           SourceSpec   `json:"-"`
	DeliveryStatus   string       `json:"delivery_status,omitempty"`
	DecisionStatus   string       `json:"decision_status,omitempty"`
	Provides         Provides     `json:"provides,omitempty"`
	TechTags         []string     `json:"tech_tags,omitempty"`
	IncludedFeatures []string     `json:"included_features,omitempty"`
	UpdateStrategy   string       `json:"update_strategy,omitempty"`
}

// boilerplateWire is the JSON shape: adapter accepts a plain string (legacy
// name) or a declarative object; source is optional.
type boilerplateWire struct {
	ID               string          `json:"id"`
	Repo             string          `json:"repo"`
	Pin              string          `json:"pin"`
	Adapter          json.RawMessage `json:"adapter"`
	Source           *SourceSpec     `json:"source"`
	DeliveryStatus   string          `json:"delivery_status"`
	DecisionStatus   string          `json:"decision_status"`
	Provides         Provides        `json:"provides"`
	TechTags         []string        `json:"tech_tags"`
	IncludedFeatures []string        `json:"included_features"`
	UpdateStrategy   string          `json:"update_strategy"`
}

// UnmarshalJSON parses both adapter forms. An object without an explicit
// name falls back to the boilerplate id so the composer eligibility
// invariant (adapter present) and tech-signal matching keep working on the
// untouched pure packages.
func (b *Boilerplate) UnmarshalJSON(raw []byte) error {
	var w boilerplateWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return err
	}
	b.ID = w.ID
	b.Repo = w.Repo
	b.Pin = w.Pin
	b.DeliveryStatus = w.DeliveryStatus
	b.DecisionStatus = w.DecisionStatus
	b.Provides = w.Provides
	b.TechTags = w.TechTags
	b.IncludedFeatures = w.IncludedFeatures
	b.UpdateStrategy = w.UpdateStrategy
	b.AdapterSpec = nil
	b.Adapter = ""
	if w.Source != nil {
		b.Source = *w.Source
	}
	if len(w.Adapter) == 0 || string(w.Adapter) == "null" {
		return nil
	}
	var name string
	if err := json.Unmarshal(w.Adapter, &name); err == nil {
		b.Adapter = name
		return nil
	}
	var spec AdapterSpec
	if err := json.Unmarshal(w.Adapter, &spec); err != nil {
		return fmt.Errorf("boilerplate %q: invalid adapter (want string or object): %s", w.ID, trimJSON(w.Adapter))
	}
	b.AdapterSpec = &spec
	b.Adapter = spec.Name
	if b.Adapter == "" {
		b.Adapter = w.ID
	}
	return nil
}

// MarshalJSON round-trips the adapter as an object when a spec is present
// and as a plain string otherwise.
func (b Boilerplate) MarshalJSON() ([]byte, error) {
	w := boilerplateWire{
		ID:               b.ID,
		Repo:             b.Repo,
		Pin:              b.Pin,
		DeliveryStatus:   b.DeliveryStatus,
		DecisionStatus:   b.DecisionStatus,
		Provides:         b.Provides,
		TechTags:         b.TechTags,
		IncludedFeatures: b.IncludedFeatures,
		UpdateStrategy:   b.UpdateStrategy,
	}
	if b.AdapterSpec != nil {
		raw, err := json.Marshal(b.AdapterSpec)
		if err != nil {
			return nil, err
		}
		w.Adapter = raw
	} else if b.Adapter != "" {
		raw, err := json.Marshal(b.Adapter)
		if err != nil {
			return nil, err
		}
		w.Adapter = raw
	}
	if b.Source.Type != "" || b.Source.Path != "" || b.Source.Repo != "" {
		src := b.Source
		w.Source = &src
	}
	return json.Marshal(w)
}

// EffectiveUpdateStrategy returns the declared strategy, defaulting to
// manual when the entry declares none. eng update reports this value; it
// never auto-applies anything in v1.
func (b Boilerplate) EffectiveUpdateStrategy() string {
	if b.UpdateStrategy == "" {
		return "manual"
	}
	return b.UpdateStrategy
}

// EffectiveRepo resolves the git URL: source-level override wins over the
// legacy repo field.
func (b Boilerplate) EffectiveRepo() string {
	if strings.TrimSpace(b.Source.Repo) != "" {
		return b.Source.Repo
	}
	return b.Repo
}

// EffectiveSpec returns the declarative adapter, defaulting legacy
// string-only adapters to fetch+copy semantics.
func (b Boilerplate) EffectiveSpec() AdapterSpec {
	if b.AdapterSpec != nil {
		return *b.AdapterSpec
	}
	return AdapterSpec{Name: b.Adapter, Operations: []string{"fetch", "copy"}}
}

// Validate checks pin/adapter/source presence required by the catalog
// validator. Adapter accepts the legacy string or the object form;
// sources accept local (path) or git (repo).
func (b Boilerplate) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return Validation("boilerplate id is required")
	}
	if strings.TrimSpace(b.Pin) == "" {
		return Validation("boilerplate pin is required")
	}
	if strings.TrimSpace(b.Adapter) == "" && b.AdapterSpec == nil {
		return Validation("boilerplate adapter is required")
	}
	if b.AdapterSpec != nil {
		if err := b.AdapterSpec.Validate(); err != nil {
			return err
		}
	}
	if err := b.Source.Validate(b.Repo); err != nil {
		return err
	}
	if b.UpdateStrategy != "" && !validUpdateStrategies[b.UpdateStrategy] {
		return Validation(fmt.Sprintf("unknown update_strategy %q (want replace|merge-seed|fork-track|manual)", b.UpdateStrategy))
	}
	return nil
}

// UpdateStrategy vocabulary (§52): how a materialized component tracks its
// upstream foundation. Empty means the entry declares no strategy and
// evolution reports it as manual.
var validUpdateStrategies = map[string]bool{
	"replace": true, "merge-seed": true, "fork-track": true, "manual": true,
}

func trimJSON(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if len(s) > 80 {
		return s[:80] + "…"
	}
	return s
}
