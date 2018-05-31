package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Virtual machines management",
}

func getVMWithNameOrID(cs *egoscale.Client, name string) (*egoscale.VirtualMachine, error) {

	vm := &egoscale.VirtualMachine{ID: name}
	if err := cs.Get(vm); err == nil {
		return vm, err
	}

	vm.Name = name
	vm.ID = ""

	if err := cs.Get(vm); err != nil {
		return nil, fmt.Errorf("Unable to get Virtual Machine %q. %s", name, err)
	}
	return vm, nil
}

func getSecurityGroup(vm *egoscale.VirtualMachine) []string {
	sgs := []string{}
	for _, sgN := range vm.SecurityGroup {
		sgs = append(sgs, sgN.Name)
	}
	return sgs
}

func init() {
	rootCmd.AddCommand(vmCmd)
}
