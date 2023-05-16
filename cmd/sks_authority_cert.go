package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

var sksAuthorityCertAuthorities = []string{
	"aggregation",
	"control-plane",
	"kubelet",
}

type sksAuthorityCertCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"authority-cert"`

	Cluster   string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Authority string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksAuthorityCertCmd) cmdAliases() []string { return nil }

func (c *sksAuthorityCertCmd) cmdShort() string {
	return "Retrieve an authority certificate for an SKS cluster"
}

func (c *sksAuthorityCertCmd) cmdLong() string {
	return fmt.Sprintf(`This command retrieves the certificate content for the specified Kubernetes
cluster authority. Supported authorities:

Supported authorities: %s`,
		strings.Join(sksAuthorityCertAuthorities, ", "))
}

func (c *sksAuthorityCertCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksAuthorityCertCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var authOK bool
	for _, v := range sksAuthorityCertAuthorities {
		if c.Authority == v {
			authOK = true
			break
		}
	}
	if !authOK {
		cmdExitOnUsageError(cmd, fmt.Sprintf("unsupported authority value %q", c.Authority))
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	b64Cert, err := globalstate.EgoscaleClient.GetSKSClusterAuthorityCert(ctx, c.Zone, cluster, c.Authority)
	if err != nil {
		return fmt.Errorf("error retrieving certificate: %w", err)
	}

	cert, err := base64.StdEncoding.DecodeString(b64Cert)
	if err != nil {
		return fmt.Errorf("error decoding certificate content: %w", err)
	}

	fmt.Print(string(cert))

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksAuthorityCertCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedSKSCmd, &sksAuthorityCertCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
