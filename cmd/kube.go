package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var (
	// kubeSecurityGroup represents the firewall Security Group to add k8s VM instances into
	kubeSecurityGroup = "exokube"

	// kubeTagKubernetes represents the name of the tag used to store the version of Kubernetes
	kubeTagKubernetes = "exokube:kubernetes"

	// kubeTagCalico represents the name of the tag used to store the version of Calico
	kubeTagCalico = "exokube:calico"

	// kubeTagDocker represents the name of the tag used to store the version of Docker
	kubeTagDocker = "exokube:docker"
)

var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Standalone Kubernetes cluster management",
	Long: `These commands allow you to bootstrap standalone Kubernetes cluster
instances in a similar fashion as Minikube. It runs a single-node Kubernetes
cluster inside an Exoscale VM for users looking to try out Kubernetes or develop
with it day-to-day.`,
}

func getKubeInstanceVersion(vm *egoscale.VirtualMachine, tagName string) string {
	for _, tag := range vm.Tags {
		if tag.Key == tagName {
			return tag.Value
		}
	}

	return ""
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

	if err := ioutil.WriteFile(path.Join(getKubeconfigPath(clusterName), key), data, 0o600); err != nil {
		return fmt.Errorf("unable to write file: %s", err)
	}

	return nil
}

func getKubeVM(clusterName string) (*egoscale.VirtualMachine, error) {
	filename := path.Join(getKubeconfigPath(clusterName), "instance")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%q: no such cluster", clusterName)
		}
		return nil, err
	}

	return getVirtualMachineByNameOrID(string(string(content)))
}

func deleteKubeData(clusterName string) error {
	folder := getKubeconfigPath(clusterName)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		if err := os.RemoveAll(folder); err != nil {
			return fmt.Errorf("the Kubernetes cluster configuration could not be deleted: %s", err)
		}
	}

	return nil
}

func init() {
	labCmd.AddCommand(kubeCmd)
}
