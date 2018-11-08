package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// envCmd represents the env command
var kubeEnvCmd = &cobra.Command{
	Use:   "env <cluster name>",
	Short: "Print a standalone Kubernetes cluster's environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		clusterName := args[0]

		fmt.Printf("export KUBECONFIG=\"%s/%s\"\n", getKubeconfigPath(clusterName), "kubeconfig")

		return nil
	},
}

func init() {
	kubeCmd.AddCommand(kubeEnvCmd)
}
