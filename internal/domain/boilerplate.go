package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Catalog-driven operation vocabulary (§71). An adapter combines these
// generic operations; adding a boilerplate that uses them needs no core
// release. "generate" covers generator CLIs that produce their own output
// directory (H4) instead of shipping a copyable tree.
var validAdapterOperations = map[string]bool{
	"fetch": true, "copy": true, "prune": true, "template": true, "compose": true,
	"generate": true,
}

// Materialization strategies: the only two behaviors the core
// understands. "copy" fetches a pinned source tree and copies it;
// "generate" acquires a pinned generator/factory, executes its curated
// argv command in an isolated workspace and keeps only the declared
// output. Provider-specific names never appear here.
const (
	StrategyCopy     = "copy"
	StrategyGenerate = "generate"
)

// ValidGeneratorPlaceholders is the closed placeholder vocabulary for
// generator templates. {name} is the logical surface name, {project} the
// normalized project name, {surface} the surface id, {profile} the
// resolved generation profile and {output} the controlled sandbox output
// location resolved at materialization time (never serialized in plans).
var ValidGeneratorPlaceholders = map[string]bool{
	"name": true, "project": true, "surface": true,
	"profile": true, "output": true,
}

// placeholderPattern matches {token} placeholders in generator templates.
var placeholderPattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// GeneratorPlaceholdersUsed returns the distinct placeholder tokens in
// template order.
func GeneratorPlaceholdersUsed(template string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range placeholderPattern.FindAllStringSubmatch(template, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// ValidateGeneratorPlaceholders rejects unknown {tokens} in one template
// value. Known placeholders are always accepted here; empty-value
// rejection happens at substitution time when runtime values are known.
func ValidateGeneratorPlaceholders(field, template string) error {
	for _, token := range GeneratorPlaceholdersUsed(template) {
		if !ValidGeneratorPlaceholders[token] {
			return Validation(fmt.Sprintf("generator %s uses unknown placeholder {%s} (want {name|project|surface|profile|output})", field, token))
		}
	}
	return nil
}

// SubstituteGeneratorPlaceholder resolves every placeholder in one value.
// Unknown tokens fail; known tokens with a missing or empty value fail so
// a curated command can never run with a silently dropped argument.
func SubstituteGeneratorPlaceholder(template string, values map[string]string) (string, error) {
	var fail error
	resolved := placeholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		if fail != nil {
			return match
		}
		token := match[1 : len(match)-1]
		if !ValidGeneratorPlaceholders[token] {
			fail = Validation(fmt.Sprintf("generator template uses unknown placeholder {%s} (want {name|project|surface|profile|output})", token))
			return match
		}
		value, ok := values[token]
		if !ok || len(value) == 0 {
			fail = Validation(fmt.Sprintf("generator placeholder {%s} has no value", token))
			return match
		}
		return value
	})
	if fail != nil {
		return "", fail
	}
	return resolved, nil
}

