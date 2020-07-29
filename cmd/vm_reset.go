package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmResetCmd represents the stop command
var vmResetCmd = &cobra.Command{
	Use:               "reset <vm name | id>+",
	Short:             "Reset virtual machine instance",
	ValidArgsFunction: completeVMNames,
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

		tasks := make([]task, 0, len(args))
		for _, v := range args {
			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to reset %q virtual machine", v)) {
					continue
				}
			}

			task, err := makeResetVirtualMachineCMD(v, diskValue, template, templateFilter)
			if err != nil {
				return err
			}

			tasks = append(tasks, *task)
		}

		taskResponses := asyncTasks(tasks)
		errors := filterErrors(taskResponses)
		if len(errors) > 0 {
			return errors[0]
		}

		return nil
	},
}

func makeResetVirtualMachineCMD(vmName string, diskValue int64PtrValue, templateName string, templateFilter string) (*task, error) {
	vm, err := getVirtualMachineByNameOrID(vmName)
	if err != nil {
		return nil, err
	}

	var template *egoscale.Template

	if templateName != "" {
		template, err = getTemplateByName(vm.ZoneID, templateName, templateFilter)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := cs.ListWithContext(gContext, egoscale.Template{
			IsFeatured: true,
			ID:         vm.TemplateID,
			ZoneID:     vm.ZoneID,
		})

		if err != nil {
			return nil, err
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
		return nil, err
	}

	volume := resp.(*egoscale.Volume)
	volumeSize := volume.Size >> 30

	rootDiskSize := int64(volumeSize)

	if diskValue.int64 != nil {
		if template != nil && *diskValue.int64 < (template.Size>>30) {
			return nil, fmt.Errorf("root disk size must be greater or equal than %dGB", template.Size>>30)
		}

		rootDiskSize = *diskValue.int64
	}

	cmd := &egoscale.RestoreVirtualMachine{
		VirtualMachineID: vm.ID,
		RootDiskSize:     rootDiskSize,
	}

	var msg string
	if template != nil {
		cmd.TemplateID = template.ID

		msg = fmt.Sprintf("Resetting %q using %q", vm.Name, template.DisplayText)
	} else {
		msg = fmt.Sprintf("Resetting %q ", vm.Name)
	}

	return &task{cmd, msg}, nil
}

func init() {
	vmCmd.AddCommand(vmResetCmd)
	diskSizeVarP := new(int64PtrValue)
	vmResetCmd.Flags().VarP(diskSizeVarP, "disk", "d", "New disk size after reset in GB")
	vmResetCmd.Flags().StringP("template", "t", "", fmt.Sprintf("<template name | id> (default: %s)", defaultTemplate))
	vmResetCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	vmResetCmd.Flags().BoolP("force", "f", false, "Attempt to reset vitual machine without prompting for confirmation")
}
