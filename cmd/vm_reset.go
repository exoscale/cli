package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmResetCmd represents the stop command
var vmResetCmd = &cobra.Command{
	Use:   "reset <vm name> [vm name] ...",
	Short: "Reset virtual machine instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		diskValue, err := getInt64CustomFlag(cmd, "disk")
		if err != nil {
			return err
		}

		templateFilterCmd, err := cmd.Flags().GetString("template-filter")
		if err != nil {
			return err
		}
		templateFilter, err := validateTemplateFilter(templateFilterCmd)
		if err != nil {
			return err
		}

		template, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		errs := []error{}
		for _, v := range args {
			if err := resetVirtualMachine(v, diskValue, template, templateFilter, force); err != nil {
				errs = append(errs, fmt.Errorf("could not reset %q: %s", v, err))
			}
		}

		if len(errs) == 1 {
			return errs[0]
		}
		if len(errs) > 1 {
			var b strings.Builder
			for _, err := range errs {
				if _, e := fmt.Fprintln(&b, err); e != nil {
					return e
				}
			}
			return errors.New(b.String())
		}

		return nil
	},
}

// resetVirtualMachine stop a virtual machine instance
func resetVirtualMachine(vmName string, diskValue int64PtrValue, templateName string, templateFilter string, force bool) error {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return err
	}

	if !force {
		if !askQuestion(fmt.Sprintf("sure you want to reset %q virtual machine", vm.Name)) {
			return nil
		}
	}

	var template *egoscale.Template

	if templateName != "" {
		template, err = getTemplateByName(vm.ZoneID, templateName, templateFilter)
		if err != nil {
			return err
		}
	} else {
		resp, err := cs.ListWithContext(gContext, egoscale.Template{
			IsFeatured: true,
			ID:         vm.TemplateID,
			ZoneID:     vm.ZoneID,
		})

		if err != nil {
			return err
		}

		if len(resp) > 0 {
			template = resp[0].(*egoscale.Template)
		}
	}

	resp, err := cs.GetWithContext(gContext, egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})
	if err != nil {
		return err
	}

	volume := resp.(*egoscale.Volume)
	volumeSize := volume.Size >> 30

	rootDiskSize := int64(volumeSize)

	if diskValue.int64 != nil {
		if template != nil && *diskValue.int64 < (template.Size>>30) {
			return fmt.Errorf("root disk size must be greater or equal than %dGB", template.Size>>30)
		}

		rootDiskSize = *diskValue.int64
	}

	cmd := &egoscale.RestoreVirtualMachine{
		VirtualMachineID: vm.ID,
		RootDiskSize:     rootDiskSize,
	}
	if template != nil {
		cmd.TemplateID = template.ID

		fmt.Printf("Resetting %q using %q", vm.Name, template.DisplayText)
	} else {
		fmt.Printf("Resetting %q ", vm.Name)
	}
	var errorReq error
	cs.AsyncRequestWithContext(gContext, cmd, func(jobResult *egoscale.AsyncJobResult, err error) bool {
		fmt.Print(".")

		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus == egoscale.Success {
			fmt.Println(" success.")
			return false
		}

		return true
	})

	if errorReq != nil {
		fmt.Println(" failure!")
	}

	return errorReq
}

func init() {
	vmCmd.AddCommand(vmResetCmd)
	diskSizeVarP := new(int64PtrValue)
	vmResetCmd.Flags().VarP(diskSizeVarP, "disk", "d", "New disk size after reset in GB")
	vmResetCmd.Flags().StringP("template", "t", "", fmt.Sprintf("<template name | id> (default: %s)", defaultTemplate))
	vmResetCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	vmResetCmd.Flags().BoolP("force", "f", false, "Attempt to reset vitual machine without prompting for confirmation")
}
