package client

import (
	"github.com/exoscale/cli/pkg/account"
	egov3 "github.com/exoscale/egoscale/v3"
)

func Get() (*egov3.ZonedClient, error) {
	return egov3.DefaultClient(egov3.ClientOptWithCredentials(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
	))
}
