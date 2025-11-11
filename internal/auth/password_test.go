package auth

import "testing"

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if !CheckPassword(hash, "secret123") {
		t.Fatalf("expected password to match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatalf("expected password not to match")
	}
}


