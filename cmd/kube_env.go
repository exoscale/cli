package cmd

import (
	"fmt"
	"os"

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

		clusterInstance, err := loadKubeData(clusterName, "instance")
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%q: no such cluster", clusterName)
			}
			return err
		}

		vm, err := getVMWithNameOrID(string(clusterInstance))
		if err != nil {
			return err
		}

		fmt.Printf(`
export KUBECONFIG="%s/kubeconfig"
export DOCKER_HOST="tcp://%s:2376"
export DOCKER_CERT_PATH="%s/docker"
export DOCKER_TLS_VERIFY=1
`,
			getKubeconfigPath(clusterName),
			vm.IP().String(),
			getKubeconfigPath(clusterName),
		)

		return nil
	},
}

func init() {
	kubeCmd.AddCommand(kubeEnvCmd)
}
