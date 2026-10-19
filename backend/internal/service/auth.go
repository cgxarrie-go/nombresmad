package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Session is an authenticated administrator.
type Session struct {
	Username           string
	MustChangePassword bool
}

// Auth checks administrator passwords and signs session tokens.
type Auth struct {
	users           AdminStore
	defaultPassword string
	secret          []byte
	dummyHash       string
}

// NewAuth uses the account store, the default password from the environment, and a session secret.
func NewAuth(users AdminStore, defaultPassword string, secret []byte) *Auth {
	dummy, err := HashPassword(randomToken())
	if err != nil {
		dummy = ""
	}
	return &Auth{users: users, defaultPassword: defaultPassword, secret: secret, dummyHash: dummy}
}

// Check reports whether the supplied credentials match a stored account.
func (a *Auth) Check(username, password string) bool {
	_, _, _, ok := a.authenticate(username, password)
	return ok
}

func (a *Auth) authenticate(username, password string) (storedUsername string, version int, mustChange bool, ok bool) {
	email, emailOK := NormalizeEmail(username)
	if a == nil || !emailOK || password == "" || len(password) > 1024 {
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

// Login checks the credentials and returns a session token.
func (a *Auth) Login(username, password string, expires time.Time) (string, Session, bool) {
	user, version, mustChange, ok := a.authenticate(username, password)
	if !ok {
		return "", Session{}, false
	}
	return a.Sign(user, version, expires), Session{Username: user, MustChangePassword: mustChange}, true
}

// Session validates a bearer token against the current credential version.
func (a *Auth) Session(token string, now time.Time) (Session, bool) {
	username, version, ok := a.Read(token, now)
	if !ok || a == nil || a.users == nil {
		return Session{}, false
	}
	email, emailOK := NormalizeEmail(username)
	if !emailOK {
		return Session{}, false
	}
	rec, err := a.users.ByUsername(email)
	if err != nil || rec.CredentialVersion != version || rec.Username != email {
		return Session{}, false
	}
	return Session{Username: rec.Username, MustChangePassword: rec.MustChangePassword}, true
}

func hashEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}

// Sign builds a session token that expires at the given time.
// version must match the account's credential version or Session rejects the token.
func (a *Auth) Sign(username string, version int, expires time.Time) string {
	payload := username + "|" + strconv.FormatInt(expires.Unix(), 10) + "|" + strconv.Itoa(version)
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Read validates a session token at the given time.
func (a *Auth) Read(token string, now time.Time) (string, int, bool) {
	payload, sig, ok := strings.Cut(token, ".")
	if !ok {
		return "", 0, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", 0, false
	}
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return "", 0, false
	}
	mac := hmac.New(sha256.New, a.secret)
	mac.Write(raw)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return "", 0, false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 || parts[0] == "" {
		return "", 0, false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || !now.Before(time.Unix(exp, 0)) {
		return "", 0, false
	}
	version, err := strconv.Atoi(parts[2])
	if err != nil || version < 1 {
		return "", 0, false
	}
	return parts[0], version, true
}
