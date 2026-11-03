package main

import "testing"

func TestReadAuthConfigRequiresGoogleConfigurationAndSigningSecret(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		secret   string
	}{
		{name: "missing client id", secret: "stable-secret"},
		{name: "missing signing secret", clientID: "web-client-id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOOGLE_CLIENT_ID", tt.clientID)
			t.Setenv("ADMIN_SECRET", tt.secret)
			t.Setenv("AUTHORIZED_EMAILS", "admin@example.com")
			if _, err := readAuthConfig(); err == nil {
				t.Fatal("incomplete authentication configuration was accepted")
			}
		})
	}
}

func TestReadAuthConfigAllowsDisabledFallbackWithAllowlist(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "web-client-id")
	t.Setenv("ADMIN_SECRET", "stable-signing-secret")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("AUTHORIZED_EMAILS", " Admin@Example.com,other@example.com ")
	config, err := readAuthConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.password != "" {
		t.Fatalf("fallback password should be disabled: %+v", config)
	}
	if len(config.authorizedEmails) != 2 || config.authorizedEmails[0] != "admin@example.com" || config.authorizedEmails[1] != "other@example.com" {
		t.Fatalf("authorized emails=%v", config.authorizedEmails)
	}
}

func TestReadAuthConfigRequiresValidAuthorizedEmails(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "web-client-id")
	t.Setenv("ADMIN_SECRET", "stable-signing-secret")
	t.Setenv("ADMIN_PASSWORD", "")
	for _, value := range []string{"", " ", "admin@example.com,", "invalid"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("AUTHORIZED_EMAILS", value)
			if _, err := readAuthConfig(); err == nil {
				t.Fatal("invalid authorized email configuration was accepted")
			}
		})
	}
}
