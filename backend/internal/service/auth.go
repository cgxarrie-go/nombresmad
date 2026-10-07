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

// Auth checks the admin password and signs session tokens.
type Auth struct {
	user     string
	password string
	secret   []byte
}

// NewAuth uses credentials and a session secret created by the caller.
func NewAuth(user, password string, secret []byte) *Auth {
	return &Auth{user: user, password: password, secret: secret}
}

// Username is the configured admin account.
func (a *Auth) Username() string { return a.user }

// Check reports whether the supplied credentials match.
func (a *Auth) Check(username, password string) bool {
	if a.password == "" {
		return false
	}
	userOK := hashEqual(username, a.user)
	passOK := hashEqual(password, a.password)
	return userOK && passOK
}

func hashEqual(a, b string) bool {
	left := sha256.Sum256([]byte(a))
	right := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}

// Sign builds a session token that expires at the given time.
func (a *Auth) Sign(username string, expires time.Time) string {
	payload := username + "|" + strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Read validates a session token at the given time.
func (a *Auth) Read(token string, now time.Time) (string, bool) {
	payload, sig, ok := strings.Cut(token, ".")
	if !ok {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", false
	}
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, a.secret)
	mac.Write(raw)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return "", false
	}
	user, expText, ok := strings.Cut(string(raw), "|")
	if !ok || user == "" {
		return "", false
	}
	exp, err := strconv.ParseInt(expText, 10, 64)
	if err != nil || !now.Before(time.Unix(exp, 0)) {
		return "", false
	}
	return user, true
}
