package service

import (
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	auth := NewAuth("admin", "secret", []byte("test-secret"))
	token := auth.Sign("admin", time.Now().Add(time.Hour))
	user, ok := auth.Read(token, time.Now())
	if !ok || user != "admin" {
		t.Fatalf("user=%q ok=%v", user, ok)
	}
	if _, ok := auth.Read(token+"x", time.Now()); ok {
		t.Fatal("tampered token accepted")
	}
	expired := auth.Sign("admin", time.Now().Add(-time.Minute))
	if _, ok := auth.Read(expired, time.Now()); ok {
		t.Fatal("expired token accepted")
	}
}

func TestCredentials(t *testing.T) {
	auth := NewAuth("admin", "secret", []byte("test-secret"))
	if !auth.Check("admin", "secret") {
		t.Fatal("expected match")
	}
	if auth.Check("admin", "wrong") || auth.Check("other", "secret") {
		t.Fatal("expected mismatch")
	}
	disabled := NewAuth("admin", "", []byte("test-secret"))
	if disabled.Check("admin", "") {
		t.Fatal("empty password must not log in")
	}
}
