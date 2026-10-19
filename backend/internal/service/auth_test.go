package service

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"nombresmad/backend/internal/model"
)

func TestMain(m *testing.M) {
	PasswordIterations = 1000
	os.Exit(m.Run())
}

func TestSessionRoundTrip(t *testing.T) {
	auth := NewAuth(nil, "", []byte("test-secret"))
	token := auth.Sign("admin@example.com", 1, time.Now().Add(time.Hour))
	user, version, ok := auth.Read(token, time.Now())
	if !ok || user != "admin@example.com" || version != 1 {
		t.Fatalf("user=%q version=%d ok=%v", user, version, ok)
	}
	if _, _, ok := auth.Read(token+"x", time.Now()); ok {
		t.Fatal("tampered token accepted")
	}
	expired := auth.Sign("admin@example.com", 1, time.Now().Add(-time.Minute))
	if _, _, ok := auth.Read(expired, time.Now()); ok {
		t.Fatal("expired token accepted")
	}
	if _, ok := auth.Session(token, time.Now()); ok {
		t.Fatal("token without an account was accepted")
	}
}

func TestCredentials(t *testing.T) {
	auth := testAuth(t, "secret")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if !auth.Check("admin@example.com", "secret") {
		t.Fatal("expected match")
	}
	if auth.Check("admin@example.com", "wrong") || auth.Check("other", "secret") || auth.Check("admin@example.com", "") {
		t.Fatal("expected mismatch")
	}
	disabled := NewAuth(NewMemoryAdminStore(), "", []byte("test-secret"))
	if err := disabled.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if disabled.Check("admin@example.com", "") || disabled.Check("admin@example.com", "secret") {
		t.Fatal("empty default password must not create a login")
	}
}

func TestPasswordIsStoredHashedAndFirstLoginMustChange(t *testing.T) {
	store := NewMemoryAdminStore()
	auth := NewAuth(store, "secret", []byte("test-secret"))
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	rec, err := store.ByUsername("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if rec.PasswordHash == "secret" || !strings.HasPrefix(rec.PasswordHash, "pbkdf2-sha256$") {
		t.Fatalf("hash=%q", rec.PasswordHash)
	}
	if !rec.MustChangePassword {
		t.Fatal("first account must change the password")
	}
	if _, ok := auth.Session(auth.Sign("admin@example.com", rec.CredentialVersion, time.Now().Add(time.Hour)), time.Now()); !ok {
		t.Fatal("expected session")
	}
}

func TestChangePasswordClearsFlagAndInvalidatesOldToken(t *testing.T) {
	auth := testAuth(t, "default-pass")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	old, session, ok := auth.Login("admin@example.com", "default-pass", time.Now().Add(time.Hour))
	if !ok || !session.MustChangePassword {
		t.Fatalf("session=%+v ok=%v", session, ok)
	}
	if _, err := auth.ChangePassword("admin@example.com", "nope", "replacement-1", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("wrong current password was accepted")
	}
	if _, err := auth.ChangePassword("admin@example.com", "default-pass", "short", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("short password was accepted")
	}
	if _, err := auth.ChangePassword("admin@example.com", "default-pass", "default-pass", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("default password was accepted as the new password")
	}
	next, err := auth.ChangePassword("admin@example.com", "default-pass", "replacement-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.Session(old, time.Now()); ok {
		t.Fatal("old token still valid")
	}
	got, ok := auth.Session(next, time.Now())
	if !ok || got.MustChangePassword || got.Username != "admin@example.com" {
		t.Fatalf("session=%+v ok=%v", got, ok)
	}
	if auth.Check("admin@example.com", "default-pass") || !auth.Check("admin@example.com", "replacement-1") {
		t.Fatal("password was not replaced")
	}
	if _, err := auth.ChangePassword("admin@example.com", "replacement-1", "default-pass", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("changing back to the default password was accepted")
	}
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if !auth.Check("admin@example.com", "replacement-1") {
		t.Fatal("startup must not reset an existing password")
	}
}

func TestAdminCanResetAnotherAdminToDefault(t *testing.T) {
	auth := testAuth(t, "secret")
	if err := auth.EnsureDefaultAdmin("admin@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.ChangePassword("admin@example.com", "secret", "replacement-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	created, err := auth.Create("bea@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !created.MustChangePassword {
		t.Fatal("created account must change the password")
	}
	if _, err := auth.Create("bea@example.com"); !errors.Is(err, model.ErrExists) {
		t.Fatalf("duplicate err=%v", err)
	}
	if _, err := auth.Create("bad name"); err == nil {
		t.Fatal("username with a space was accepted")
	}
	token, _, ok := auth.Login("bea@example.com", "secret", time.Now().Add(time.Hour))
	if !ok {
		t.Fatal("expected default password to log in")
	}
	if _, err := auth.ChangePassword("bea@example.com", "secret", "bea-private", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := auth.ResetPassword("admin@example.com", "admin@example.com"); err == nil {
		t.Fatal("self reset was accepted")
	}
	if err := auth.ResetPassword("admin@example.com", "missing@example.com"); err == nil {
		t.Fatal("missing user was accepted")
	}
	if err := auth.ResetPassword("admin@example.com", "bea@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.Session(token, time.Now()); ok {
		t.Fatal("reset account kept its session")
	}
	_, session, ok := auth.Login("bea@example.com", "secret", time.Now().Add(time.Hour))
	if !ok || !session.MustChangePassword {
		t.Fatalf("session=%+v ok=%v", session, ok)
	}
	if auth.Check("bea@example.com", "bea-private") {
		t.Fatal("old password still works")
	}
	rows, err := auth.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Username != "admin@example.com" || rows[1].Username != "bea@example.com" || !rows[1].MustChangePassword {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestNonEmailAdminIsRejected(t *testing.T) {
	store := NewMemoryAdminStore()
	auth := NewAuth(store, "secret", []byte("test-secret"))
	if err := store.Insert("admin", "not-a-hash", false); err != nil {
		t.Fatal(err)
	}
	if auth.Check("admin", "secret") || auth.Check("not-an-email", "secret") {
		t.Fatal("non-email account was admitted")
	}
	if err := auth.EnsureDefaultAdmin("admin"); err == nil {
		t.Fatal("non-email default user was accepted")
	}
	if err := auth.EnsureDefaultAdmin("Admin@Example.com"); err != nil {
		t.Fatal(err)
	}
	_, session, ok := auth.Login("ADMIN@example.com", "secret", time.Now().Add(time.Hour))
	if !ok || session.Username != "admin@example.com" {
		t.Fatalf("session=%+v ok=%v", session, ok)
	}
	if _, err := auth.Create("bea"); err == nil {
		t.Fatal("non-email create was accepted")
	}
	if _, err := auth.Create("Admin@Example.com"); !errors.Is(err, model.ErrExists) {
		t.Fatalf("duplicate email err=%v", err)
	}
	rows, err := auth.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Username != "admin@example.com" {
		t.Fatalf("rows=%+v", rows)
	}
	legacy := auth.Sign("admin", 1, time.Now().Add(time.Hour))
	if _, ok := auth.Session(legacy, time.Now()); ok {
		t.Fatal("non-email session was admitted")
	}
}

func testAuth(t *testing.T, defaultPassword string) *Auth {
	t.Helper()
	return NewAuth(NewMemoryAdminStore(), defaultPassword, []byte("test-secret"))
}
