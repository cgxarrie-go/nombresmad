package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"nombresmad/backend/internal/model"
)

const (
	minChosenPasswordLength = 8
	maxChosenPasswordLength = 128
)

func randomToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "dummy-password"
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

// EnsureDefaultAdmin inserts the configured email administrator when none exists.
// The password is ADMIN_PASSWORD and the account must change it at first login.
// Later startups leave an existing email account unchanged.
func (a *Auth) EnsureDefaultAdmin(username string) error {
	if a == nil || a.users == nil {
		return errors.New("admin store is not configured")
	}
	rows, err := a.users.List()
	if err != nil {
		return err
	}
	for _, row := range rows {
		if _, ok := NormalizeEmail(row.Username); ok {
			return nil
		}
	}
	email, ok := NormalizeEmail(username)
	if !ok {
		if a.defaultPassword == "" {
			log.Println("ADMIN_PASSWORD is empty; admin login is disabled")
			return nil
		}
		return errors.New("ADMIN_USER must be an email address")
	}
	if a.defaultPassword == "" {
		log.Println("ADMIN_PASSWORD is empty; admin login is disabled")
		return nil
	}
	if err := a.insertAdmin(email, a.defaultPassword, true); err != nil {
		if errors.Is(err, model.ErrExists) {
			return nil
		}
		return err
	}
	log.Printf("created default admin %q; password change is required at first login", email)
	return nil
}

// Create adds an administrator whose password is the configured default.
// The new account must change it at first login.
func (a *Auth) Create(username string) (model.AdminView, error) {
	email, ok := NormalizeEmail(username)
	if !ok {
		return model.AdminView{}, model.Invalid("El usuario debe ser un correo electrónico")
	}
	username = email
	if a.defaultPassword == "" {
		return model.AdminView{}, model.Invalid("La contraseña por defecto no está configurada")
	}
	if err := a.insertAdmin(username, a.defaultPassword, true); err != nil {
		return model.AdminView{}, err
	}
	return model.AdminView{Username: username, MustChangePassword: true}, nil
}

// ResetPassword sets another administrator's password back to the default
// and requires a change at the next login. Existing sessions for that account stop working.
func (a *Auth) ResetPassword(actor, target string) error {
	actorEmail, actorOK := NormalizeEmail(actor)
	targetEmail, targetOK := NormalizeEmail(target)
	if !actorOK || !targetOK {
		return model.Invalid("El usuario debe ser un correo electrónico")
	}
	actor, target = actorEmail, targetEmail
	if actor == target {
		return model.Invalid("No puedes restablecer tu propia contraseña")
	}
	if a.defaultPassword == "" {
		return model.Invalid("La contraseña por defecto no está configurada")
	}
	if _, err := a.users.ByUsername(target); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.Invalid("Ese administrador no existe")
		}
		return err
	}
	hash, err := HashPassword(a.defaultPassword)
	if err != nil {
		return err
	}
	_, err = a.users.UpdatePassword(target, hash, true)
	return err
}

// ChangePassword replaces the caller's password and returns a new session token.
func (a *Auth) ChangePassword(username, current, next string, expires time.Time) (string, error) {
	user, _, _, ok := a.authenticate(username, current)
	if !ok {
		return "", model.Invalid("La contraseña actual no es correcta")
	}
	if err := validateChosenPassword(next, current, a.defaultPassword); err != nil {
		return "", err
	}
	hash, err := HashPassword(next)
	if err != nil {
		return "", err
	}
	version, err := a.users.UpdatePassword(user, hash, false)
	if err != nil {
		return "", err
	}
	return a.Sign(user, version, expires), nil
}

// List returns administrators without password hashes.
func (a *Auth) List() ([]model.AdminView, error) {
	rows, err := a.users.List()
	if err != nil {
		return nil, err
	}
	out := make([]model.AdminView, 0, len(rows))
	for _, row := range rows {
		email, ok := NormalizeEmail(row.Username)
		if !ok {
			continue
		}
		out = append(out, model.AdminView{Username: email, MustChangePassword: row.MustChangePassword})
	}
	return out, nil
}

func (a *Auth) insertAdmin(username, password string, mustChange bool) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return a.users.Insert(username, hash, mustChange)
}

func validateChosenPassword(next, current, fallback string) error {
	if len(next) < minChosenPasswordLength {
		return model.Invalid("La nueva contraseña debe tener al menos 8 caracteres")
	}
	if len(next) > maxChosenPasswordLength {
		return model.Invalid("La nueva contraseña es demasiado larga")
	}
	if hashEqual(next, current) {
		return model.Invalid("La nueva contraseña debe ser distinta de la actual")
	}
	if fallback != "" && hashEqual(next, fallback) {
		return model.Invalid("Elige una contraseña distinta de la contraseña por defecto")
	}
	return nil
}

// NormalizeEmail returns the lowercase address when it is a usable email.
func NormalizeEmail(address string) (string, bool) {
	address = strings.ToLower(strings.TrimSpace(address))
	if address == "" || len(address) > 254 || strings.ContainsAny(address, " \t|") {
		return "", false
	}
	local, domain, ok := strings.Cut(address, "@")
	if !ok || strings.Contains(domain, "@") || !validEmailLocal(local) || !validEmailDomain(domain) {
		return "", false
	}
	return address, true
}

func validEmailLocal(local string) bool {
	if local == "" || len(local) > 64 || strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
		return false
	}
	for _, r := range local {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || strings.ContainsRune("._%+-", r) {
			continue
		}
		return false
	}
	return true
}

func validEmailDomain(domain string) bool {
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") || strings.Contains(domain, "..") {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels[len(labels)-1]) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
	}
	return true
}
