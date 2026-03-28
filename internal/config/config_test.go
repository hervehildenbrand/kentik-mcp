package config

import (
	"os"
	"testing"
)

func TestLoad_MissingCredentials(t *testing.T) {
	os.Unsetenv("KENTIK_EMAIL")
	os.Unsetenv("KENTIK_API_TOKEN")

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
