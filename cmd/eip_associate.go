package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

var eipAssociateCmd = &cobra.Command{
	Use:     "associate IP-ADDRESS INSTANCE-NAME|ID",
	Short:   "Associate an Elastic IP to a Compute instance",
	Aliases: gAssociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		tasks := make([]task, 0, len(args[1:]))

		ipaddr := args[0]
		ip := net.ParseIP(ipaddr)
		if ip == nil {
			return fmt.Errorf("invalid IP address %q", ipaddr)
		}

		for _, arg := range args[1:] {
			vm, err := getVirtualMachineByNameOrID(arg)
			if err != nil {
				return err
			}

			cmd, err := prepareAssociateIP(vm, ip)
			if err != nil {
				return err
			}
			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Associating Elastic IP %s to %s", cmd.IPAddress.String(), vm.Name),
			})
		}
		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}

func prepareAssociateIP(vm *egoscale.VirtualMachine, ip net.IP) (*egoscale.AddIPToNic, error) {
	defaultNic := vm.DefaultNic()
	if defaultNic == nil {
		return nil, fmt.Errorf("the Compute instance %q has no default NIC", vm.ID)
	}

	return &egoscale.AddIPToNic{NicID: defaultNic.ID, IPAddress: ip}, nil
}

func init() {
	eipCmd.AddCommand(eipAssociateCmd)
}
