package auth

import "testing"

func TestSignAndVerify(t *testing.T) {
	token, err := newToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	signed := sign(token, "secret-1")
	got, ok := verify(signed, "secret-1")
	if !ok {
		t.Fatal("verify rejected a validly signed token")
	}
	if got != token {
		t.Errorf("verify returned %q, want %q", got, token)
	}
}

func TestVerify_WrongSecretRejected(t *testing.T) {
	token, _ := newToken()
	signed := sign(token, "secret-1")

	if _, ok := verify(signed, "secret-2"); ok {
		t.Error("verify accepted a token signed with a different secret")
	}
}

func TestVerify_TamperedTokenRejected(t *testing.T) {
	token, _ := newToken()
	signed := sign(token, "secret-1")

	tampered := "not-the-real-token" + signed[len(token):]
	if _, ok := verify(tampered, "secret-1"); ok {
		t.Error("verify accepted a tampered cookie value")
	}
}

func TestVerify_MalformedRejected(t *testing.T) {
	cases := []string{"", "no-dot-here", ".", "."}
	for _, c := range cases {
		if _, ok := verify(c, "secret-1"); ok {
			t.Errorf("verify accepted malformed input %q", c)
		}
	}
}

func TestNewToken_Unique(t *testing.T) {
	a, _ := newToken()
	b, _ := newToken()
	if a == b {
		t.Error("two calls to newToken produced the same value")
	}
}
