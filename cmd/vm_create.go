package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
)

var vmCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Deploy a Compute instance",
	Long: fmt.Sprintf(`This command deploys a new Compute instance.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vmShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		vmName := args[0]

		userDataPath, err := cmd.Flags().GetString("cloud-init-file")
		if err != nil {
			return err
		}
		userDataCompress, err := cmd.Flags().GetBool("cloud-init-compress")
		if err != nil {
			return err
		}
		userData := ""
		if userDataPath != "" {
			userData, err = getUserDataFromFile(userDataPath, userDataCompress)
			if err != nil {
				return err
			}
		}

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByNameOrID(zoneName)
		if err != nil {
			return err
		}

		templateFilter, err := cmd.Flags().GetString("template-filter")
		if err != nil {
			return err
		}
		if templateFilter, err = validateTemplateFilter(templateFilter); err != nil {
			return err
		}

		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		template, err := getTemplateByNameOrID(zone.ID, templateName, templateFilter)
		if err != nil {
			return err
		}

		diskSize, err := cmd.Flags().GetInt64("disk")
		if err != nil {
			return err
		}

		keypair, err := cmd.Flags().GetString("keypair")
		if err != nil {
			return err
		}
		if keypair == "" {
			keypair = account.CurrentAccount.DefaultSSHKey
		}

		sg, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}

		sgs, err := getSecurityGroupIDs(sg)
		if err != nil {
			return err
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		privnets, err := cmd.Flags().GetStringSlice("privnet")
		if err != nil {
			return err
		}

		pvs, err := getPrivnetIDs(privnets, zone.ID)
		if err != nil {
			return err
		}

		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		servOffering, err := getServiceOfferingByNameOrID(so)
		if err != nil {
			return err
		}

		antiAffinityGroups, err := cmd.Flags().GetStringSlice("anti-affinity-group")
		if err != nil {
			return err
		}

		aags, err := getAffinityGroupIDs(antiAffinityGroups)
		if err != nil {
			return err
		}

		vm, err := createVM(&egoscale.DeployVirtualMachine{
			Name:              vmName,
			UserData:          userData,
			ZoneID:            zone.ID,
			TemplateID:        template.ID,
			RootDiskSize:      diskSize,
			KeyPair:           keypair,
			SecurityGroupIDs:  sgs,
			IP6:               &ipv6,
			NetworkIDs:        pvs,
			ServiceOfferingID: servOffering.ID,
			AffinityGroupIDs:  aags,
		})
		if err != nil {
			return err
		}

		if !globalstate.Quiet {
			return printOutput(showVM(vm.ID.String()))
		}

		return nil
	},
}

func createVM(deployVM *egoscale.DeployVirtualMachine) (*egoscale.VirtualMachine, error) {
	var (
		sshKey          *egoscale.SSHKeyPair
		singleUseSSHKey bool
	)

	// If not SSH key is specified, create a single-use SSH key, store the private key locally
	// and delete the public key from the API once the Instance has been deployed.
	if deployVM.KeyPair == "" {
		singleUseSSHKey = true

		if !globalstate.Quiet {
			fmt.Fprintln(os.Stderr, "Creating single-use SSH key")
		}

		keyName, err := utils.RandStringBytes(64)
		if err != nil {
			return nil, err
		}

		sshKey, err = createSSHKey(keyName)
		if err != nil {
			return nil, fmt.Errorf("error creating single-use SSH keypair: %s", err)
		}
		deployVM.KeyPair = sshKey.Name

		defer deleteSSHKey(sshKey.Name) // nolint: errcheck
	}

	resp := asyncTasks([]task{{deployVM, fmt.Sprintf("Deploying %q", deployVM.Name)}})
	errors := filterErrors(resp)
	if len(errors) > 0 {
		return nil, errors[0]
	}
	vm := resp[0].resp.(*egoscale.VirtualMachine)

	if singleUseSSHKey {
		saveKeyPair(sshKey, *vm.ID)
	}

	return vm, nil
}

func init() {
	vmCreateCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	vmCreateCmd.Flags().StringP("template", "t", defaultTemplate, "template NAME|ID")
	vmCreateCmd.Flags().StringP("template-filter", "", defaultTemplateFilter, templateFilterHelp)
	vmCreateCmd.Flags().StringP("service-offering", "o", defaultServiceOffering, serviceOfferingHelp)
	vmCreateCmd.Flags().Int64P("disk", "d", 50, "disk size")
	vmCreateCmd.Flags().StringP("keypair", "k", "", "SSH keypair name. If not specified, a single-use SSH key will be created.")
	vmCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security Group NAME|ID. Can be specified multiple times.")
	vmCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Private Network NAME|ID. Can be specified multiple times.")
	vmCreateCmd.Flags().StringSliceP("anti-affinity-group", "a", nil, "Anti-Affinity Group NAME|ID. Can be specified multiple times.")
	vmCreateCmd.Flags().StringP("cloud-init-file", "f", "", "instance cloud-init userdata")
	vmCreateCmd.Flags().BoolP("cloud-init-compress", "", false, "compress instance cloud-init user data")
	vmCreateCmd.Flags().BoolP("ipv6", "6", false, "enable IPv6")
	vmCmd.AddCommand(vmCreateCmd)
}
