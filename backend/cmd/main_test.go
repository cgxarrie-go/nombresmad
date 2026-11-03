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
			if _, err := readAuthConfig(); err == nil {
				t.Fatal("incomplete authentication configuration was accepted")
			}
		})
	}
}

func TestReadAuthConfigAllowsDisabledFallbackWithoutAllowlist(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "web-client-id")
	t.Setenv("ADMIN_SECRET", "stable-signing-secret")
	t.Setenv("ADMIN_PASSWORD", "")
	config, err := readAuthConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.password != "" {
		t.Fatalf("fallback password should be disabled: %+v", config)
	}
}
