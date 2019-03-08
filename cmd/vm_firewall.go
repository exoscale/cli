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
	Use:   "set <vm name> <firewall name> [firewall name] ...",
	Short: "set the firewalls of a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getFirewallsByNameOrID(args[1:])
		if err != nil {
			return err
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}

		return setVirtualMachineFirewalls(vm, sgs)
	},
}

// vmFirewallAddCmd represents the firewall add command
var vmFirewallAddCmd = &cobra.Command{
	Use:   "add <vm name> <firewall name> [firewall name] ...",
	Short: "add firewalls to a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getFirewallsByNameOrID(args[1:])
		if err != nil {
			return err
		}

		vm, err := getVirtualMachineByNameOrID(args[0])
		if err != nil {
			return err
		}

		sgToAdd := make([]egoscale.SecurityGroup, 0)

	next:
		// Check if requested firewalls are not already set to the VM instance
		for i := range sgs {
			for j := range vm.SecurityGroup {
				if sgs[i].ID.Equal(*vm.SecurityGroup[j].ID) {
					continue next
				}
			}
			sgToAdd = append(sgToAdd, sgs[i])
		}

		return setVirtualMachineFirewalls(vm, append(vm.SecurityGroup, sgToAdd...))
	},
}

// vmFirewallRemoveCmd represents the firewall remove command
var vmFirewallRemoveCmd = &cobra.Command{
	Use:   "remove <vm name> <firewall name> [firewall name] ...",
	Short: "remove firewalls from a virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		sgs, err := getFirewallsByNameOrID(args[1:])
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

		return setVirtualMachineFirewalls(vm, sgRemaining)
	},
}

// setVirtualMachineFirewalls sets a virtual machine instance firewalls
func setVirtualMachineFirewalls(vm *egoscale.VirtualMachine, firewalls []egoscale.SecurityGroup) error {
	state := (string)(egoscale.VirtualMachineStopped)
	if vm.State != state {
		return fmt.Errorf("this operation is not permitted if your VM is not stopped")
	}

	ids := make([]egoscale.UUID, len(firewalls))
	for i := range firewalls {
		ids[i] = *firewalls[i].ID
	}

	resp, err := cs.RequestWithContext(gContext, &egoscale.UpdateVirtualMachine{
		ID:               vm.ID,
		SecurityGroupIDs: ids,
	})
	if err != nil {
		return err
	}

	vm, ok := resp.(*egoscale.VirtualMachine)
	if !ok {
		return fmt.Errorf("wrong type expected %q, got %T", "egoscale.VirtualMachine", resp)
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{vm.Name})
	sgs := getSecurityGroup(vm)
	table.Append([]string{"Security Groups", strings.Join(sgs, " - ")})
	table.Render()

	return nil
}

// getFirewallsByNameOrID tries to retrieve a list of firewalls by their name or ID.
func getFirewallsByNameOrID(firewalls []string) ([]egoscale.SecurityGroup, error) {
	sgs := make([]egoscale.SecurityGroup, len(firewalls))
	for i, s := range firewalls {
		sg, err := getSecurityGroupByNameOrID(s)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve firewall %q: %s", s, err)
		}
		sgs[i] = *sg
	}

	return sgs, nil
}

func init() {
	vmFirewallCmd.AddCommand(vmFirewallAddCmd)
	vmFirewallCmd.AddCommand(vmFirewallRemoveCmd)
	vmFirewallCmd.AddCommand(vmFirewallSetCmd)
	vmCmd.AddCommand(vmFirewallCmd)
}
