package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
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
	Username           string   `json:"username"`
	SSHKey             string   `json:"ssh_key"`
	SecurityGroups     []string `json:"security_groups,omitempty"`
	AntiAffinityGroups []string `json:"antiaffinity_groups,omitempty" outputLabel:"Anti-Affinity Groups"`
	PrivateNetworks    []string `json:"private_networks,omitempty"`
}

func (o *vmShowOutput) Type() string { return "Instance" }
func (o *vmShowOutput) toJSON()      { outputJSON(o) }
func (o *vmShowOutput) toText()      { outputText(o) }
func (o *vmShowOutput) toTable()     { outputTable(o) }

func init() {
	vmShowCmd := &cobra.Command{
		Use:   "show <name | id>",
		Short: "Show a virtual machine details",
		Long: fmt.Sprintf(`This command shows a Compute instance details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&vmShowOutput{}), ", ")),
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

			return output(showVM(args[0]))
		},
	}

	vmShowCmd.Flags().Bool("user-data", false, "Show current cloud-init user data configuration")
	vmCmd.AddCommand(vmShowCmd)
}

func showVM(name string) (outputter, error) {
	vm, err := getVirtualMachineByNameOrID(name)
	if err != nil {
		return nil, err
	}

	resp, err := cs.GetWithContext(gContext, &egoscale.Template{
		IsFeatured: true,
		ID:         vm.TemplateID,
		ZoneID:     vm.ZoneID,
	})
	if err != nil {
		return nil, err
	}
	template := resp.(*egoscale.Template)

	resp, err = cs.GetWithContext(gContext, &egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})
	if err != nil {
		return nil, err
	}

	volume := resp.(*egoscale.Volume)

	out := vmShowOutput{
		ID:                 vm.ID.String(),
		Name:               vm.DisplayName,
		CreationDate:       vm.Created,
		Size:               vm.ServiceOfferingName,
		Template:           vm.TemplateName,
		Zone:               vm.ZoneName,
		State:              vm.State,
		DiskSize:           humanize.IBytes(volume.Size),
		IPAddress:          vm.IP().String(),
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

	resp, err := cs.SyncRequestWithContext(gContext, &egoscale.GetVirtualMachineUserData{
		VirtualMachineID: vm.ID,
	})
	if err != nil {
		return err
	}

	userData, err := base64.StdEncoding.DecodeString(resp.(*egoscale.VirtualMachineUserData).UserData)
	if err != nil {
		return err
	}

	fmt.Println(string(userData))

	return nil
}
