package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var nlbServiceUpdateCmd = &cobra.Command{
	Use:   "update NLB-NAME|ID SERVICE-NAME|ID",
	Short: "Update a Network Load Balancer service",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "missing arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			svc    *exov2.NetworkLoadBalancerService
			nlbRef = args[0]
			svcRef = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := lookupNLB(ctx, zone, nlbRef)
		if err != nil {
			return err
		}

		for _, s := range nlb.Services {
			if s.ID == svcRef || s.Name == svcRef {
				svc = s
				break
			}
		}
		if svc == nil {
			return errors.New("service not found")
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") {
			svc.Name = name
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("description") {
			svc.Description = description
		}

		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("protocol") {
			svc.Protocol = protocol
		}

		port, err := cmd.Flags().GetUint16("port")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("port") {
			svc.Port = port
		}

		targetPort, err := cmd.Flags().GetUint16("target-port")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("target-port") {
			svc.TargetPort = targetPort
		}

		strategy, err := cmd.Flags().GetString("strategy")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("strategy") {
			svc.Strategy = strategy
		}

		healthcheckMode, err := cmd.Flags().GetString("healthcheck-mode")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-mode") {
			svc.Healthcheck.Mode = healthcheckMode
		}

		healthcheckURI, err := cmd.Flags().GetString("healthcheck-uri")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-uri") {
			svc.Healthcheck.URI = healthcheckURI
		}
		if strings.HasPrefix(healthcheckMode, "http") && healthcheckURI == "" {
			return errors.New(`an healthcheck URI is required in "http(s)" mode`)
		}

		healthcheckPort, err := cmd.Flags().GetUint16("healthcheck-port")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-port") {
			svc.Healthcheck.Port = healthcheckPort
		}

		healthcheckInterval, err := cmd.Flags().GetInt64("healthcheck-interval")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-interval") {
			svc.Healthcheck.Interval = time.Duration(healthcheckInterval) * time.Second
		}

		healthcheckTimeout, err := cmd.Flags().GetInt64("healthcheck-timeout")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-timeout") {
			svc.Healthcheck.Timeout = time.Duration(healthcheckTimeout) * time.Second
		}

		healthcheckRetries, err := cmd.Flags().GetInt64("healthcheck-retries")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-retries") {
			svc.Healthcheck.Retries = healthcheckRetries
		}

		healthcheckTLSSNI, err := cmd.Flags().GetString("healthcheck-tls-sni")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("healthcheck-tls-sni") {
			svc.Healthcheck.TLSSNI = healthcheckTLSSNI
		}
		if healthcheckTLSSNI != "" && healthcheckMode != "https" {
			return errors.New(`a healthcheck TLS SNI can only be specified in https mode`)
		}

		if err := nlb.UpdateService(ctx, svc); err != nil {
			return fmt.Errorf("unable to update service: %s", err)
		}

		if !gQuiet {
			return output(showNLBService(zone, nlb.ID, svc.ID))
		}

		return nil
	},
}

func init() {
	nlbServiceUpdateCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbServiceUpdateCmd.Flags().String("name", "", "service name")
	nlbServiceUpdateCmd.Flags().String("description", "", "service description")
	nlbServiceUpdateCmd.Flags().String("protocol", "", "protocol of the service (tcp|udp)")
	nlbServiceUpdateCmd.Flags().Uint16("port", 0, "service port")
	nlbServiceUpdateCmd.Flags().Uint16("target-port", 0, "port to forward traffic to on target instances")
	nlbServiceUpdateCmd.Flags().String("strategy", "", "load balancing strategy (round-robin|source-hash)")
	nlbServiceUpdateCmd.Flags().String("healthcheck-mode", "", "service health checking mode (tcp|http|https)")
	nlbServiceUpdateCmd.Flags().String("healthcheck-uri", "", "service health checking URI (required in http(s) mode)")
	nlbServiceUpdateCmd.Flags().Uint16("healthcheck-port", 0, "service health checking port")
	nlbServiceUpdateCmd.Flags().Int64("healthcheck-interval", 0, "service health checking interval in seconds")
	nlbServiceUpdateCmd.Flags().Int64("healthcheck-timeout", 0, "service health checking timeout in seconds")
	nlbServiceUpdateCmd.Flags().Int64("healthcheck-retries", 0, "service health checking retries")
	nlbServiceUpdateCmd.Flags().String("healthcheck-tls-sni", "", "service health checking server name to present with SNI in https mode")
	nlbServiceCmd.AddCommand(nlbServiceUpdateCmd)
}
