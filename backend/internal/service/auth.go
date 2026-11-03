package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

const (
	passwordAuthMethod = "password"
	googleAuthMethod   = "google"
)

// Session is an authenticated administrator.
type Session struct {
	Username           string
	MustChangePassword bool
}

// GoogleIdentity contains the verified identity claims required for sign-in.
type GoogleIdentity struct {
	Email         string
	EmailVerified bool
}

// GoogleTokenVerifier verifies an ID token for the configured OAuth client.
type GoogleTokenVerifier interface {
	Verify(ctx context.Context, credential, audience string) (GoogleIdentity, error)
}

type googleIDTokenVerifier struct{}

func (googleIDTokenVerifier) Verify(ctx context.Context, credential, audience string) (GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, credential, audience)
	if err != nil {
		return GoogleIdentity{}, err
	}
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	return GoogleIdentity{Email: email, EmailVerified: emailVerified}, nil
}

// Auth checks administrator credentials and signs session tokens.
type Auth struct {
	users            AdminStore
	defaultPassword  string
	secret           []byte
	dummyHash        string
	googleClientID   string
	authorizedEmails map[string]struct{}
	googleVerifier   GoogleTokenVerifier
}

// NewAuth uses the account store, fallback credentials, session secret, Google configuration, and authorized emails.
func NewAuth(users AdminStore, defaultPassword string, secret []byte, googleClientID string, authorizedEmails []string, verifier GoogleTokenVerifier) *Auth {
	dummy, err := HashPassword(randomToken())
	if err != nil {
		dummy = ""
	}
	if verifier == nil {
		verifier = googleIDTokenVerifier{}
	}
	allowed := make(map[string]struct{}, len(authorizedEmails))
	for _, address := range authorizedEmails {
		if email, ok := NormalizeEmail(address); ok {
			allowed[email] = struct{}{}
		}
	}
	return &Auth{
		users:            users,
		defaultPassword:  defaultPassword,
		secret:           secret,
		dummyHash:        dummy,
		googleClientID:   googleClientID,
		authorizedEmails: allowed,
		googleVerifier:   verifier,
	}
}

// Check reports whether the supplied credentials match a stored account.
func (a *Auth) Check(username, password string) bool {
	_, _, _, ok := a.authenticate(username, password)
	return ok
}

func (a *Auth) authenticate(username, password string) (storedUsername string, version int, mustChange bool, ok bool) {
	email, emailOK := NormalizeEmail(username)
	if a == nil || !emailOK || !a.isAuthorized(email) || password == "" || len(password) > 1024 {
		return "", 0, false, false
	}
	username = email
	hash := a.dummyHash
	var recOK bool
	if a.users != nil {
		rec, err := a.users.ByUsername(username)
		if err == nil && rec.PasswordHash != "" {
			hash = rec.PasswordHash
			storedUsername = rec.Username
			version = rec.CredentialVersion
			mustChange = rec.MustChangePassword
			recOK = true
		}
	}
	if hash == "" || !CheckPassword(hash, password) || !recOK {
		return "", 0, false, false
	}
	return storedUsername, version, mustChange, true
}

// Login checks the fallback password credentials and returns a session token.
func (a *Auth) Login(username, password string, expires time.Time) (string, Session, bool) {
	user, version, mustChange, ok := a.authenticate(username, password)
	if !ok {
		return "", Session{}, false
	}
	return a.sign(passwordAuthMethod, user, version, expires), Session{Username: user, MustChangePassword: mustChange}, true
}

// LoginGoogle verifies a Google ID token and returns a session token for its verified email.
func (a *Auth) LoginGoogle(credential string, expires time.Time) (string, Session, bool) {
	if a == nil || a.googleVerifier == nil || a.googleClientID == "" || credential == "" {
		return "", Session{}, false
	}
	identity, err := a.googleVerifier.Verify(context.Background(), credential, a.googleClientID)
	if err != nil || !identity.EmailVerified {
		return "", Session{}, false
	}
	email, ok := NormalizeEmail(identity.Email)
	if !ok || !a.isAuthorized(email) {
		return "", Session{}, false
	}
	return a.sign(googleAuthMethod, email, 0, expires), Session{Username: email}, true
}

// Session validates a bearer token against the current account state.
func (a *Auth) Session(token string, now time.Time) (Session, bool) {
	principal, version, method, ok := a.read(token, now)
	if !ok || a == nil {
		return Session{}, false
	}
	switch method {
	case googleAuthMethod:
		email, emailOK := NormalizeEmail(principal)
		if !emailOK || email != principal || !a.isAuthorized(email) {
			return Session{}, false
		}
		return Session{Username: principal}, true
	case passwordAuthMethod:
		if a.users == nil {
			return Session{}, false
		}
		email, emailOK := NormalizeEmail(principal)
		if !emailOK || !a.isAuthorized(email) {
			return Session{}, false
		}
		rec, err := a.users.ByUsername(email)
		if err != nil || rec.CredentialVersion != version || rec.Username != email {
			return Session{}, false
		}
		return Session{Username: rec.Username, MustChangePassword: rec.MustChangePassword}, true
	default:
		return Session{}, false
	}
}

func (a *Auth) isAuthorized(email string) bool {
	_, ok := a.authorizedEmails[email]
	return ok
}

func hashEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}

// Sign builds a password-fallback session token that expires at the given time.
// version must match the account's credential version or Session rejects the token.
func (a *Auth) Sign(username string, version int, expires time.Time) string {
	return a.sign(passwordAuthMethod, username, version, expires)
}

func (a *Auth) sign(method, principal string, version int, expires time.Time) string {
	payload := strings.Join([]string{"v1", method, principal, strconv.FormatInt(expires.Unix(), 10), strconv.Itoa(version)}, "|")
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Read validates a signed session token and returns its principal and credential version.
func (a *Auth) Read(token string, now time.Time) (string, int, bool) {
	principal, version, _, ok := a.read(token, now)
	return principal, version, ok
}

func (a *Auth) read(token string, now time.Time) (string, int, string, bool) {
	if a == nil || len(token) > 4096 {
		return "", 0, "", false
	}
	payload, sig, ok := strings.Cut(token, ".")
	if !ok {
		return "", 0, "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", 0, "", false
	}
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return "", 0, "", false
	}
	mac := hmac.New(sha256.New, a.secret)
	mac.Write(raw)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return "", 0, "", false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 5 || parts[0] != "v1" || parts[2] == "" {
		return "", 0, "", false
	}
	exp, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil || !now.Before(time.Unix(exp, 0)) {
		return "", 0, "", false
	}
	version, err := strconv.Atoi(parts[4])
	if err != nil || version < 0 {
		return "", 0, "", false
	}
	switch parts[1] {
	case googleAuthMethod:
		if version != 0 {
			return "", 0, "", false
		}
	case passwordAuthMethod:
		if version < 1 {
			return "", 0, "", false
		}
	default:
		return "", 0, "", false
	}
	return parts[2], version, parts[1], true
}
