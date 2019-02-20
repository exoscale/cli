package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var eipDeleteCmd = &cobra.Command{
	Use:     "delete <ip | eip id>",
	Short:   "Delete EIP",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		tasks := make([]task, 0, len(args))

		for _, arg := range args {
			cmd, err := prepareDeleteEip(arg)
			if err != nil {
				return err
			}

			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to delete %q EIP", cmd.ID.String())) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Remove %q EIP", cmd.ID.String()),
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

// Builds the disassociateIPAddress request
func prepareDeleteEip(ip string) (*egoscale.DisassociateIPAddress, error) {
	dissocReq := &egoscale.DisassociateIPAddress{}

	ipAddr := net.ParseIP(ip)

	if ipAddr == nil {
		id, err := egoscale.ParseUUID(ip)
		if err != nil {
			return nil, fmt.Errorf("delete the eip by ID or IP address, gotb %q", ip)
		}
		dissocReq.ID = id
	} else {
		req := &egoscale.IPAddress{IPAddress: ipAddr, IsElastic: true}
		resp, err := cs.GetWithContext(gContext, req)
		if err != nil {
			return nil, err
		}
		dissocReq.ID = resp.(*egoscale.IPAddress).ID
	}

	return dissocReq, nil
}

func init() {
	eipDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove EIP without prompting for confirmation")
	eipCmd.AddCommand(eipDeleteCmd)
}
