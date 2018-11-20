package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	// "github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var (
	// kubeSecurityGroup represents the firewall security group to add k8s VM instances into
	kubeSecurityGroup = "exokube"

	// kubeInstanceTagKey represents the VM instance grouping tag key
	kubeInstanceTagKey = "exokube"

	// kubeInstanceTagValue represents the VM instance grouping tag value
	kubeInstanceTagValue = "true"
)

// kubeCmd represents the kube command
var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Standalone Kubernetes clusters management",
	Long: `These commands allow you to bootstrap standalone Kubernetes cluster
instances in a similar fashion as Minikube. It runs a single-node Kubernetes
cluster inside an Exoscale VM for users looking to try out Kubernetes or develop
with it day-to-day.`,
}

func getKubeconfigPath(clusterName string) string {
	return path.Join(gConfigFolder, "kube", clusterName)
}

func saveKubeData(clusterName, key string, data []byte) error {
	if _, err := os.Stat(getKubeconfigPath(clusterName)); os.IsNotExist(err) {
		if err := os.MkdirAll(getKubeconfigPath(clusterName), os.ModePerm); err != nil {
			return fmt.Errorf("unable to create directory: %s", err)
		}
	}

	if err := ioutil.WriteFile(path.Join(getKubeconfigPath(clusterName), key), data, 0600); err != nil {
		return fmt.Errorf("unable to write file: %s", err)
	}

	return nil
}

func loadKubeData(clusterName, key string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(getKubeconfigPath(clusterName), key))
}

func deleteKubeData(clusterName string) {
	folder := getKubeconfigPath(clusterName)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			log.Fatalf("Kubernetes cluster configuration could not be deleted: %s", err)
		}
	}
}

func init() {
	RootCmd.AddCommand(kubeCmd)
}
