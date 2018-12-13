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

func getVirtualMachineByNameOrID(name string) (*egoscale.VirtualMachine, error) {
	vmQuery := egoscale.VirtualMachine{}
	id, err := egoscale.ParseUUID(name)
	if err != nil {
		vmQuery.Name = name
	} else {
		vmQuery.ID = id
	}

	vms, err := cs.ListWithContext(gContext, vmQuery)
	if err != nil {
		return nil, err
	}

	var vm *egoscale.VirtualMachine
	switch len(vms) {
	case 0:
		return nil, fmt.Errorf("no VMs has been found")
	case 1:
		vm = vms[0].(*egoscale.VirtualMachine)
	default:
		names := []string{}
		for _, i := range vms {
			v := i.(*egoscale.VirtualMachine)
			if v.Name != vmQuery.Name {
				continue
			}

			vm = v
			names = append(names, fmt.Sprintf("\t%s\t%s\t%s", v.ID.String(), v.ZoneName, v.IP()))
		}

		if len(names) == 1 {
			break
		}

		fmt.Println("More than one VM has been found, use the ID instead:")
		for _, name := range names {
			fmt.Println(name)
		}
		return nil, fmt.Errorf("abort vm name %q is ambiguous", vmQuery.Name)
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
	RootCmd.AddCommand(vmCmd)
}
