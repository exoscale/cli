package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksRotateCCMCredentialsCmd = &cobra.Command{
	Use:   "rotate-ccm-credentials <cluster name | ID>",
	Short: "Rotate the Exoscale Cloud Controller IAM credentials for a SKS cluster",
	Long: `This command rotates the Exoscale IAM credentials managed by the SKS control
plane for the Kubernetes Exoscale Cloud Controller Manager.
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		if err := cluster.RotateCCMCredentials(ctx); err != nil {
			return fmt.Errorf("error rotating credentials: %s", err)
		}

		fmt.Println("CCM credentials rotated successfully")

		return nil
	},
}

func init() {
	sksRotateCCMCredentialsCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksRotateCCMCredentialsCmd)
}
