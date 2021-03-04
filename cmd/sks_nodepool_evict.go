package cmd

import (
	"errors"
	"fmt"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolEvictCmd = &cobra.Command{
	Use:   "evict <cluster name | ID> <Nodepool name | ID> <Node name | ID> [Node name | ID ...]",
	Short: "Evict SKS cluster Nodepool members",
	Long: `This command evicts specific members from a SKS cluster Nodepool, effectively
shrinking down the Nodepool similar to the "exo sks nodepool scale" command.

Note: Kubernetes Nodes should be drained from their workload prior to being
evicted from their Nodepool, e.g. using "kubectl drain".`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c  = args[0]
			np = args[1]

			nodepool *exov2.SKSNodepool
		)

		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return fmt.Errorf("error retrieving zone: %s", err)
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Do you really want to evict %v from Nodepool %q?", args[2:], np)) {
				return nil
			}
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
		cluster, err := lookupSKSCluster(ctx, zone.Name, c)
		if err != nil {
			return err
		}

		for _, n := range cluster.Nodepools {
			if n.ID == np || n.Name == np {
				nodepool = n
				break
			}
		}
		if nodepool == nil {
			return errors.New("Nodepool not found") // nolint:golint
		}

		nodes := make([]string, len(args[2:]))
		for i, n := range args[2:] {
			instance, err := getVirtualMachineByNameOrID(n)
			if err != nil {
				return fmt.Errorf("invalid Node %q: %s", n, err)
			}
			nodes[i] = instance.ID.String()
		}

		decorateAsyncOperation(fmt.Sprintf("Evicting Nodes from Nodepool %q...", np), func() {
			err = cluster.EvictNodepoolMembers(ctx, nodepool, nodes)
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			return output(showSKSNodepool(zone, cluster.ID, nodepool.ID))
		}

		return nil
	},
}

func init() {
	sksNodepoolEvictCmd.Flags().BoolP("force", "f", false, "Attempt to evict without prompting for confirmation")
	sksNodepoolEvictCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolCmd.AddCommand(sksNodepoolEvictCmd)
}
