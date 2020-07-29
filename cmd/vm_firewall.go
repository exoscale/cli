package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmFirewallCmd represents the vm firewall command
var vmFirewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Virtual machines firewall management",
}

// vmFirewallSetCmd represents the firewall set command
var vmFirewallSetCmd = &cobra.Command{
	Use:   "set <vm name | id> <SG name | id>+",
	Short: "Set the security groups of a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getSecurityGroupsByNameOrID(args[1:])
		if err != nil {
			return err
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}

		if err := setVirtualMachineSecurityGroups(vm, sgs); err != nil {
			return err
		}

		return printVirtualMachineSecurityGroups(vm)
	},
}

// vmFirewallAddCmd represents the firewall add command
var vmFirewallAddCmd = &cobra.Command{
	Use:   "add <vm name | id> <SG name | id>+",
	Short: "Add security groups to a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getSecurityGroupsByNameOrID(args[1:])
		if err != nil {
			return err
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}

		sgToAdd := make([]egoscale.SecurityGroup, 0)

	next:
		// Check if requested security groups are not already set to the VM instance
		for i := range sgs {
			for j := range vm.SecurityGroup {
				if sgs[i].ID.Equal(*vm.SecurityGroup[j].ID) {
					continue next
				}
			}
			sgToAdd = append(sgToAdd, sgs[i])
		}

		if err := setVirtualMachineSecurityGroups(vm, append(vm.SecurityGroup, sgToAdd...)); err != nil {
			return err
		}

		return printVirtualMachineSecurityGroups(vm)
	},
}

// vmFirewallRemoveCmd represents the firewall remove command
var vmFirewallRemoveCmd = &cobra.Command{
	Use:   "remove <vm name | id> <SG name | id>+",
	Short: "Remove security groups from a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getSecurityGroupsByNameOrID(args[1:])
		if err != nil {
			return err
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}

		sgRemaining := make([]egoscale.SecurityGroup, 0)

	next:
		for i := range vm.SecurityGroup {
			for j := range sgs {
				if sgs[j].ID.Equal(*vm.SecurityGroup[i].ID) {
					continue next
				}
			}
			sgRemaining = append(sgRemaining, vm.SecurityGroup[i])
		}

		if err := setVirtualMachineSecurityGroups(vm, sgRemaining); err != nil {
			return err
		}

		return printVirtualMachineSecurityGroups(vm)
	},
}

// setVirtualMachineSecurityGroups sets a virtual machine instance security groups.
func setVirtualMachineSecurityGroups(vm *egoscale.VirtualMachine, sgs []egoscale.SecurityGroup) error {
	state := (string)(egoscale.VirtualMachineStopped)
	if vm.State != state {
		return fmt.Errorf("this operation is not permitted if your VM is not stopped")
	}

	ids := make([]egoscale.UUID, len(sgs))
	for i := range sgs {
		ids[i] = *sgs[i].ID
	}

	_, err := cs.RequestWithContext(gContext, &egoscale.UpdateVirtualMachine{
		ID:               vm.ID,
		SecurityGroupIDs: ids,
	})
	if err != nil {
		return err
	}

	return nil
}

// printVirtualMachineSecurityGroups prints a virtual machine instance security groups to standard output.
func printVirtualMachineSecurityGroups(vm *egoscale.VirtualMachine) error {
	if !gQuiet {
		// Refresh the vm object to ensure its properties are up-to-date
		vm, err := getVirtualMachineByNameOrID(vm.ID.String())
		if err != nil {
			return err
		}

		sgs := []string{}
		for _, sgN := range vm.SecurityGroup {
			sgs = append(sgs, sgN.Name)
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{vm.Name})
		table.Append([]string{"Security Groups", strings.Join(sgs, " - ")})
		table.Render()
	}

	return nil
}

// getSecurityGroupsByNameOrID tries to retrieve a list of security groups by their name or ID.
func getSecurityGroupsByNameOrID(sgs []string) ([]egoscale.SecurityGroup, error) {
	res := make([]egoscale.SecurityGroup, len(sgs))
	for i, s := range sgs {
		sg, err := getSecurityGroupByNameOrID(s)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve security group %q: %s", s, err)
		}
		res[i] = *sg
	}

	return res, nil
}

func init() {
	vmFirewallCmd.AddCommand(vmFirewallAddCmd)
	vmFirewallCmd.AddCommand(vmFirewallRemoveCmd)
	vmFirewallCmd.AddCommand(vmFirewallSetCmd)
	vmCmd.AddCommand(vmFirewallCmd)
}
