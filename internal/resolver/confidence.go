package resolver

// Confidence thresholds for R9.
const (
	highScore  = 80
	highMargin = 12
	medScore   = 60
	medMargin  = 5
)

// ConfidenceLevel derives high/medium/low from winner score, margin vs #2
// and the count of unresolved dimensions.
func ConfidenceLevel(winner, margin, unresolved int) string {
	if unresolved > 0 {
		return "low"
	}
	if winner >= highScore && margin >= highMargin {
		return "high"
	}
	if winner >= medScore && margin >= medMargin {
		return "medium"
	}
	return "low"
}
