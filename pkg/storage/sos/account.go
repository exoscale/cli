package sos

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/exoscale/egoscale"
)

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

func (a Account) AccountName() string {
	if a.Name == "" {
		resp, err := cs.GetWithContext(gContext, egoscale.Account{})
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
	Accounts            []AccountConfig
}

var GAllAccount *AccountConfig

func (a Account) IsDefault() bool {
	return a.Name == GAllAccount.DefaultAccount
}
