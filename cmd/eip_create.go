package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var eipCreateCmd = &cobra.Command{
	Use:     "create [zone name | zone id]",
	Short:   "Create EIP",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone := gCurrentAccount.DefaultZone
		if len(args) >= 1 {
			zone = args[0]
		}
		description, err := cmd.Flags().GetString("description")
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
		req := egoscale.AssociateIPAddress{
			Description:            description,
			HealthcheckInterval:    interval,
			HealthcheckMode:        mode,
			HealthcheckPath:        path,
			HealthcheckPort:        port,
			HealthcheckStrikesFail: strikesFail,
			HealthcheckStrikesOk:   strikesOK,
			HealthcheckTimeout:     timeout,
		}

		return associateIPAddress(req, zone)
	},
}

func associateIPAddress(associateIPAddress egoscale.AssociateIPAddress, zone string) error {
	zoneID, err := getZoneIDByName(zone)
	if err != nil {
		return err
	}
	associateIPAddress.ZoneID = zoneID

	resp, err := cs.RequestWithContext(gContext, associateIPAddress)
	if err != nil {
		return err
	}

	ipResp := resp.(*egoscale.IPAddress)

	if !gQuiet {
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Description", "ID", "IP", "Zone"})
		table.Append([]string{
			ipResp.Description,
			ipResp.ID.String(),
			ipResp.IPAddress.String(),
			ipResp.ZoneName})
		table.Render()
	}

	return nil
}

func init() {
	eipCreateCmd.Flags().StringP("description", "", "", "the IP address description.")
	eipCreateCmd.Flags().Int64P("healthcheck-interval", "", 0, "the time in seconds to wait for between each healthcheck.")
	eipCreateCmd.Flags().StringP("healthcheck-mode", "", "", "the healthcheck type. Should be tcp or http.")
	eipCreateCmd.Flags().StringP("healthcheck-path", "", "", "the healthcheck path. Required if mode is http.")
	eipCreateCmd.Flags().Int64P("healthcheck-port", "", 0, "the healthcheck port (e.g. 80 for http).")
	eipCreateCmd.Flags().Int64P("healthcheck-strikes-fail", "", 0, "the number of times to retry before declaring the IP dead.")
	eipCreateCmd.Flags().Int64P("healthcheck-strikes-ok", "", 0, "the number of times to retry before declaring the IP healthy.")
	eipCreateCmd.Flags().Int64P("healthcheck-timeout", "", 0, "the timeout in seconds to wait for each check (default is 2). Should be lower than the interval.")
	eipCmd.AddCommand(eipCreateCmd)
}
