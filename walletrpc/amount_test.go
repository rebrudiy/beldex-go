package walletrpc

import "testing"

func TestFormatParseBDX(t *testing.T) {
	cases := map[string]uint64{
		"0":           0,
		"1":           AtomicPerBDX,
		"1.5":         1_500_000_000,
		"0.000000001": 1,
		"12345.6789":  12_345_678_900_000,
	}
	for s, atomic := range cases {
		if got := FormatBDX(atomic); got != s {
			t.Errorf("FormatBDX(%d) = %q, want %q", atomic, got, s)
		}
		got, err := ParseBDX(s)
		if err != nil || got != atomic {
			t.Errorf("ParseBDX(%q) = %d, %v; want %d", s, got, err, atomic)
		}
	}
	for _, bad := range []string{"", "abc", "-1", "0.0000000001"} {
		if _, err := ParseBDX(bad); err == nil {
			t.Errorf("ParseBDX(%q) succeeded, want error", bad)
		}
	}
}
