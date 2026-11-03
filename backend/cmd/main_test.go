package main

import "testing"

func TestReadAuthConfigRequiresGoogleConfigurationAndSigningSecret(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		emails   string
		secret   string
	}{
		{name: "missing client id", emails: "admin@example.com", secret: "stable-secret"},
		{name: "missing allowlist", clientID: "web-client-id", secret: "stable-secret"},
		{name: "missing signing secret", clientID: "web-client-id", emails: "admin@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOOGLE_CLIENT_ID", tt.clientID)
			t.Setenv("AUTHORIZED_EMAILS", tt.emails)
			t.Setenv("ADMIN_SECRET", tt.secret)
			if _, err := readAuthConfig(); err == nil {
				t.Fatal("incomplete authentication configuration was accepted")
			}
		})
	}
}

func TestReadAuthConfigAllowsDisabledFallback(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "web-client-id")
	t.Setenv("AUTHORIZED_EMAILS", " Admin@Example.com,admin@example.com ")
	t.Setenv("ADMIN_SECRET", "stable-signing-secret")
	t.Setenv("ADMIN_PASSWORD", "")
	config, err := readAuthConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.password != "" || len(config.authorizedEmails) != 1 || config.authorizedEmails[0] != "admin@example.com" {
		t.Fatalf("config=%+v", config)
	}
}
