package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// eipUpdateCmd represents the update command
var eipUpdateCmd = &cobra.Command{
	Use:   "update [eip ID]",
	Short: "update EIP",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		id, err := egoscale.ParseUUID(args[0])
		if err != nil {
			return err
		}
		interval, err := cmd.Flags().GetInt64("healthcheck-interval")
		if err != nil {
			return err
		}
		mode, err := cmd.Flags().GetString("healthcheck-mode")
		if err != nil {
			return err
		}
		path, err := cmd.Flags().GetString("healthcheck-path")
		if err != nil {
			return err
		}
		port, err := cmd.Flags().GetInt64("healthcheck-port")
		if err != nil {
			return err
		}
		strikesFail, err := cmd.Flags().GetInt64("healthcheck-strikes-fail")
		if err != nil {
			return err
		}
		strikesOK, err := cmd.Flags().GetInt64("healthcheck-strikes-ok")
		if err != nil {
			return err
		}
		timeout, err := cmd.Flags().GetInt64("healthcheck-timeout")
		if err != nil {
			return err
		}
		req := egoscale.UpdateIPAddress{
			Description:            description,
			HealthcheckInterval:    interval,
			HealthcheckMode:        mode,
			HealthcheckPath:        path,
			HealthcheckPort:        port,
			HealthcheckStrikesFail: strikesFail,
			HealthcheckStrikesOk:   strikesOK,
			HealthcheckTimeout:     timeout,
			ID:                     id,
		}

		return updateIPAddress(req)
	},
}

func updateIPAddress(associateIPAddress egoscale.UpdateIPAddress) error {
	resp, err := asyncRequest(
		associateIPAddress,
		fmt.Sprintf("Updating the IP address %q ", associateIPAddress.ID),
	)
	if err != nil {
		return err
	}

	ip := resp.(*egoscale.IPAddress)

	if !gQuiet {
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Zone", "IP", "Description", "ID"})
		table.Append([]string{
			ip.ZoneName,
			ip.IPAddress.String(),
			ip.Description,
			ip.ID.String()})
		table.Render()
	}

	return nil
}

func init() {
	eipUpdateCmd.Flags().StringP("description", "", "", "the IP address description.")
	eipUpdateCmd.Flags().Int64P("healthcheck-interval", "", 0, "the time in seconds to wait for between each healthcheck.")
	eipUpdateCmd.Flags().StringP("healthcheck-mode", "", "", "the healthcheck type. Should be tcp or http.")
	eipUpdateCmd.Flags().StringP("healthcheck-path", "", "", "the healthcheck path. Required if mode is http.")
	eipUpdateCmd.Flags().Int64P("healthcheck-port", "", 0, "the healthcheck port (e.g. 80 for http).")
	eipUpdateCmd.Flags().Int64P("healthcheck-strikes-fail", "", 0, "the number of times to retry before declaring the IP dead.")
	eipUpdateCmd.Flags().Int64P("healthcheck-strikes-ok", "", 0, "the number of times to retry before declaring the IP healthy.")
	eipUpdateCmd.Flags().Int64P("healthcheck-timeout", "", 0, "the timeout in seconds to wait for each check (default is 2). Should be lower than the interval.")
	eipCmd.AddCommand(eipUpdateCmd)
}
