package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/table"
)

var vmFirewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Virtual machines firewall management",
}

var vmFirewallSetCmd = &cobra.Command{
	Use:   "set INSTANCE-NAME|ID SECURITY-GROUP-NAME|ID...",
	Short: "Set the Security Groups for a Compute instance",
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

var vmFirewallAddCmd = &cobra.Command{
	Use:   "add INSTANCE-NAME|ID SECURITY-GROUP-NAME|ID...",
	Short: "Add Security Groups to a Compute instance",
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
		// Check if requested Security Groups are not already set to the VM instance
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

var vmFirewallRemoveCmd = &cobra.Command{
	Use:   "remove INSTANCE-NAME|ID SECURITY-GROUP-NAME|ID...",
	Short: "Remove Security Groups from a Compute instance",
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

// setVirtualMachineSecurityGroups sets a Compute instance Security Groups.
func setVirtualMachineSecurityGroups(vm *egoscale.VirtualMachine, sgs []egoscale.SecurityGroup) error {
	ids := make([]egoscale.UUID, len(sgs))
	for i := range sgs {
		ids[i] = *sgs[i].ID
	}

	_, err := cs.RequestWithContext(gContext, &egoscale.UpdateVirtualMachineSecurityGroups{
		ID:               vm.ID,
		SecurityGroupIDs: ids,
	})
	if err != nil {
		return err
	}

	return nil
}

// printVirtualMachineSecurityGroups prints a Compute instance Security Groups to standard output.
func printVirtualMachineSecurityGroups(vm *egoscale.VirtualMachine) error {
	if !globalstate.Quiet {
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

// getSecurityGroupsByNameOrID tries to retrieve a list of Security Groups by their name or ID.
func getSecurityGroupsByNameOrID(sgs []string) ([]egoscale.SecurityGroup, error) {
	res := make([]egoscale.SecurityGroup, len(sgs))
	for i, s := range sgs {
		sg, err := getSecurityGroupByNameOrID(s)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve Security Group %q: %s", s, err)
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
