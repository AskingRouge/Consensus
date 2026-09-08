package auth

import "testing"

func TestAllowed(t *testing.T) {
	allowlist := []string{"alice@unc.edu", "bob@unc.edu"}

	cases := []struct {
		email string
		want  bool
	}{
		{"alice@unc.edu", true},
		{"Alice@UNC.edu", true},   // case-insensitive
		{"  bob@unc.edu  ", true}, // trims whitespace
		{"eve@unc.edu", false},
		{"", false},
	}

	for _, c := range cases {
		if got := Allowed(c.email, allowlist); got != c.want {
			t.Errorf("Allowed(%q) = %v, want %v", c.email, got, c.want)
		}
	}
}

func TestAllowed_EmptyAllowlistDeniesEveryone(t *testing.T) {
	if Allowed("anyone@unc.edu", nil) {
		t.Error("an empty allowlist should deny everyone (fail closed)")
	}
}
