package cmd

import (
	"encoding/base64"
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksAuthorityCertAuthorities = []string{
	"aggregation",
	"kubelet",
}

var sksAuthorityCertCmd = &cobra.Command{
	Use:   "authority-cert CLUSTER-NAME|ID AUTHORITY",
	Short: "Retrieve an authority certificate for a SKS cluster",
	Long: fmt.Sprintf(`This command retrieves the certificate content for the specified Kubernetes
cluster authority. Supported authorities:

Supported authorities: %s`,
		strings.Join(sksAuthorityCertAuthorities, ", ")),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		var authOK bool
		for _, v := range sksAuthorityCertAuthorities {
			if args[1] == v {
				authOK = true
				break
			}
		}
		if !authOK {
			cmdExitOnUsageError(cmd, fmt.Sprintf("unsupported authority value %q", args[1]))
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c         = args[0]
			authority = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		b64Cert, err := cluster.AuthorityCert(ctx, authority)
		if err != nil {
			return fmt.Errorf("error retrieving certificate: %s", err)
		}

		cert, err := base64.StdEncoding.DecodeString(b64Cert)
		if err != nil {
			return fmt.Errorf("error decoding certificate content: %s", err)
		}

		fmt.Print(string(cert))

		return nil
	},
}

func init() {
	sksAuthorityCertCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksAuthorityCertCmd)
}
