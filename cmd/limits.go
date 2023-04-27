package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

const (
	limitComputeInstances  = "instance"
	limitDatabases         = "database"
	limitElasticIPs        = "elastic-ip"
	limitIAMAPIKeys        = "iam-key"
	limitInstanceGPUs      = "gpu"
	limitInstanceSnapshots = "snapshot"
	limitInstanceTemplates = "template"
	limitNLB               = "network-load-balancer"
	limitPrivateNetworks   = "private-network"
	limitSKSClusters       = "sks-cluster"
	limitSOSBuckets        = "bucket"
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

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Current account limits",
	Long: fmt.Sprintf(`This command lists the safety limits currently enforced on your account.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&LimitsOutput{}), ", ")),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceLimitLabels := map[string]string{
			limitComputeInstances:  "Compute instances",
			limitDatabases:         "Databases",
			limitElasticIPs:        "Elastic IP addresses",
			limitIAMAPIKeys:        "IAM API keys",
			limitInstanceGPUs:      "Compute instance GPUs",
			limitInstanceSnapshots: "Compute instance snapshots",
			limitInstanceTemplates: "Compute instance templates",
			limitNLB:               "Network Load Balancers",
			limitPrivateNetworks:   "Private networks",
			limitSKSClusters:       "SKS clusters",
			limitSOSBuckets:        "SOS buckets",
		}

		out := LimitsOutput{}

		quotas, err := globalstate.EgoscaleClient.ListQuotas(gContext, account.CurrentAccount.DefaultZone)
		if err != nil {
			return err
		}

		for _, quota := range quotas {
			if _, ok := resourceLimitLabels[*quota.Resource]; !ok {
				continue
			}

			out = append(out, LimitsItemOutput{
				Resource: resourceLimitLabels[*quota.Resource],
				Used:     *quota.Usage,
				Max:      *quota.Limit,
			})
		}

		return printOutput(&out, nil)
	},
}

func init() {
	RootCmd.AddCommand(limitsCmd)
}
