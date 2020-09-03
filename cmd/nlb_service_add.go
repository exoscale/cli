package cmd

import (
	"fmt"
	"time"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var nlbServiceAddCmd = &cobra.Command{
	Use:   "add <NLB name | ID> <service name>",
	Short: "Add a service to a Network Load Balancer",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{
			"healthcheck-interval",
			"healthcheck-retries",
			"healthcheck-timeout",
			"instance-pool-id",
			"port",
			"protocol",
			"strategy",
			"zone",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			nlbRef = args[0]
			name   = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		instancePoolID, err := cmd.Flags().GetString("instance-pool-id")
		if err != nil {
			return err
		}

		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			return err
		}

		port, err := cmd.Flags().GetUint16("port")
		if err != nil {
			return err
		}

		targetPort, err := cmd.Flags().GetUint16("target-port")
		if err != nil {
			return err
		}
		if targetPort == 0 {
			targetPort = port
		}

		strategy, err := cmd.Flags().GetString("strategy")
		if err != nil {
			return err
		}

		healthcheckMode, err := cmd.Flags().GetString("healthcheck-mode")
		if err != nil {
			return err
		}

		healthcheckURI, err := cmd.Flags().GetString("healthcheck-uri")
		if err != nil {
			return err
		}

		healthcheckPort, err := cmd.Flags().GetUint16("healthcheck-port")
		if err != nil {
			return err
		}
		if healthcheckPort == 0 {
			healthcheckPort = targetPort
		}

		healthcheckInterval, err := cmd.Flags().GetInt64("healthcheck-interval")
		if err != nil {
			return err
		}

		healthcheckTimeout, err := cmd.Flags().GetInt64("healthcheck-timeout")
		if err != nil {
			return err
		}

		healthcheckRetries, err := cmd.Flags().GetInt64("healthcheck-retries")
		if err != nil {
			return err
		}

		healthcheckTLSSNI, err := cmd.Flags().GetString("healthcheck-tls-sni")
		if err != nil {
			return err
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		nlb, err := lookupNLB(ctx, zone, nlbRef)
		if err != nil {
			return err
		}

		svc, err := nlb.AddService(ctx, &egoscale.NetworkLoadBalancerService{
			Name:           name,
			Description:    description,
			InstancePoolID: instancePoolID,
			Protocol:       protocol,
			Port:           port,
			TargetPort:     targetPort,
			Strategy:       strategy,
			Healthcheck: egoscale.NetworkLoadBalancerServiceHealthcheck{
				Mode:     healthcheckMode,
				Port:     healthcheckPort,
				URI:      healthcheckURI,
				Interval: time.Duration(healthcheckInterval) * time.Second,
				Timeout:  time.Duration(healthcheckTimeout) * time.Second,
				Retries:  healthcheckRetries,
				TLSSNI:   healthcheckTLSSNI,
			},
		})
		if err != nil {
			return fmt.Errorf("unable to add service: %s", err)
		}

		if !gQuiet {
			return output(showNLBService(zone, nlb.ID, svc.ID))
		}

		return nil
	},
}

func init() {
	nlbServiceAddCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbServiceAddCmd.Flags().String("instance-pool-id", "",
		"ID of the Instance Pool to forward traffic to")
	nlbServiceAddCmd.Flags().String("description", "", "service description")
	nlbServiceAddCmd.Flags().String("protocol", "tcp", "protocol of the service (tcp|udp)")
	nlbServiceAddCmd.Flags().Uint16("port", 0, "service port")
	nlbServiceAddCmd.Flags().Uint16("target-port", 0, "port to forward traffic to on target instances (defaults to service port)")
	nlbServiceAddCmd.Flags().String("strategy", "round-robin", "load balancing strategy (round-robin|source-hash)")
	nlbServiceAddCmd.Flags().String("healthcheck-mode", "tcp", "service health checking mode (tcp|http|https)")
	nlbServiceAddCmd.Flags().String("healthcheck-uri", "", "service health checking URI (required in http(s) mode)")
	nlbServiceAddCmd.Flags().Uint16("healthcheck-port", 0, "service health checking port (defaults to target port)")
	nlbServiceAddCmd.Flags().Int64("healthcheck-interval", 10, "service health checking interval in seconds")
	nlbServiceAddCmd.Flags().Int64("healthcheck-timeout", 5, "service health checking timeout in seconds")
	nlbServiceAddCmd.Flags().Int64("healthcheck-retries", 1, "service health checking retries")
	nlbServiceAddCmd.Flags().String("healthcheck-tls-sni", "", "service health checking server name to present with SNI in https mode")
	nlbServiceCmd.AddCommand(nlbServiceAddCmd)
}
