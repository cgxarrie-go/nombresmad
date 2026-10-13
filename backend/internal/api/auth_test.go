package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nombresmad/backend/internal/service"
)

func TestLoginReturnsBearerToken(t *testing.T) {
	auth := service.NewAuth("admin", "secret", []byte("test-secret"))
	srv := &Server{auth: auth}
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	rec := httptest.NewRecorder()
	srv.login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Username string `json:"username"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Username != "admin" || body.Token == "" {
		t.Fatalf("body=%+v", body)
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == legacySessionCookie && cookie.Value != "" && cookie.MaxAge >= 0 {
			t.Fatalf("login set a session cookie: %+v", cookie)
		}
	}

	next := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	next.Header.Set("Authorization", "Bearer "+body.Token)
	session := httptest.NewRecorder()
	srv.session(session, next)
	if session.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", session.Code, session.Body.String())
	}

	cookieOnly := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	cookieOnly.AddCookie(&http.Cookie{Name: legacySessionCookie, Value: body.Token})
	if _, ok := srv.sessionUser(cookieOnly); ok {
		t.Fatal("cookie must not authenticate")
	}

	missing := httptest.NewRecorder()
	if srv.requireAuth(missing, httptest.NewRequest(http.MethodPost, "/api/items", nil)) {
		t.Fatal("missing token was accepted")
	}
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", missing.Code)
	}

	bad := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"no"}`))
	denied := httptest.NewRecorder()
	srv.login(denied, bad)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", denied.Code)
	}
}
