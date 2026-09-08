package version

import "testing"

func TestParseVersionForms(t *testing.T) {
	cases := []struct {
		in   string
		want [3]int
	}{
		{"1.0.0", [3]int{1, 0, 0}},
		{"v1.0.0", [3]int{1, 0, 0}},
		{"V2.10.3", [3]int{2, 10, 3}},
		{"1.0", [3]int{1, 0, 0}},
		{"v1.2", [3]int{1, 2, 0}},
		{"1.2.3-rc.1", [3]int{1, 2, 3}},
		{"1.2.3+build.5", [3]int{1, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseVersion(tc.in)
			if err != nil {
				t.Fatalf("parse %q: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("parse %q = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseVersionRejects(t *testing.T) {
	for _, in := range []string{"", "dev", "1", "1.0.0.0", "a.b.c", "1.x", "1..0"} {
		t.Run("reject-"+in, func(t *testing.T) {
			if _, err := ParseVersion(in); err == nil {
				t.Fatalf("parse %q succeeded, want error", in)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.0", "1.0.0", 0},
		{"1.0.0", "1.1.0", -1},
		{"1.2.0", "1.0.0", 1},
		{"2.0.0", "10.0.0", -1},
		{"2026.09.07", "2026.09.07", 0},
	}
	for _, tc := range cases {
		t.Run(tc.a+"-vs-"+tc.b, func(t *testing.T) {
			got, err := CompareVersions(tc.a, tc.b)
			if err != nil {
				t.Fatalf("compare: %v", err)
			}
			if got != tc.want {
				t.Fatalf("compare(%q,%q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestCheckCompatibility(t *testing.T) {
	t.Run("dev bypasses gates", func(t *testing.T) {
		if err := CheckCompatibility("dev", "9.9.9", ""); err != nil {
			t.Fatalf("dev should bypass: %v", err)
		}
	})
	t.Run("equal and greater accepted", func(t *testing.T) {
		for _, core := range []string{"1.0.0", "1.2.0", "v1.0.0", "2.0"} {
			if err := CheckCompatibility(core, "1.0.0", ""); err != nil {
				t.Fatalf("core %q rejected: %v", core, err)
			}
		}
	})
	t.Run("older core rejected with exact vocabulary", func(t *testing.T) {
		err := CheckCompatibility("1.0.0", "1.1.0", "")
		if err == nil {
			t.Fatal("expected rejection, got nil")
		}
		want := "catalog requires core >= 1.1.0, running core is 1.0.0"
		if err.Error() != want {
			t.Fatalf("error = %q, want %q", err.Error(), want)
		}
	})
	t.Run("newer core rejected by max bound", func(t *testing.T) {
		err := CheckCompatibility("2.1.0", "1.0.0", "2.0.0")
		if err == nil {
			t.Fatal("expected max rejection, got nil")
		}
		want := "catalog requires core <= 2.0.0, running core is 2.1.0"
		if err.Error() != want {
			t.Fatalf("error = %q, want %q", err.Error(), want)
		}
	})
}
