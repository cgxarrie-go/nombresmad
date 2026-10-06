package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookie = "nombresmad_session"
	sessionTTL    = 12 * time.Hour
)

var (
	adminUser     string
	adminPass     string
	sessionSecret []byte
)

func initAuth() {
	adminUser = os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass = os.Getenv("ADMIN_PASSWORD")
	if secret := os.Getenv("ADMIN_SECRET"); secret != "" {
		sessionSecret = []byte(secret)
	} else {
		sessionSecret = make([]byte, 32)
		if _, err := rand.Read(sessionSecret); err != nil {
			log.Fatal(err)
		}
	}
	if adminPass == "" {
		log.Println("ADMIN_PASSWORD is empty; admin login is disabled")
	}
}

func credentialsMatch(username, password string) bool {
	if adminPass == "" {
		return false
	}
	userOK := hashEqual(username, adminUser)
	passOK := hashEqual(password, adminPass)
	return userOK && passOK
}

func hashEqual(a, b string) bool {
	da := sha256.Sum256([]byte(a))
	db := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(da[:], db[:]) == 1
}

func signSession(username string, expires time.Time) string {
	payload := username + "|" + strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, sessionSecret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func readSession(token string, now time.Time) (string, bool) {
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
	mac := hmac.New(sha256.New, sessionSecret)
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

func setSessionCookie(w http.ResponseWriter, username string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    signSession(username, time.Now().Add(sessionTTL)),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func sessionUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", false
	}
	return readSession(cookie.Value, time.Now())
}

func requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if _, ok := sessionUser(r); ok {
		return true
	}
	http.Error(w, "no autorizado", http.StatusUnauthorized)
	return false
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !credentialsMatch(strings.TrimSpace(body.Username), body.Password) {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
		return
	}
	setSessionCookie(w, adminUser)
	writeJSON(w, map[string]string{"username": adminUser})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clearSessionCookie(w)
	writeJSON(w, map[string]string{"status": "logged out"})
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := sessionUser(r)
	if !ok {
		http.Error(w, "no autorizado", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{"username": user})
}
