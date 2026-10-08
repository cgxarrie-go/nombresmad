package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nombresmad/backend/internal/service"
)

func TestLoginSetsCookie(t *testing.T) {
	auth := service.NewAuth("admin", "secret", []byte("test-secret"))
	srv := &Server{auth: auth}
	body := strings.NewReader(`{"username":"admin","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/login", body)
	rec := httptest.NewRecorder()
	srv.login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookie || !cookies[0].HttpOnly {
		t.Fatalf("cookies=%v", cookies)
	}
	next := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	next.AddCookie(cookies[0])
	if user, ok := srv.sessionUser(next); !ok || user != "admin" {
		t.Fatalf("user=%q ok=%v", user, ok)
	}
	bad := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"admin","password":"no"}`))
	denied := httptest.NewRecorder()
	srv.login(denied, bad)
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", denied.Code)
	}
}
