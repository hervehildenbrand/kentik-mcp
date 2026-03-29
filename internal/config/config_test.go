package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingCredentials(t *testing.T) {
	os.Unsetenv("KENTIK_EMAIL")
	os.Unsetenv("KENTIK_API_TOKEN")
	os.Unsetenv("KENTIK_PROFILE")

	_, err := Load()
	if err == nil {
		t.Error("expected error when credentials missing")
	}
}

func TestLoad_DefaultRegionEU(t *testing.T) {
	t.Setenv("KENTIK_EMAIL", "test@example.com")
	t.Setenv("KENTIK_API_TOKEN", "token123")
	os.Unsetenv("KENTIK_REGION")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Region != "EU" {
		t.Errorf("expected region EU, got %s", cfg.Region)
	}
	if cfg.V5Base != "https://api.kentik.eu/api/v5" {
		t.Errorf("unexpected V5Base: %s", cfg.V5Base)
	}
	if cfg.V6Base != "https://grpc.api.kentik.eu" {
		t.Errorf("unexpected V6Base: %s", cfg.V6Base)
	}
}

func TestLoad_USRegion(t *testing.T) {
	t.Setenv("KENTIK_EMAIL", "test@example.com")
	t.Setenv("KENTIK_API_TOKEN", "token123")
	t.Setenv("KENTIK_REGION", "US")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Region != "US" {
		t.Errorf("expected region US, got %s", cfg.Region)
	}
	if cfg.V5Base != "https://api.kentik.com/api/v5" {
		t.Errorf("unexpected V5Base: %s", cfg.V5Base)
	}
	if cfg.V6Base != "https://grpc.api.kentik.com" {
		t.Errorf("unexpected V6Base: %s", cfg.V6Base)
	}
}

func TestLoad_ProfileFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".kentik-mcp.json")
	configData := `{
		"profiles": {
			"default": {
				"email": "default@company.com",
				"api_token": "default-token",
				"region": "EU"
			},
			"us-prod": {
				"email": "prod@company.com",
				"api_token": "prod-token",
				"region": "US"
			}
		}
	}`
	os.WriteFile(configPath, []byte(configData), 0600)

	// Override home dir for test
	t.Setenv("HOME", tmpDir)
	os.Unsetenv("KENTIK_EMAIL")
	os.Unsetenv("KENTIK_API_TOKEN")
	os.Unsetenv("KENTIK_REGION")
	os.Unsetenv("KENTIK_PROFILE")

	// Test default profile
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Email != "default@company.com" {
		t.Errorf("expected default email, got %s", cfg.Email)
	}
	if cfg.Region != "EU" {
		t.Errorf("expected EU, got %s", cfg.Region)
	}
}

func TestLoad_NamedProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".kentik-mcp.json")
	configData := `{
		"profiles": {
			"default": {
				"email": "default@company.com",
				"api_token": "default-token",
				"region": "EU"
			},
			"us-prod": {
				"email": "prod@company.com",
				"api_token": "prod-token",
				"region": "US"
			}
		}
	}`
	os.WriteFile(configPath, []byte(configData), 0600)

	t.Setenv("HOME", tmpDir)
	t.Setenv("KENTIK_PROFILE", "us-prod")
	os.Unsetenv("KENTIK_EMAIL")
	os.Unsetenv("KENTIK_API_TOKEN")
	os.Unsetenv("KENTIK_REGION")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Email != "prod@company.com" {
		t.Errorf("expected prod email, got %s", cfg.Email)
	}
	if cfg.Region != "US" {
		t.Errorf("expected US, got %s", cfg.Region)
	}
	if cfg.V5Base != "https://api.kentik.com/api/v5" {
		t.Errorf("expected US V5 base, got %s", cfg.V5Base)
	}
}

func TestLoad_EnvOverridesProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".kentik-mcp.json")
	configData := `{
		"profiles": {
			"default": {
				"email": "profile@company.com",
				"api_token": "profile-token",
				"region": "EU"
			}
		}
	}`
	os.WriteFile(configPath, []byte(configData), 0600)

	t.Setenv("HOME", tmpDir)
	t.Setenv("KENTIK_EMAIL", "env@override.com")
	t.Setenv("KENTIK_API_TOKEN", "env-token")
	t.Setenv("KENTIK_REGION", "US")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Email != "env@override.com" {
		t.Errorf("env should override profile, got %s", cfg.Email)
	}
	if cfg.APIToken != "env-token" {
		t.Errorf("env should override profile, got %s", cfg.APIToken)
	}
}

func TestLoad_CustomURLOverride(t *testing.T) {
	t.Setenv("KENTIK_EMAIL", "test@example.com")
	t.Setenv("KENTIK_API_TOKEN", "token")
	t.Setenv("KENTIK_V5_URL", "https://custom.api.local/api/v5")
	t.Setenv("KENTIK_V6_URL", "https://custom.grpc.local")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.V5Base != "https://custom.api.local/api/v5" {
		t.Errorf("expected custom V5 URL, got %s", cfg.V5Base)
	}
	if cfg.V6Base != "https://custom.grpc.local" {
		t.Errorf("expected custom V6 URL, got %s", cfg.V6Base)
	}
}
