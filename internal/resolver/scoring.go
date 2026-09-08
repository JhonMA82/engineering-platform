package resolver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jhonma82/engineering-platform/internal/catalog"
	"github.com/jhonma82/engineering-platform/internal/domain"
)

// Scoring weights sum to 100; the product-feature bonus is a tie-breaker <= 3.
const (
	wSurface    = 35
	wArchCap    = 20
	wData       = 15
	wOps        = 10
	wCuration   = 8
	wSimplicity = 7
	wPreference = 5
	maxBonus    = 3
)

// Scored is an eligibility verdict plus its R8 score breakdown.
type Scored struct {
	Recipe   string
	Version  string
	Eligible bool
	Score    int
	Positive []string
	Negative []string
}

// ScoreAll runs R8 over every candidate; only eligible ones receive scores.
func ScoreAll(n Normalized, sp Split, derived []domain.DerivedRequirement, elig []Eligibility, idx catalog.Index) []Scored {
	out := make([]Scored, 0, len(elig))
	for _, e := range elig {
		s := Scored{
			Recipe:   e.Recipe.ID,
			Version:  e.Recipe.Version,
			Eligible: e.Eligible,
			Positive: append([]string{}, e.Positive...),
			Negative: append([]string{}, e.Negative...),
		}
		if e.Eligible {
			s.Score, s.Positive = scoreEligible(n, sp, derived, e, idx)
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Recipe < out[j].Recipe
	})
	return out
}

func scoreEligible(n Normalized, sp Split, derived []domain.DerivedRequirement, e Eligibility, idx catalog.Index) (int, []string) {
	pos := append([]string{}, e.Positive...)
	total := wSurface + wArchCap
	data, dataReason := dataFit(n, derived, e.Recipe)
	total += data
	pos = append(pos, dataReason)
	ops, opsReason := opsFit(n, derived, e.Recipe)
	total += ops
	pos = append(pos, opsReason)
	cur := curation(e.Recipe)
	total += cur
	pos = append(pos, fmt.Sprintf("curation: recipe status %s", e.Recipe.Status))
	simp := simplicity(e.Recipe)
	total += simp
	if simp == wSimplicity {
		pos = append(pos, "composition simplicity: single-surface foundation")
	}
	pref, prefReason := preferenceFit(sp, e.Recipe)
	total += pref
	if prefReason != "" {
		pos = append(pos, prefReason)
	}
	bonus, bonusReasons := productBonus(n, e.Recipe, idx)
	total += bonus
	pos = append(pos, bonusReasons...)
	return total, pos
}

func needsShared(n Normalized, derived []domain.DerivedRequirement) bool {
	if n.Intent.Data.MultiUser || n.Intent.Data.Persistence == "shared" {
		return true
	}
	for _, d := range derived {
		if d.ID == "shared-backend" {
			return true
		}
	}
	return false
}

func dataFit(n Normalized, derived []domain.DerivedRequirement, r domain.Recipe) (int, string) {
	pol := r.DatabasePolicy
	name := pol.DefaultProfile
	if name == "" {
		name = "default"
	}
	if needsShared(n, derived) {
		if pol.SharedBackend {
			return wData, "database fit: shared backend on " + name
		}
		return 6, "database fit: single-node policy cannot share across clients"
	}
	if strings.ToLower(n.Intent.Data.Persistence) == "local" {
		for _, p := range pol.AllowedProfiles {
			if p == "sqlite-local" {
				return wData, "database fit: local sqlite profile"
			}
		}
		return 9, "database fit: no local profile allowed, needs managed"
	}
	return 12, "database fit: default profile " + name
}

func opsFit(n Normalized, derived []domain.DerivedRequirement, r domain.Recipe) (int, string) {
	needRealtime := n.Intent.Ops.Realtime
	needOffline := n.Intent.Ops.OfflineOperation || n.Intent.Data.OfflineSync
	for _, d := range derived {
		if d.ID == "realtime" {
			needRealtime = true
		}
		if d.ID == "offline-operation" {
			needOffline = true
		}
	}
	if !needRealtime && !needOffline {
		return 8, "operational fit: no special runtime needs"
	}
	ok := true
	if needRealtime && !providesCapability(r, "realtime") {
		ok = false
	}
	if needOffline && !providesCapability(r, "offline-operation") {
		ok = false
	}
	if ok {
		return wOps, "operational fit: realtime/offline needs covered"
	}
	return 4, "operational fit: runtime need partially covered"
}

func curation(r domain.Recipe) int {
	switch r.Status {
	case "stable":
		return wCuration
	case "active":
		return 6
	default:
		return 4
	}
}

func simplicity(r domain.Recipe) int {
	v := wSimplicity - 2*(len(r.Provides.Surfaces)-1)
	if v < 0 {
		return 0
	}
	return v
}

func preferenceFit(sp Split, r domain.Recipe) (int, string) {
	for _, p := range sp.Preferences {
		v := strings.ToLower(strings.TrimSpace(p.Value))
		switch p.Kind {
		case "prefer":
			if techMatch(r.TechTags, v) {
				return wPreference, "preference match: " + v
			}
		case "avoid":
			if techMatch(r.TechTags, v) {
				return 1, "preference tension: avoids " + v
			}
		}
	}
	return 3, ""
}

func productBonus(n Normalized, r domain.Recipe, idx catalog.Index) (int, []string) {
	features := map[string]bool{}
	for _, id := range r.PrimaryBoilerplates {
		b, ok := idx.Boilerplate(id)
		if !ok {
			continue
		}
		for _, f := range b.IncludedFeatures {
			features[strings.ToLower(f)] = true
		}
	}
	bonus := 0
	var reasons []string
	matched := map[string]bool{}
	consider := func(text string) {
		t := strings.ToLower(text)
		for f := range features {
			if f == "" || matched[f] {
				continue
			}
			if t == f || strings.Contains(t, f) || strings.Contains(f, t) {
				matched[f] = true
				bonus++
				reasons = append(reasons, fmt.Sprintf("includes product feature: %s (+1 tie-break)", f))
			}
		}
	}
	for _, p := range n.Intent.ProductRequirements {
		consider(p.ID)
		consider(p.Description)
	}
	if bonus > maxBonus {
		bonus = maxBonus
		reasons = reasons[:maxBonus]
	}
	return bonus, reasons
}
