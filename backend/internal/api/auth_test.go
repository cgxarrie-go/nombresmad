package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"nombresmad/backend/internal/service"
)

func TestMain(m *testing.M) {
	service.PasswordIterations = 1000
	os.Exit(m.Run())
}

func TestLoginReturnsBearerToken(t *testing.T) {
	auth := newTestAuth(t, "secret")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	srv := &Server{auth: auth}
	body := postJSON(t, srv.login, "/api/login", `{"username":"admin@example.com","password":"secret"}`, "")
	if body.Username != "admin@example.com" || body.Token == "" || !body.MustChangePassword {
		t.Fatalf("body=%+v", body)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+body.Token)
	sessionRec := httptest.NewRecorder()
	srv.session(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionRec.Code, sessionRec.Body.String())
	}
	var session authResponse
	if err := json.NewDecoder(sessionRec.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if session.Username != "admin@example.com" || !session.MustChangePassword {
		t.Fatalf("session=%+v", session)
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

	blocked := httptest.NewRecorder()
	gated := httptest.NewRequest(http.MethodPost, "/api/items", nil)
	gated.Header.Set("Authorization", "Bearer "+body.Token)
	if srv.requireAuth(blocked, gated) {
		t.Fatal("password change was not required")
	}
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", blocked.Code, blocked.Body.String())
	}

	changed := postJSON(t, srv.changePassword, "/api/account/password", `{"currentPassword":"secret","newPassword":"replacement-1"}`, body.Token)
	if changed.MustChangePassword || changed.Token == "" {
		t.Fatalf("changed=%+v", changed)
	}
	if _, ok := srv.sessionUser(requestWithToken(http.MethodGet, "/api/session", body.Token)); ok {
		t.Fatal("token from before the password change was accepted")
	}
	allowed := httptest.NewRecorder()
	if !srv.requireAuth(allowed, requestWithToken(http.MethodPost, "/api/items", changed.Token)) {
		t.Fatalf("status=%d body=%s", allowed.Code, allowed.Body.String())
	}

	denied := httptest.NewRecorder()
	srv.login(denied, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin@example.com","password":"no"}`)))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", denied.Code)
	}

	notEmail := httptest.NewRecorder()
	srv.login(notEmail, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"secret"}`)))
	if notEmail.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", notEmail.Code, notEmail.Body.String())
	}
}

func TestAdminResetForcesPasswordChange(t *testing.T) {
	auth := newTestAuth(t, "secret")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ChangePassword("admin@example.com", "secret", "replacement-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	srv := &Server{auth: auth}
	admin := postJSON(t, srv.login, "/api/login", `{"username":"admin@example.com","password":"replacement-1"}`, "")
	rejected := httptest.NewRecorder()
	rejectedReq := httptest.NewRequest(http.MethodPost, "/api/admins", strings.NewReader(`{"username":"bea"}`))
	rejectedReq.Header.Set("Authorization", "Bearer "+admin.Token)
	srv.admins(rejected, rejectedReq)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("create status=%d body=%s", rejected.Code, rejected.Body.String())
	}
	created := postJSON(t, srv.admins, "/api/admins", `{"username":"bea@example.com"}`, admin.Token)
	if created.Username != "bea@example.com" || !created.MustChangePassword {
		t.Fatalf("created=%+v", created)
	}

	bea := postJSON(t, srv.login, "/api/login", `{"username":"bea@example.com","password":"secret"}`, "")
	if !bea.MustChangePassword {
		t.Fatal("new admin was not required to change the password")
	}
	chosen := postJSON(t, srv.changePassword, "/api/account/password", `{"currentPassword":"secret","newPassword":"bea-private"}`, bea.Token)

	self := httptest.NewRecorder()
	selfReq := httptest.NewRequest(http.MethodPost, "/api/admins/reset", strings.NewReader(`{"username":"admin@example.com"}`))
	selfReq.Header.Set("Authorization", "Bearer "+admin.Token)
	srv.resetAdminPassword(self, selfReq)
	if self.Code != http.StatusBadRequest {
		t.Fatalf("self reset status=%d body=%s", self.Code, self.Body.String())
	}

	reset := httptest.NewRecorder()
	resetReq := httptest.NewRequest(http.MethodPost, "/api/admins/reset", strings.NewReader(`{"username":"bea@example.com"}`))
	resetReq.Header.Set("Authorization", "Bearer "+admin.Token)
	srv.resetAdminPassword(reset, resetReq)
	if reset.Code != http.StatusOK {
		t.Fatalf("reset status=%d body=%s", reset.Code, reset.Body.String())
	}
	if _, ok := srv.sessionUser(requestWithToken(http.MethodGet, "/api/session", chosen.Token)); ok {
		t.Fatal("reset admin kept the previous session")
	}
	again := postJSON(t, srv.login, "/api/login", `{"username":"bea@example.com","password":"secret"}`, "")
	if !again.MustChangePassword {
		t.Fatal("reset did not force a password change")
	}
	blocked := httptest.NewRecorder()
	if srv.requireAuth(blocked, requestWithToken(http.MethodPost, "/api/admins", again.Token)) {
		t.Fatal("admin who must change the password was allowed to manage accounts")
	}
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("status=%d", blocked.Code)
	}
}

func newTestAuth(t *testing.T, defaultPassword string) *service.Auth {
	t.Helper()
	return service.NewAuth(service.NewMemoryAdminStore(), defaultPassword, []byte("test-secret"))
}

func requestWithToken(method, path, token string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func postJSON(t *testing.T, handler http.HandlerFunc, path, payload, token string) authResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(payload))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == legacySessionCookie && cookie.Value != "" && cookie.MaxAge >= 0 {
			t.Fatalf("set a session cookie: %+v", cookie)
		}
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("%s status=%d body=%s", path, rec.Code, rec.Body.String())
	}
	var body authResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}
