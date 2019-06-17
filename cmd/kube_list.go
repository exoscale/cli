package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type kubeListItemOutput struct {
	Name          string `json:"name"`
	IPAddress     string `json:"ipaddress"`
	State         string `json:"state"`
	Size          string `json:"size"`
	K8sVersion    string `json:"k8s_version" outputLabel:"k8s Version"`
	DockerVersion string `json:"docker_version"`
	CalicoVersion string `json:"calico_version"`
}

type kubeListOutput []kubeListItemOutput

func (o *kubeListOutput) toJSON()  { outputJSON(o) }
func (o *kubeListOutput) toText()  { outputText(o) }
func (o *kubeListOutput) toTable() { outputTable(o) }

func init() {
	kubeCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List running Kubernetes cluster instances",
		Long: fmt.Sprintf(`This command lists existing Kubernetes cluster instances.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&kubeListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listKubeInstances())
		},
	})
}

func listKubeInstances() (outputter, error) {
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{
		Tags: []egoscale.ResourceTag{{
			Key:   "managedby",
			Value: "exokube",
		}},
	})
	if err != nil {
		return nil, err
	}

	out := kubeListOutput{}

	for _, key := range vms {
		vm := key.(*egoscale.VirtualMachine)

		out = append(out, kubeListItemOutput{
			Name:          vm.Name,
			IPAddress:     vm.IP().String(),
			State:         vm.State,
			Size:          vm.ServiceOfferingName,
			K8sVersion:    getKubeInstanceVersion(vm, kubeTagKubernetes),
			DockerVersion: getKubeInstanceVersion(vm, kubeTagDocker),
			CalicoVersion: getKubeInstanceVersion(vm, kubeTagCalico),
		})
	}

	return &out, nil
}
