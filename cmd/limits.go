package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/exoscale/egoscale"
	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

const (
	limitComputeInstances = "user_vm"
	limitGPUs             = "gpu"
	limitSnapshots        = "snapshot"
	limitTemplates        = "template"
	limitIPAddresses      = "public_elastic_ip"
	limitPrivateNetworks  = "network"
	limitNLBs             = "network_load_balancer"
	limitIAMAPIKeys       = "iam_key"
	limitSOSBuckets       = "bucket"
	limitSKSClusters      = "sks_cluster"
)

type LimitsItemOutput struct {
	Resource string `json:"resource"`
	Used     int    `json:"used"`
	Max      int    `json:"max"`
}

type LimitsOutput []LimitsItemOutput

func (o *LimitsOutput) toJSON()  { outputJSON(o) }
func (o *LimitsOutput) toText()  { outputText(o) }
func (o *LimitsOutput) toTable() { outputTable(o) }

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Current account limits",
	Long: fmt.Sprintf(`This command lists the safety limits currently enforced on your account.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&LimitsOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		var curUsage sync.Map

		// Global resources ///////////////////////////////////////////////

		res, err := cs.RequestWithContext(gContext, &egoscale.ListAPIKeys{})
		if err != nil {
			return fmt.Errorf("unable to list IAM API keys: %s", err)
		}
		curUsage.Store(limitIAMAPIKeys, res.(*egoscale.ListAPIKeysResponse).Count)

		res, err = cs.RequestWithContext(gContext, egoscale.ListBucketsUsage{})
		if err != nil {
			return fmt.Errorf("unable to list SOS buckets: %s", err)
		}
		curUsage.Store(limitSOSBuckets, res.(*egoscale.ListBucketsUsageResponse).Count)

		// Zone-local resources /////////////////////////////////////////////

		instanceTypes := make(map[string]*exov2.InstanceType) // For caching

		err = forEachZone(allZones, func(zone string) error {
			ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

			instances, err := cs.ListInstances(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list Compute instances: %s", err)
			}
			curUsage.Store(limitGPUs, 0)
			cur, _ := curUsage.LoadOrStore(limitComputeInstances, 0)
			curUsage.Store(limitComputeInstances, cur.(int)+len(instances))

			for _, instance := range instances {
				instanceType, cached := instanceTypes[*instance.InstanceTypeID]
				if !cached {
					instanceType, err = cs.GetInstanceType(ctx, zone, *instance.InstanceTypeID)
					if err != nil {
						return fmt.Errorf(
							"unable to retrieve Compute instance type %q: %s",
							*instance.InstanceTypeID,
							err)
					}
					instanceTypes[*instance.InstanceTypeID] = instanceType
				}

				if strings.HasSuffix(*instanceType.Family, "gpu") {
					cur, _ = curUsage.Load(limitGPUs)
					curUsage.Store(limitGPUs, cur.(int)+int(*instanceType.GPUs))
				}
			}

			snapshots, err := cs.ListSnapshots(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list snapshots: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitSnapshots, 0)
			curUsage.Store(limitSnapshots, cur.(int)+len(snapshots))

			templates, err := cs.ListTemplates(ctx, zone, exov2.ListTemplatesWithVisibility("private"))
			if err != nil {
				return fmt.Errorf("unable to list templates: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitTemplates, 0)
			curUsage.Store(limitTemplates, cur.(int)+len(templates))

			elasticIPs, err := cs.ListElasticIPs(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list IP addresses: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitIPAddresses, 0)
			curUsage.Store(limitIPAddresses, cur.(int)+len(elasticIPs))

			privateNetworks, err := cs.ListPrivateNetworks(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list Private Networks: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitPrivateNetworks, 0)
			curUsage.Store(limitPrivateNetworks, cur.(int)+len(privateNetworks))

			nlbs, err := cs.ListNetworkLoadBalancers(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list Network Load Balancers: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitNLBs, 0)
			curUsage.Store(limitNLBs, cur.(int)+len(nlbs))

			sksClusters, err := cs.ListSKSClusters(ctx, zone)
			if err != nil {
				return fmt.Errorf("unable to list SKS clusters: %s", err)
			}
			cur, _ = curUsage.LoadOrStore(limitSKSClusters, 0)
			curUsage.Store(limitSKSClusters, cur.(int)+len(sksClusters))

			return nil
		})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
		}

		resourceLimitLabels := map[string]string{
			limitComputeInstances: "Instances",
			limitGPUs:             "GPUs",
			limitSnapshots:        "Snapshots",
			limitTemplates:        "Templates",
			limitIPAddresses:      "IP addresses",
			limitPrivateNetworks:  "Private Networks",
			limitNLBs:             "Network Load Balancers",
			limitIAMAPIKeys:       "IAM API keys",
			limitSOSBuckets:       "SOS buckets",
			limitSKSClusters:      "SKS clusters",
		}

		out := LimitsOutput{}

		limits, err := cs.ListWithContext(gContext, &egoscale.ResourceLimit{})
		if err != nil {
			return err
		}

		for _, key := range limits {
			limit := key.(*egoscale.ResourceLimit)

			cur, ok := curUsage.Load(limit.ResourceTypeName)
			if ok {
				out = append(out, LimitsItemOutput{
					Resource: resourceLimitLabels[limit.ResourceTypeName],
					Used:     cur.(int),
					Max:      int(limit.Max),
				})
			}
		}

		return output(&out, nil)
	},
}

func init() {
	RootCmd.AddCommand(limitsCmd)
}
