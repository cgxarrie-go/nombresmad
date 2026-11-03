package api

import (
	"context"
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
	if notEmail.Code != http.StatusUnauthorized {
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

type apiGoogleVerifier struct{}

func (apiGoogleVerifier) Verify(_ context.Context, credential, audience string) (service.GoogleIdentity, error) {
	if credential != "valid-credential" || audience != "web-client-id" {
		return service.GoogleIdentity{}, os.ErrInvalid
	}
	return service.GoogleIdentity{Email: "admin@example.com", EmailVerified: true}, nil
}

type apiGoogleIdentityVerifier struct {
	identity service.GoogleIdentity
}

func (v apiGoogleIdentityVerifier) Verify(_ context.Context, credential, audience string) (service.GoogleIdentity, error) {
	if credential != "valid-credential" || audience != "web-client-id" {
		return service.GoogleIdentity{}, os.ErrInvalid
	}
	return v.identity, nil
}

func TestGoogleLoginAndRequestValidation(t *testing.T) {
	auth := service.NewAuth(nil, "", []byte("test-secret"), "web-client-id", []string{"admin@example.com"}, apiGoogleVerifier{})
	srv := &Server{auth: auth}
	google := postJSON(t, srv.loginGoogle, "/api/login/google", `{"credential":"valid-credential"}`, "")
	if google.Username != "admin@example.com" || google.Token == "" || google.MustChangePassword {
		t.Fatalf("body=%+v", google)
	}

	sessionReq := requestWithToken(http.MethodGet, "/api/session", google.Token)
	sessionRec := httptest.NewRecorder()
	srv.session(sessionRec, sessionReq)
	if sessionRec.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionRec.Code, sessionRec.Body.String())
	}

	unauthorizedAuth := service.NewAuth(nil, "", []byte("test-secret"), "web-client-id", []string{"admin@example.com"}, apiGoogleIdentityVerifier{
		identity: service.GoogleIdentity{Email: "outside@example.com", EmailVerified: true},
	})
	unauthorizedGoogle := httptest.NewRecorder()
	(&Server{auth: unauthorizedAuth}).loginGoogle(unauthorizedGoogle, httptest.NewRequest(http.MethodPost, "/api/login/google", strings.NewReader(`{"credential":"valid-credential"}`)))
	if unauthorizedGoogle.Code != http.StatusUnauthorized || strings.TrimSpace(unauthorizedGoogle.Body.String()) != "No autorizado" {
		t.Fatalf("unauthorized Google status=%d body=%q", unauthorizedGoogle.Code, unauthorizedGoogle.Body.String())
	}

	badPassword := httptest.NewRecorder()
	srv.login(badPassword, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin@example.com","password":"wrong"}`)))
	notAllowedGoogle := httptest.NewRecorder()
	srv.loginGoogle(notAllowedGoogle, httptest.NewRequest(http.MethodPost, "/api/login/google", strings.NewReader(`{"credential":"not-allowed"}`)))
	if badPassword.Code != http.StatusUnauthorized || notAllowedGoogle.Code != http.StatusUnauthorized {
		t.Fatalf("password status=%d Google status=%d", badPassword.Code, notAllowedGoogle.Code)
	}
	if strings.TrimSpace(badPassword.Body.String()) != strings.TrimSpace(notAllowedGoogle.Body.String()) {
		t.Fatalf("unauthorized errors disclose different details: password=%q Google=%q", badPassword.Body.String(), notAllowedGoogle.Body.String())
	}

	for _, handler := range []http.HandlerFunc{srv.login, srv.loginGoogle} {
		malformed := httptest.NewRecorder()
		handler(malformed, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{`)))
		if malformed.Code != http.StatusBadRequest {
			t.Fatalf("malformed status=%d body=%s", malformed.Code, malformed.Body.String())
		}
		oversized := httptest.NewRecorder()
		handler(oversized, httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"credential":"`+strings.Repeat("x", maxLoginBodyBytes)+`"}`)))
		if oversized.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("oversized status=%d body=%s", oversized.Code, oversized.Body.String())
		}
	}
}

func TestExpiredSessionAndLogout(t *testing.T) {
	auth := newTestAuth(t, "secret")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	srv := &Server{auth: auth}
	expired := auth.Sign("admin@example.com", 1, time.Now().Add(-time.Minute))
	rec := httptest.NewRecorder()
	srv.session(rec, requestWithToken(http.MethodGet, "/api/session", expired))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired session status=%d body=%s", rec.Code, rec.Body.String())
	}

	logout := httptest.NewRecorder()
	srv.logout(logout, httptest.NewRequest(http.MethodPost, "/api/logout", nil))
	if logout.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logout.Code, logout.Body.String())
	}
	cookies := logout.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != legacySessionCookie || cookies[0].MaxAge >= 0 {
		t.Fatalf("legacy cookie was not cleared: %+v", cookies)
	}
}

func newTestAuth(t *testing.T, defaultPassword string) *service.Auth {
	t.Helper()
	return service.NewAuth(service.NewMemoryAdminStore(), defaultPassword, []byte("test-secret"), "", []string{"admin@example.com", "bea@example.com"}, nil)
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
