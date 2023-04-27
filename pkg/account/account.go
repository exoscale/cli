package account

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
)

var (
	GAllAccount    *AccountConfig
	CurrentAccount *Account
)

type Account struct {
	Name                 string
	Account              string
	Endpoint             string
	ComputeEndpoint      string // legacy config.
	DNSEndpoint          string
	SosEndpoint          string
	RunstatusEndpoint    string
	Environment          string
	Key                  string
	Secret               string
	SecretCommand        []string
	DefaultZone          string
	DefaultSSHKey        string
	DefaultTemplate      string
	DefaultRunstatusPage string
	DefaultOutputFormat  string
	ClientTimeout        int
	CustomHeaders        map[string]string
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

func (a Account) AccountName(ctx context.Context) string {
	if a.Name == "" {
		resp, err := globalstate.EgoscaleClient.GetWithContext(ctx, egoscale.Account{})
		if err != nil {
			log.Fatal(err)
		}
		acc := resp.(*egoscale.Account)
		return acc.Name
	}

	return a.Name
}

type AccountConfig struct {
	DefaultAccount      string
	DefaultOutputFormat string
	Accounts            []Account
}

func (a Account) IsDefault() bool {
	return a.Name == GAllAccount.DefaultAccount
}
