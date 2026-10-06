package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	sessionSecret = []byte("test-secret")
	token := signSession("admin", time.Now().Add(time.Hour))
	user, ok := readSession(token, time.Now())
	if !ok || user != "admin" {
		t.Fatalf("user=%q ok=%v", user, ok)
	}
	if _, ok := readSession(token+"x", time.Now()); ok {
		t.Fatal("tampered token accepted")
	}
	expired := signSession("admin", time.Now().Add(-time.Minute))
	if _, ok := readSession(expired, time.Now()); ok {
		t.Fatal("expired token accepted")
	}
}

func TestCredentials(t *testing.T) {
	adminUser, adminPass = "admin", "secret"
	if !credentialsMatch("admin", "secret") {
		t.Fatal("expected match")
	}
	if credentialsMatch("admin", "wrong") || credentialsMatch("other", "secret") {
		t.Fatal("expected mismatch")
	}
	adminPass = ""
	if credentialsMatch("admin", "") {
		t.Fatal("empty password must not log in")
	}
}

func TestLoginSetsCookie(t *testing.T) {
	adminUser, adminPass = "admin", "secret"
	sessionSecret = []byte("test-secret")
	body := strings.NewReader(`{"username":"admin","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	rec := httptest.NewRecorder()
	loginHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookie || !cookies[0].HttpOnly {
		t.Fatalf("cookies=%v", cookies)
	}
	next := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	next.AddCookie(cookies[0])
	if user, ok := sessionUser(next); !ok || user != "admin" {
		t.Fatalf("user=%q ok=%v", user, ok)
	}
	bad := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"no"}`))
	denied := httptest.NewRecorder()
	loginHandler(denied, bad)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", denied.Code)
	}
}
