package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

const (
	limitComputeInstances    = "instance"
	limitDatabases           = "database"
	limitElasticIPs          = "elastic-ip"
	limitIAMAPIKeys          = "iam-key"
	limitInstanceGPUs        = "gpu"
	limitInstanceSnapshots   = "snapshot"
	limitInstanceTemplates   = "template"
	limitNLB                 = "network-load-balancer"
	limitPrivateNetworks     = "private-network"
	limitSKSClusters         = "sks-cluster"
	limitSOSBuckets          = "bucket"
	limitBlockStorageVolumes = "block-storage-volume"
	limitBlockStorage        = "block-storage"
	limitBlockStorageMaxSize = "block-storage-max-size"

	gpu2          = "gpu2"
	gpu3          = "gpu3"
	gpua30        = "gpua30"
	gpu3080ti     = "gpu3080ti"
	gpua5000      = "gpua5000"
	gpurtx6000pro = "gpurtx6000pro"
)

type LimitsItemOutput struct {
	Resource string `json:"resource"`
	Used     int64  `json:"used"`
	Max      int64  `json:"max"`
}

type LimitsOutput []LimitsItemOutput

func (o *LimitsOutput) ToJSON()  { output.JSON(o) }
func (o *LimitsOutput) ToText()  { output.Text(o) }
func (o *LimitsOutput) ToTable() { output.Table(o) }

var showDetailedGPU bool

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Current account limits",
	Long: fmt.Sprintf(`This command lists the safety limits currently enforced on your account.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&LimitsOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := GContext
		client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
		if err != nil {
			return err
		}

		quotas, err := client.ListQuotas(ctx)
		if err != nil {
			return err
		}

		gpuLabels := map[string]string{
			limitInstanceGPUs:   "GPU - Compute instance GPUs",
			gpu2:                "GPU - GPU2",
			gpu3:                "GPU - GPU3",
			gpua30:              "GPU - A30",
			gpu3080ti:           "GPU - 3080 Ti",
			gpua5000:            "GPU - A5000",
			gpurtx6000pro:       "GPU - RTX 6000 Pro",
		}

		resourceLimitLabels := map[string]string{
			limitComputeInstances:    "Compute instances",
			limitDatabases:           "Databases",
			limitElasticIPs:          "Elastic IP addresses",
			limitIAMAPIKeys:          "IAM API keys",
			limitInstanceGPUs:        "Compute instance GPUs",
			limitInstanceSnapshots:   "Compute instance snapshots",
			limitInstanceTemplates:   "Compute instance templates",
			limitNLB:                 "Network Load Balancers",
			limitPrivateNetworks:     "Private networks",
			limitSKSClusters:         "SKS clusters",
			limitSOSBuckets:          "SOS buckets",
			limitBlockStorageVolumes: "Block Storage Volumes",
			limitBlockStorage:        "Block Storage cumulative size (GiB)",
			limitBlockStorageMaxSize: "Max size of a Block Storage Volume (GiB)",
		}

		out := LimitsOutput{}
		for _, quota := range quotas.Quotas {
			if label, ok := resourceLimitLabels[quota.Resource]; ok {
				if showDetailedGPU && quota.Resource == limitInstanceGPUs {
					continue
				}
				out = append(out, LimitsItemOutput{
					Resource: label,
					Used:     quota.Usage,
					Max:      quota.Limit,
				})
			} else if showDetailedGPU {
				if label, ok := gpuLabels[quota.Resource]; ok {
					out = append(out, LimitsItemOutput{
						Resource: label,
						Used:     quota.Usage,
						Max:      quota.Limit,
					})
				}
			}
		}

		sort.Slice(out, func(i, j int) bool {
			return out[i].Resource < out[j].Resource
		})

		return utils.PrintOutput(&out, nil)
	},
}

func init() {
	RootCmd.AddCommand(limitsCmd)
	limitsCmd.Flags().BoolVar(&showDetailedGPU, "gpu", false, "Also show per-family GPU limits")
}