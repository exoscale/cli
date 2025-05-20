package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

var sksAuthorityCertAuthorities = []v3.GetSKSClusterAuthorityCertAuthority{
	"aggregation",
	"control-plane",
	"kubelet",
}

type sksAuthorityCertCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"authority-cert"`

	Cluster   string                                 `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Authority v3.GetSKSClusterAuthorityCertAuthority `cli-arg:"#"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksAuthorityCertCmd) CmdAliases() []string { return nil }

func (c *sksAuthorityCertCmd) CmdShort() string {
	return "Retrieve an authority certificate for an SKS cluster"
}

func (c *sksAuthorityCertCmd) CmdLong() string {
	stringAuthorities := make([]string, len(sksAuthorityCertAuthorities))
	for i, v := range sksAuthorityCertAuthorities {
		stringAuthorities[i] = string(v)
	}

	return fmt.Sprintf(`This command retrieves the certificate content for the specified Kubernetes
cluster authority. Supported authorities:

Supported authorities: %s`,
		strings.Join(stringAuthorities, ", "))
}

func (c *sksAuthorityCertCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksAuthorityCertCmd) CmdRun(cmd *cobra.Command, _ []string) error {
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

	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	getCertResponse, err := client.GetSKSClusterAuthorityCert(ctx, cluster.ID, c.Authority)
	if err != nil {
		return fmt.Errorf("error retrieving certificate: %w", err)
	}

	cert, err := base64.StdEncoding.DecodeString(getCertResponse.Cacert)
	if err != nil {
		return fmt.Errorf("error decoding certificate content: %w", err)
	}

	fmt.Print(string(cert))

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(sksCmd, &sksAuthorityCertCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
