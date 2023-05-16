package cmd

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
)

var eipDeleteCmd = &cobra.Command{
	Use:     "delete IP-ADDRESS|ID",
	Short:   "Delete an Elastic IP",
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
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete Elastic IP %q?", cmd.ID.String())) {
					continue
				}
			}

			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("Deleting Elastic IP %q", cmd.ID.String()),
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
		resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, req)
		if err != nil {
			return nil, err
		}
		dissocReq.ID = resp.(*egoscale.IPAddress).ID
	}

	return dissocReq, nil
}

func init() {
	eipDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	eipCmd.AddCommand(eipDeleteCmd)
}
