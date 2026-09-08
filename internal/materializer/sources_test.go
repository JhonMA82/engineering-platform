package materializer

import "testing"

// TestIsSHAPin guards the fetch-path split: full commit SHAs take the
// fetch-one-commit path while tags, short SHAs and empty pins keep the
// clone --branch path.
func TestIsSHAPin(t *testing.T) {
	cases := []struct {
		pin  string
		want bool
	}{
		{"5c449810b763140ac72133ff4ae63d8497cce77a", true},
		{"E6E5D3BDB7974D4A2283DF763EA2DD222D82E1F0", true},
		{"v1.0.0", false},
		{"main", false},
		{"5c44981", false},
		{"", false},
		{"5c449810b763140ac72133ff4ae63d8497cce77z", false},
		{"5c449810b763140ac72133ff4ae63d8497cce77a ", false},
	}
	for _, tc := range cases {
		t.Run(tc.pin, func(t *testing.T) {
			if got := isSHAPin(tc.pin); got != tc.want {
				t.Fatalf("isSHAPin(%q) = %v, want %v", tc.pin, got, tc.want)
			}
		})
	}
}
