package account

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	GAllAccount    *Config
	CurrentAccount *Account
)

type Account struct {
	Name    string
	Account string
	// TODO: remove it to replace it with the new API listZones.
	SosEndpoint string
	// Endpoint is optional.
	Endpoint string
	// Environment will be deprecated and removed,
	// once everything is egoscale v3 migrated.
	Environment         string
	Key                 string
	Secret              string
	SecretCommand       []string
	DefaultZone         string
	DefaultSSHKey       string
	DefaultTemplate     string
	DefaultOutputFormat string
	ClientTimeout       int
	CustomHeaders       map[string]string
}

func (a Account) APISecret() string {
	if len(a.SecretCommand) != 0 {
		cmd := exec.Command(a.SecretCommand[0], a.SecretCommand[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		return strings.TrimRight(string(out), "\n")
	}

	return a.Secret
}

type Config struct {
	DefaultAccount      string
	DefaultOutputFormat string
	Accounts            []Account
}

func (a Account) IsDefault() bool {
	return a.Name == GAllAccount.DefaultAccount
}
