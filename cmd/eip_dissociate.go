package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"

	"github.com/exoscale/egoscale"
)

var eipDissociateCmd = &cobra.Command{
	Use:     "dissociate IP-ADDRESS INSTANCE-NAME|ID",
	Short:   "Dissociate an Elastic IP from a Compute instance",
	Aliases: gDissociateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args[1:]))

		ipAddr := args[0]
		ip := net.ParseIP(ipAddr)
		if ip == nil {
			return fmt.Errorf("invalid IP address %q", ipAddr)
		}

		for _, arg := range args[1:] {
			vm, err := getVirtualMachineByNameOrID(arg)
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to dissociate %q EIP from %q?", ip.String(), vm.Name)) {
					continue
				}
			}

			cmd, err := prepareDissociateIP(vm, ip)
			if err != nil {
				return err
			}
			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("dissociate %q EIP", cmd.ID.String()),
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

func prepareDissociateIP(vm *egoscale.VirtualMachine, ip net.IP) (*egoscale.RemoveIPFromNic, error) {
	defaultNic := vm.DefaultNic()
	if defaultNic == nil {
		return nil, fmt.Errorf("the instance %q has no default NIC", vm.ID)
	}

	eipID, err := getSecondaryIP(defaultNic, ip)
	if err != nil {
		return nil, err
	}

	return &egoscale.RemoveIPFromNic{ID: eipID}, nil
}

func getSecondaryIP(nic *egoscale.Nic, ip net.IP) (*egoscale.UUID, error) {
	for _, sIP := range nic.SecondaryIP {
		if sIP.IPAddress.Equal(ip) {
			return sIP.ID, nil
		}
	}
	return nil, fmt.Errorf("Elastic IP %q not found", ip) // nolint
}

func init() {
	eipDissociateCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	eipCmd.AddCommand(eipDissociateCmd)
}
