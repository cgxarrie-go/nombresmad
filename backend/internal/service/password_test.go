package service

import (
	"encoding/hex"
	"testing"
)

func TestPBKDF2SHA256Vector(t *testing.T) {
	// RFC 6070-style vectors published for PBKDF2-HMAC-SHA256.
	cases := []struct {
		iterations int
		want       string
	}{
		{1, "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"},
		{2, "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"},
		{4096, "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a"},
	}
	for _, tc := range cases {
		got := hex.EncodeToString(pbkdf2SHA256([]byte("password"), []byte("salt"), tc.iterations, 32))
		if got != tc.want {
			t.Fatalf("iterations=%d got=%s", tc.iterations, got)
		}
	}
}

func TestHashPasswordIsOneWay(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "secret" || len(hash) < 20 {
		t.Fatalf("hash=%q", hash)
	}
	if !CheckPassword(hash, "secret") {
		t.Fatal("expected match")
	}
	if CheckPassword(hash, "secret2") || CheckPassword("not-a-hash", "secret") || CheckPassword(hash, "") {
		t.Fatal("expected mismatch")
	}
}
