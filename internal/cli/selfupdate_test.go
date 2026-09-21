package cli

import "testing"

// TestSelfUpdateRejectsUsageErrors proves flag-level failures exit 2
// without any network: unknown flags and positional args never reach the
// release line.
func TestSelfUpdateRejectsUsageErrors(t *testing.T) {
	if got := runSelfUpdate([]string{"--nope"}); got != 2 {
		t.Fatalf("runSelfUpdate(--nope) = %d, want 2", got)
	}
	if got := runSelfUpdate([]string{"extra-arg"}); got != 2 {
		t.Fatalf("runSelfUpdate(extra-arg) = %d, want 2", got)
	}
}
