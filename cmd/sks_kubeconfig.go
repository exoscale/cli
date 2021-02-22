package cmd

import (
	"encoding/base64"
	"fmt"
	"time"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksKubeconfigCmd = &cobra.Command{
	Use:     "kubeconfig <cluster name | ID> <user>",
	Aliases: []string{"kc"},
	Short:   "Generate a Kubernetes kubeconfig file for a SKS cluster",
	Long: `This command generates a kubeconfig file to be used for authenticating to a SKS
cluster API.

The "user" command argument corresponds to the CN field of the generated X.509
client certificate. Optionally, you can specify client certificate groups
using the "-g|--group" option: those groups will be set in the "O" field of
the certificate. See [1] for more information about Kubernetes authentication
certificates.

Example usage:

    # Obtain "cluster-admin" credentials
    $ exo sks kubeconfig my-cluster admin \
        -g system:masters \
        -t $((86400 * 7)) > $HOME/.kube/my-cluster.config
    $ kubeconfig --kubeconfig=$HOME/.kube/my-cluster.config get pods

Note: if no TTL value is specified, the API applies a default value as a
safety measure. Please look up the API documentation for more information.

[1]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{
			"user",
			"zone",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c    = args[0]
			user = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		groups, err := cmd.Flags().GetStringSlice("group")
		if err != nil {
			return err
		}

		ttl, err := cmd.Flags().GetInt64("ttl")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		b64Kubeconfig, err := cluster.RequestKubeconfig(ctx, user, groups, time.Duration(ttl)*time.Second)
		if err != nil {
			return fmt.Errorf("error retrieving kubeconfig: %s", err)
		}

		kubeconfig, err := base64.StdEncoding.DecodeString(b64Kubeconfig)
		if err != nil {
			return fmt.Errorf("error decoding kubeconfig content: %s", err)
		}

		fmt.Println(string(kubeconfig))

		return nil
	},
}

func init() {
	sksKubeconfigCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksKubeconfigCmd.Flags().StringSliceP("group", "g", nil,
		"client certificate group. Can be specified multiple times.")
	sksKubeconfigCmd.Flags().Int64P("ttl", "t", 0,
		"client certificate validity duration in seconds")
	sksCmd.AddCommand(sksKubeconfigCmd)
}
