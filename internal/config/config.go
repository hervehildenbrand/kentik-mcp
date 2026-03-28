package config

import (
	"errors"
	"os"
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

func Load() (*Config, error) {
	email := os.Getenv("KENTIK_EMAIL")
	apiToken := os.Getenv("KENTIK_API_TOKEN")
	region := strings.ToUpper(os.Getenv("KENTIK_REGION"))

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

	if region == "EU" {
		cfg.V5Base = "https://api.kentik.eu/api/v5"
		cfg.V6Base = "https://grpc.api.kentik.eu"
	} else {
		cfg.V5Base = "https://api.kentik.com/api/v5"
		cfg.V6Base = "https://grpc.api.kentik.com"
	}

	return cfg, nil
}
