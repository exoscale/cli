package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type vmShowOutput struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	CreationDate       string   `json:"creation_date"`
	Size               string   `json:"size"`
	DiskSize           string   `json:"disk_size"`
	Template           string   `json:"template"`
	Zone               string   `json:"zone"`
	State              string   `json:"state"`
	IPAddress          string   `json:"ip_address"`
	ReverseDNS         string   `json:"reverse_dns"`
	Username           string   `json:"username"`
	SSHKey             string   `json:"ssh_key"`
	SecurityGroups     []string `json:"security_groups,omitempty"`
	AntiAffinityGroups []string `json:"antiaffinity_groups,omitempty" outputLabel:"Anti-Affinity Groups"`
	PrivateNetworks    []string `json:"private_networks,omitempty"`
}

func (o *vmShowOutput) Type() string { return "Instance" }
func (o *vmShowOutput) ToJSON()      { output.JSON(o) }
func (o *vmShowOutput) ToText()      { output.Text(o) }
func (o *vmShowOutput) ToTable()     { output.Table(o) }

func init() {
	vmShowCmd := &cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show a Compute instance details",
		Long: fmt.Sprintf(`This command shows a Compute instance details.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&vmShowOutput{}), ", ")),
		Aliases:           gShowAlias,
		ValidArgsFunction: completeVMNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			userDataOnly, err := cmd.Flags().GetBool("user-data")
			if err != nil {
				return err
			}
			if userDataOnly {
				return showVMUserData(args[0])
			}

			return printOutput(showVM(args[0]))
		},
	}

	vmShowCmd.Flags().Bool("user-data", false, "Show current cloud-init user data configuration")
	vmCmd.AddCommand(vmShowCmd)
}

func showVM(name string) (output.Outputter, error) {
	vm, err := getVirtualMachineByNameOrID(name)
	if err != nil {
		return nil, err
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, &egoscale.Template{
		IsFeatured: true,
		ID:         vm.TemplateID,
		ZoneID:     vm.ZoneID,
	})
	if err != nil {
		return nil, err
	}
	template := resp.(*egoscale.Template)

	resp, err = globalstate.EgoscaleClient.GetWithContext(gContext, &egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})
	if err != nil {
		return nil, err
	}
	volume := resp.(*egoscale.Volume)

	reverseDNS, err := globalstate.EgoscaleClient.RequestWithContext(gContext, &egoscale.QueryReverseDNSForVirtualMachine{ID: vm.ID})
	if err != nil {
		return nil, err
	}

	out := vmShowOutput{
		ID:           vm.ID.String(),
		Name:         vm.DisplayName,
		CreationDate: vm.Created,
		Size:         vm.ServiceOfferingName,
		Template:     vm.TemplateName,
		Zone:         vm.ZoneName,
		State:        vm.State,
		DiskSize:     humanize.IBytes(volume.Size),
		IPAddress:    vm.IP().String(),
		ReverseDNS: func(vm *egoscale.VirtualMachine) string {
			if len(vm.DefaultNic().ReverseDNS) > 0 {
				return vm.DefaultNic().ReverseDNS[0].DomainName
			}
			return "n/a"
		}(reverseDNS.(*egoscale.VirtualMachine)),
		Username:           "n/a",
		SSHKey:             vm.KeyPair,
		SecurityGroups:     make([]string, len(vm.SecurityGroup)),
		AntiAffinityGroups: make([]string, len(vm.AffinityGroup)),
		PrivateNetworks:    make([]string, 0),
	}

	for i, sg := range vm.SecurityGroup {
		out.SecurityGroups[i] = sg.Name
	}

	for i, aag := range vm.AffinityGroup {
		out.AntiAffinityGroups[i] = aag.Name
	}

	for _, nic := range vm.Nic {
		if nic.IsDefault {
			continue
		}

		out.PrivateNetworks = append(out.PrivateNetworks, nic.NetworkName)
	}

	if username, ok := template.Details["username"]; ok {
		out.Username = username
	}

	// If a single-use SSH keypair has been created for this instance,
	// report the private key file location instead of the API SSH key name.
	sshKeyPath := getKeyPairPath(vm.ID.String())
	if _, err := os.Stat(sshKeyPath); err == nil && out.SSHKey == "" {
		out.SSHKey = sshKeyPath
	}

	return &out, nil
}

func showVMUserData(name string) error {
	vm, err := getVirtualMachineByNameOrID(name)
	if err != nil {
		return err
	}

	resp, err := globalstate.EgoscaleClient.SyncRequestWithContext(gContext, &egoscale.GetVirtualMachineUserData{
		VirtualMachineID: vm.ID,
	})
	if err != nil {
		return err
	}

	userData, err := decodeUserData(resp.(*egoscale.VirtualMachineUserData).UserData)
	if err != nil {
		return err
	}

	fmt.Println(userData)

	return nil
}
