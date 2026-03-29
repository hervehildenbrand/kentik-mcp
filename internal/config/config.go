package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrMissingCredentials = errors.New("KENTIK_EMAIL and KENTIK_API_TOKEN are required")

type Config struct {
	Email    string
	APIToken string
	Region   string
	V5Base   string
	V6Base   string
}

// profileFile represents the ~/.kentik-mcp.json structure.
type profileFile struct {
	Profiles map[string]profileEntry `json:"profiles"`
}

type profileEntry struct {
	Email    string `json:"email"`
	APIToken string `json:"api_token"`
	Region   string `json:"region"`
}

// Load creates a Config with this precedence:
//  1. Environment variables (KENTIK_EMAIL, KENTIK_API_TOKEN, KENTIK_REGION)
//  2. Named profile from ~/.kentik-mcp.json (selected by KENTIK_PROFILE env var)
//  3. "default" profile from ~/.kentik-mcp.json
func Load() (*Config, error) {
	email := os.Getenv("KENTIK_EMAIL")
	apiToken := os.Getenv("KENTIK_API_TOKEN")
	region := strings.ToUpper(os.Getenv("KENTIK_REGION"))

	// If env vars are incomplete, try loading from profile file
	if email == "" || apiToken == "" {
		profile := loadProfile()
		if profile != nil {
			if email == "" {
				email = profile.Email
			}
			if apiToken == "" {
				apiToken = profile.APIToken
			}
			if region == "" {
				region = strings.ToUpper(profile.Region)
			}
		}
	}

	if email == "" || apiToken == "" {
		return nil, ErrMissingCredentials
	}

	if region == "" {
		region = "EU"
	}

	cfg := &Config{
		Email:    email,
		APIToken: apiToken,
		Region:   region,
	}

	// Allow full URL override for custom/on-prem deployments
	if v5url := os.Getenv("KENTIK_V5_URL"); v5url != "" {
		cfg.V5Base = v5url
	} else if region == "EU" {
		cfg.V5Base = "https://api.kentik.eu/api/v5"
	} else {
		cfg.V5Base = "https://api.kentik.com/api/v5"
	}

	if v6url := os.Getenv("KENTIK_V6_URL"); v6url != "" {
		cfg.V6Base = v6url
	} else if region == "EU" {
		cfg.V6Base = "https://grpc.api.kentik.eu"
	} else {
		cfg.V6Base = "https://grpc.api.kentik.com"
	}

	return cfg, nil
}

func loadProfile() *profileEntry {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(home, ".kentik-mcp.json"))
	if err != nil {
		return nil
	}

	var pf profileFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil
	}

	profileName := os.Getenv("KENTIK_PROFILE")
	if profileName == "" {
		profileName = "default"
	}

	if p, ok := pf.Profiles[profileName]; ok {
		return &p
	}

	return nil
}