// SubstituteGeneratorPlaceholders resolves placeholders element-wise so
// every argv token stays a separate argument (never a shell string).
func SubstituteGeneratorPlaceholders(argv []string, values map[string]string) ([]string, error) {
	out := make([]string, len(argv))
	for i, arg := range argv {
		resolved, err := SubstituteGeneratorPlaceholder(arg, values)
		if err != nil {
			return nil, err
		}
		out[i] = resolved
	}
	return out, nil
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

// ProfileRequirements declares the structured decision signals a
// generation profile needs. Capabilities and surfaces match the
// resolver-derived signals (never free text); requirement_ids match stable
// product-requirement IDs recorded in the decision reasons as
// "includes product feature: <id>" and only optimize an already-selected
// foundation — an unmet requirement never disqualifies the foundation, it
// falls back to a smaller profile. UnlessCapabilities excludes a profile
// when a capability is present (e.g. a public-access capability excludes
// an authenticated-only profile). All lists are exact lowercase ids.
type ProfileRequirements struct {
	Capabilities       []string `json:"capabilities,omitempty"`
	Surfaces           []string `json:"surfaces,omitempty"`
	RequirementIDs     []string `json:"requirement_ids,omitempty"`
	UnlessCapabilities []string `json:"unless_capabilities,omitempty"`
}

// Validate checks the requirements shape structurally.
func (r ProfileRequirements) Validate(profileID string) error {
	for _, v := range append(append(append(append([]string{}, r.Capabilities...), r.Surfaces...), r.RequirementIDs...), r.UnlessCapabilities...) {
		if len(v) == 0 {
			return Validation(fmt.Sprintf("generator profile %q has a blank requirement entry", profileID))
		}
	}
	return nil
}

// GeneratorProfile is one curated generation variant of a factory. The
// core never interprets what an id means (minimal, authenticated, ... are
// foundation concepts); it only selects deterministically and passes the
// declared arguments. Arguments are non-runtime argv fragments:
// temporary paths are never recorded here.
type GeneratorProfile struct {
	ID        string               `json:"id"`
	Arguments []string             `json:"arguments,omitempty"`
	Requires  *ProfileRequirements `json:"requires,omitempty"`
}

// Validate checks the profile shape and its placeholder vocabulary.
func (p GeneratorProfile) Validate() error {
	if len(p.ID) == 0 {
		return Validation("generator profile id is required")
	}
	for _, arg := range p.Arguments {
		if len(arg) == 0 {
			return Validation(fmt.Sprintf("generator profile %q has an empty argument", p.ID))
		}
		if err := ValidateGeneratorPlaceholders("profile "+p.ID+" argument", arg); err != nil {
			return err
		}
	}
	if p.Requires != nil {
		if err := p.Requires.Validate(p.ID); err != nil {
			return err
		}
	}
	return nil
}

// GenerateSpec is the declarative form of the generic "generate"
// operation. Some foundations are not copyable trees but generator
// factories that scaffold a fresh directory. Prepare runs dependency
// steps inside the factory workspace (never the final destination); Run
// is argv-only — never a shell string; Output is the path the command
// must produce. Run, Prepare and Output accept the closed placeholder
// vocabulary ({name}, {project}, {surface}, {profile}, {output}); every
// other byte is literal. Profiles declare the curated variants;
// DefaultProfile is the smallest valid one.
type GenerateSpec struct {
	Prepare        []AdapterCommand   `json:"prepare,omitempty"`
	Run            AdapterCommand     `json:"run"`
	Output         string             `json:"output"`
	DefaultProfile string             `json:"default_profile,omitempty"`
	Profiles       []GeneratorProfile `json:"profiles,omitempty"`
}

// UnmarshalJSON accepts {run, output} where run is any AdapterCommand
// shape (argv array or {run|command+args} object).
func (g *GenerateSpec) UnmarshalJSON(raw []byte) error {
	var obj struct {
		Prepare        []json.RawMessage  `json:"prepare"`
		Run            json.RawMessage    `json:"run"`
		Output         string             `json:"output"`
		DefaultProfile string             `json:"default_profile"`
		Profiles       []GeneratorProfile `json:"profiles"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("generate spec must be an object with run and output: %s", trimJSON(raw))
	}
	if len(obj.Run) == 0 {
		return fmt.Errorf("generate spec requires run (argv array or {run|command+args} object)")
	}
	var run AdapterCommand
	if err := json.Unmarshal(obj.Run, &run); err != nil {
		return fmt.Errorf("generate run: %v", err)
	}
	g.Run = run
	g.Output = obj.Output
	g.DefaultProfile = obj.DefaultProfile
	g.Profiles = obj.Profiles
	for _, rawPrepare := range obj.Prepare {
		var cmd AdapterCommand
		if err := json.Unmarshal(rawPrepare, &cmd); err != nil {
			return fmt.Errorf("generate prepare: %v", err)
		}
		g.Prepare = append(g.Prepare, cmd)
	}
	return nil
}

// Validate checks the argv shapes structurally, the closed placeholder
// vocabulary and the profile contract. Shell/metacharacter rejection runs
// at execution (materializer) and pre-validation, mirroring
// setup/checks. Output allows placeholders: each token is validated by
// substituting a dummy segment so {output} and friends stay confined to
// clean relative paths.
func (g GenerateSpec) Validate() error {
	for _, cmd := range g.Prepare {
		if err := cmd.Validate(); err != nil {
			return err
		}
		for _, arg := range cmd.Run {
			if err := ValidateGeneratorPlaceholders("prepare argument", arg); err != nil {
				return err
			}
		}
	}
	if err := g.Run.Validate(); err != nil {
		return err
	}
	for _, arg := range g.Run.Run {
		if err := ValidateGeneratorPlaceholders("run argument", arg); err != nil {
			return err
		}
	}
	if err := ValidateGeneratorPlaceholders("output", g.Output); err != nil {
		return err
	}
	dummy, err := SubstituteGeneratorPlaceholder(g.Output, map[string]string{
		"name": "x", "project": "x", "surface": "x", "profile": "x", "output": "x",
	})
	if err != nil {
		return err
	}
	if err := checkRelativePath("generate output", dummy); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, pr := range g.Profiles {
		if err := pr.Validate(); err != nil {
			return err
		}
		if seen[pr.ID] {
			return Validation(fmt.Sprintf("duplicate generator profile id %q", pr.ID))
		}
		seen[pr.ID] = true
	}
	if len(g.Profiles) == 0 {
		if len(g.DefaultProfile) != 0 {
			return Validation("generate default_profile requires at least one profile")
		}
		return nil
	}
	if len(g.DefaultProfile) == 0 {
		return Validation("generate profiles require a default_profile (smallest valid profile)")
	}
	if !seen[g.DefaultProfile] {
		return Validation(fmt.Sprintf("generate default_profile %q matches no declared profile", g.DefaultProfile))
	}
	return nil
}

// AdapterFingerprint is the deterministic sha256 hex over the executable
// adapter contract: strategy, operations, prepare commands, run template,
// output declaration, profile definitions, setup, checks, managed files
// and prune paths. A plan records it so an old plan never silently
// executes a modified adapter even when provider and pin are unchanged.
func AdapterFingerprint(spec AdapterSpec) string {
	h := sha256.New()
	write := func(v string) {
		h.Write([]byte(v))
		h.Write([]byte{0})
	}
	write(spec.Strategy())
	for _, op := range spec.Operations {
		write(op)
	}
	if spec.Generate != nil {
		for _, cmd := range spec.Generate.Prepare {
			write(joinArgs(cmd.Run))
		}
		write(joinArgs(spec.Generate.Run.Run))
		write(spec.Generate.Output)
		write(spec.Generate.DefaultProfile)
		for _, pr := range spec.Generate.Profiles {
			write(pr.ID)
			write(joinArgs(pr.Arguments))
			if pr.Requires != nil {
				caps := append([]string{}, pr.Requires.Capabilities...)
				sort.Strings(caps)
				surfaces := append([]string{}, pr.Requires.Surfaces...)
				sort.Strings(surfaces)
				ids := append([]string{}, pr.Requires.RequirementIDs...)
				sort.Strings(ids)
				unless := append([]string{}, pr.Requires.UnlessCapabilities...)
				sort.Strings(unless)
				for _, v := range append(append(append(caps, surfaces...), ids...), unless...) {
					write(v)
				}
			}
		}
	}
	for _, cmd := range spec.Setup {
		write(joinArgs(cmd.Run))
	}
	for _, cmd := range spec.Checks {
		write(joinArgs(cmd.Run))
	}
	for _, entry := range spec.PrunePaths {
		write(entry)
	}
	for _, entry := range spec.ManagedFiles {
		write(entry)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// joinArgs flattens one argv command for fingerprinting.
func joinArgs(argv []string) string {
	return strings.Join(argv, "\x1f")
}

// Provenance records where a catalog entry came from without digging
// through commit history (§12). Historical statuses preserve the legacy
// migration semantics; the live decision_status/delivery_status fields stay
// honest for the v1 gates (see Validate and the curation contract), and any
// intentional downgrade is documented in docs/decisions/migration-notes.md.
type Provenance struct {
	Source                   string `json:"source,omitempty"`
	LegacyCatalogVersion     string `json:"legacy_catalog_version,omitempty"`
	LegacyID                 string `json:"legacy_id,omitempty"`
	HistoricalDecisionStatus string `json:"historical_decision_status,omitempty"`
	HistoricalDeliveryStatus string `json:"historical_delivery_status,omitempty"`
	MigrationReason          string `json:"migration_reason,omitempty"`
}

// Curation is the formal curation link (§3.1). Status reuses the delivery
// status vocabulary — there is a single status axis, so a set status must
// equal the boilerplate delivery_status (checked by catalog validation).
// Evidence is a relative path confined to the catalog directory
// (e.g. "curation/tanstack-admin.md").
type Curation struct {
	Status   string `json:"status,omitempty"`
	Evidence string `json:"evidence,omitempty"`
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
	Generate     *GenerateSpec    `json:"generate,omitempty"`
}

// Strategy reports the materialization behavior of an adapter: generate
// when the generic generate operation is declared (a Generate spec is then
// required), copy otherwise. Legacy string-only adapters default to copy.
func (s AdapterSpec) Strategy() string {
	for _, op := range s.Operations {
		if op == "generate" {
			return StrategyGenerate
		}
	}
	if s.Generate != nil {
		return StrategyGenerate
	}
	return StrategyCopy
}

// Validate enforces the operation vocabulary, relative safe paths and
// well-formed commands.
func (s AdapterSpec) Validate() error {
	seen := map[string]bool{}
	for _, op := range s.Operations {
		if !validAdapterOperations[op] {
			return Validation(fmt.Sprintf("unknown adapter operation %q (want fetch|copy|prune|template|compose|generate)", op))
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
	hasGenerate := false
	for _, op := range s.Operations {
		if op == "generate" {
			hasGenerate = true
		}
	}
	if hasGenerate && s.Generate == nil {
		return Validation("adapter declares generate operation but no generate spec")
	}
	if s.Generate != nil {
		if !hasGenerate {
			return Validation("adapter declares a generate spec but operations lacks generate")
		}
		if err := s.Generate.Validate(); err != nil {
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

// Technology is the catalog-known technical metadata of a boilerplate,
// keyed by constraint target (§2.5). Values are honest catalog data: only
// list what the curation evidence or the boilerplate contract supports;
// unknown languages or runtimes are omitted, never invented.
type Technology struct {
	Framework []string `json:"framework,omitempty"`
	Language  []string `json:"language,omitempty"`
	Runtime   []string `json:"runtime,omitempty"`
}

// Values returns the declared values for a constraint target, or nil for
// targets without boilerplate-level metadata (database, deployment,
// provider resolve through profiles, deployment knowledge and the
// boilerplate identity respectively).
func (t Technology) Values(target string) []string {
	switch target {
	case "framework":
		return t.Framework
	case "language":
		return t.Language
	case "runtime":
		return t.Runtime
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
	Curation         Curation     `json:"curation,omitempty"`
	Provides         Provides     `json:"provides,omitempty"`
	TechTags         []string     `json:"tech_tags,omitempty"`
	Technology       Technology   `json:"technology,omitempty"`
	IncludedFeatures []string     `json:"included_features,omitempty"`
	UpdateStrategy   string       `json:"update_strategy,omitempty"`
	Provenance       Provenance   `json:"provenance,omitempty"`
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
	Curation         *Curation       `json:"curation,omitempty"`
	Provides         Provides        `json:"provides"`
	TechTags         []string        `json:"tech_tags"`
	Technology       Technology      `json:"technology"`
	IncludedFeatures []string        `json:"included_features"`
	UpdateStrategy   string          `json:"update_strategy"`
	Provenance       *Provenance     `json:"provenance,omitempty"`
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
	if w.Curation != nil {
		b.Curation = *w.Curation
	}
	b.Provides = w.Provides
	b.TechTags = w.TechTags
	b.Technology = w.Technology
	b.IncludedFeatures = w.IncludedFeatures
	b.UpdateStrategy = w.UpdateStrategy
	if w.Provenance != nil {
		b.Provenance = *w.Provenance
	}
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
		Technology:       b.Technology,
		IncludedFeatures: b.IncludedFeatures,
		UpdateStrategy:   b.UpdateStrategy,
	}
	if b.Provenance != (Provenance{}) {
		prov := b.Provenance
		w.Provenance = &prov
	}
	if b.Curation != (Curation{}) {
		w.Curation = &b.Curation
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

// MaterializationStrategy reports copy or generate for a boilerplate
// from its effective adapter. It is the only strategy signal the core
// uses; provider identities never branch behavior.
func (b Boilerplate) MaterializationStrategy() string {
	return b.EffectiveSpec().Strategy()
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
// sources accept local (path) or git (repo). Entries with delivery_status
// catalog-only are catalog knowledge, not materializable foundations
// (§5.2-5.3): they carry no pin and may declare no adapter, and the
// composer eligibility gate already excludes them (no pin/adapter and a
// non-available delivery state), so validation stays honest without
// forcing invented pins or fake adapters.
func (b Boilerplate) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return Validation("boilerplate id is required")
	}
	catalogOnly := strings.TrimSpace(b.DeliveryStatus) == "catalog-only"
	if !catalogOnly && strings.TrimSpace(b.Pin) == "" {
		return Validation("boilerplate pin is required")
	}
	if !catalogOnly && strings.TrimSpace(b.Adapter) == "" && b.AdapterSpec == nil {
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
