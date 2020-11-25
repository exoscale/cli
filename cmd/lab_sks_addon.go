package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sksLabAddonCmd = &cobra.Command{
	Use:   "addon <cluster name | ID> <add-on name>",
	Short: "Deploy add-ons to a SKS cluster",
	Long: `This commands deploys well-known Kubernetes manifests to a SKS cluster.

To deploy the selected add-ons to the specified SKS cluster, valid Kubernetes
cluster API are requested for authentication. To this end, this command can
either be provided with an existing kubeconfig file via the option
"-c|--kubeconfig"; if no kubeconfig is provided, the command will generate a
temporary short-lived one with *cluster-admin* privileges which is deleted upon
successful add-on deployment.

================================= DISCLAIMER =================================
The manifests are not managed by Exoscale, therefore Exoscale cannot be held
responsible for any damage caused to an existing SKS cluster following
deployment of an add-on.
==============================================================================
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			cluster = args[0]
			addon   string
		)

		if len(args) == 1 {
			// TODO: display the list of available add-ons and exit.
			return nil
		} else {
			addon = args[1]
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		kubeconfig, err := cmd.Flags().GetString("kubeconfig")
		if err != nil {
			return err
		}

		// If no kubeconfig specified, generate a single-use one for this deployment.
		if kubeconfig == "" {
			if kubeconfig, err = labSKSFetchTempKubeconfig(zone, cluster); err != nil {
				return fmt.Errorf("unable to generate single-use kubeconfig: %s", err)
			}
		}

		fmt.Printf("Installing add-on %s\n", addon)
		fmt.Printf("kubectl --kubeconfig='%s' apply -f ... \n", kubeconfig)

		return nil
	},
}

func init() {
	sksLabAddonCmd.Flags().BoolP("force", "f", false,
		"Attempt to delete without prompting for confirmation")
	sksLabAddonCmd.Flags().BoolP("print", "p", false,
		"print kubectl command but don't execute it")
	sksLabAddonCmd.Flags().StringP("kubeconfig", "c", "",
		"kubeconfig file to use for deploying manifests.")
	sksLabAddonCmd.Flags().StringP("kubectl-options", "o", "",
		"Additional options to pass to the \"kubectl\" command (e.g. -o \"--v=4\")")
	sksLabAddonCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	labSKSCmd.AddCommand(sksLabAddonCmd)
}
