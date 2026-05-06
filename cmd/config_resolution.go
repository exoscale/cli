package cmd

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/exoscale/cli/pkg/account"
)

// envSources holds account-level values read from environment variables.
type envSources struct {
	apiKey         string
	apiSecret      string
	apiEndpoint    string
	apiEnvironment string
	sosEndpoint    string
	zone           string
	clientTimeout  *int // nil when EXOSCALE_API_TIMEOUT is not set
}

// fileSources holds the raw loaded config and the selected account profile.
// profile is nil when the config file is missing or the account was not found.
type fileSources struct {
	config  *account.Config
	profile *account.Account
}

// readEnvSources populates an envSources from the current environment.
func readEnvSources() envSources {
	s := envSources{
		apiEndpoint:    os.Getenv("EXOSCALE_API_ENDPOINT"),
		apiEnvironment: readFromEnv("EXOSCALE_API_ENVIRONMENT"),
		apiKey: readFromEnv(
			"EXOSCALE_API_KEY",
			"EXOSCALE_KEY",
			"CLOUDSTACK_KEY",
			"CLOUDSTACK_API_KEY",
		),
		apiSecret: readFromEnv(
			"EXOSCALE_API_SECRET",
			"EXOSCALE_SECRET",
			"EXOSCALE_SECRET_KEY",
			"CLOUDSTACK_SECRET",
			"CLOUDSTACK_SECRET_KEY",
		),
		sosEndpoint: readFromEnv(
			"EXOSCALE_STORAGE_API_ENDPOINT",
			"EXOSCALE_SOS_ENDPOINT",
		),
		zone: readFromEnv("EXOSCALE_ZONE"),
	}

	if raw := readFromEnv("EXOSCALE_API_TIMEOUT"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			s.clientTimeout = &n
		}
	}

	return s
}

// hasCredentials reports whether both API key and secret are set.
func (s envSources) hasCredentials() bool {
	return s.apiKey != "" && s.apiSecret != ""
}

// loadFileSources reads the config file from v and selects the named account.
// Returns a zero fileSources when the file is not found.
func loadFileSources(v *viper.Viper, accountName string) (fileSources, error) {
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fileSources{}, nil
		}
		return fileSources{}, err
	}

	cfg := &account.Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return fileSources{}, fmt.Errorf("couldn't read config: %w", err)
	}

	if accountName == "" {
		accountName = cfg.DefaultAccount
	}

	for i := range cfg.Accounts {
		if cfg.Accounts[i].Name == accountName {
			return fileSources{config: cfg, profile: &cfg.Accounts[i]}, nil
		}
	}

	// Keep config even without a matching profile so callers can list accounts.
	return fileSources{config: cfg, profile: nil}, nil
}

// resolve merges env, file profile, and built-in defaults in that order of precedence.
func resolve(env envSources, file fileSources) account.Account {
	var acc account.Account
	if file.profile != nil {
		acc = *file.profile
	}

	acc.Environment = cmp.Or(env.apiEnvironment, acc.Environment, DefaultEnvironment)
	acc.DefaultZone = cmp.Or(env.zone, acc.DefaultZone, DefaultZone)
	acc.SosEndpoint = strings.TrimRight(cmp.Or(env.sosEndpoint, acc.SosEndpoint, DefaultSosEndpoint), "/")
	acc.Endpoint = cmp.Or(env.apiEndpoint, acc.Endpoint)
	if env.clientTimeout != nil {
		acc.ClientTimeout = *env.clientTimeout
	}
	if env.hasCredentials() {
		acc.Key, acc.Secret, acc.SecretCommand = env.apiKey, env.apiSecret, nil
	}

	return acc
}
