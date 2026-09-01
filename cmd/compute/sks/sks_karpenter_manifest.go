package sks

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

// type sksKubeconfigCmd struct {
type sksKarpenterManifestCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"karpenter-manifests"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`

	ExoscaleNodeclass bool        `cli-short:"e" cli-usage:"generate ExoscaleNodeclass resource with important fields pre-filled"`
	Nodepool          bool        `cli-short:"n" cli-usage:"generate a Nodepool manifest to define the scaling policies and constraints for the nodes created by Karpenter"`
	Zone              v3.ZoneName `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksKarpenterManifestCmd) CmdAliases() []string { return []string{"km"} }

func (c *sksKarpenterManifestCmd) CmdShort() string {
	return "Generate Karpenter-related manifests for an SKS cluster"
}

func (c *sksKarpenterManifestCmd) CmdLong() string {
	return `This command generates the Karpenter manifests for an SKS cluster

It create the code for the Kubernetes resources "ExoscaleNodeClass" and/or
"NodePool" relevant to the given cluster.

Example usage:

    # Generate ExoscaleNodeClass and NodePool resources
    $ exo compute sks karpenter-manifests my-cluster --zone de-fra-1
	---
	apiVersion: karpenter.exoscale.com/v1
	kind: ExoscaleNodeClass
	metadata:
	[...]
	---
	apiVersion: karpenter.sh/v1
	kind: NodePool
	metadata:
  	  name: standard
	[...]
	
	$ exo compute sks km my-cluster -n
	---
	apiVersion: karpenter.sh/v1
	kind: NodePool
	metadata:
  	  name: standard
	[...]

You can get only the NodePool or the ExoscaleNodeClass resource by using --nodepool (-n)
or --exoscale-nodeclass (-e) respectively.

When no flags are given, both is assumed. It's equivalent to "-en".

The output can be piped to a "kubectl apply -f -" command or written to a file for further
customization.

Notes:

* Refer to Karpenter documentation for more information on the resources NodeClasses [1]
and NodePools [2]

[1]: https://karpenter.sh/docs/concepts/nodeclasses/
[2]: https://karpenter.sh/docs/concepts/nodepools/
`
}

func (c *sksKarpenterManifestCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksKarpenterManifestCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := resp.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	// if no options are given, defaults to both
	if !c.Nodepool && !c.ExoscaleNodeclass {
		c.Nodepool = true
		c.ExoscaleNodeclass = true
	}

	re := regexp.MustCompile(`(\n| )+$`)
	if c.ExoscaleNodeclass {
		respExoscaleNodeClass, err := client.GenerateSKSKarpenterExoscaleNodeclass(ctx, cluster.ID)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", re.ReplaceAllString(respExoscaleNodeClass.ExoscaleNodeclass, ""))
	}

	if c.Nodepool {
		respNodePool, err := client.GenerateSKSKarpenterNodepool(ctx, cluster.ID)
		if err != nil {
			return err
		}

		fmt.Printf("---\n%s\n", re.ReplaceAllString(respNodePool.Nodepool, ""))
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksKarpenterManifestCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
