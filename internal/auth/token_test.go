package auth

import (
	"fmt"
	"hash/crc32"
	"strings"
	"testing"
)

// A hash produced by Laravel's Hash::make('rahasia123') (bcrypt, cost 12).
const laravelHash = "$2y$12$n/.LKWeKfX/eQfFgE02.IuXGo8oUTUiwab6o1668ax/rAZ7IQtQlK"

func TestCheckLaravelHash(t *testing.T) {
	if !CheckPassword("rahasia123", laravelHash) {
		t.Fatal("Laravel bcrypt hash should verify")
	}
	if CheckPassword("salah", laravelHash) {
		t.Fatal("wrong password verified")
	}
}

func TestHashPasswordUses2y(t *testing.T) {
	h, err := (&Service{BcryptCost: 4}).HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$2y$04$") || !CheckPassword("secret", h) {
		t.Fatalf("unexpected hash %q", h)
	}
}

func TestTokenFormat(t *testing.T) {
	tok := GenerateTokenString("")
	if len(tok) != 48 {
		t.Fatalf("len %d", len(tok))
	}
	if want := fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(tok[:40]))); tok[40:] != want {
		t.Fatalf("crc mismatch")
	}
	// sha256 of "test" — what Sanctum stores in personal_access_tokens.token
	if HashToken("test") != "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08" {
		t.Fatal("sha256 mismatch")
	}
}
