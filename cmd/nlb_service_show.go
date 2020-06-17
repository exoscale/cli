package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/table"
)

type nlbServerStatusShowOutput struct {
	InstanceIP string `json:"instance_ip"`
	Status     string `json:"status"`
}

type nlbServiceHealthcheckShowOutput struct {
	Mode     string        `json:"mode"`
	Port     uint16        `json:"port"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int64         `json:"retries"`
	URI      string        `json:"uri"`
}

type nlbServiceShowOutput struct {
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	InstancePoolID    string                          `json:"instance_pool_id"`
	Protocol          string                          `json:"protocol"`
	Port              uint16                          `json:"port"`
	TargetPort        uint16                          `json:"target_port"`
	Strategy          string                          `json:"strategy"`
	Healthcheck       nlbServiceHealthcheckShowOutput `json:"healthcheck"`
	HealthcheckStatus []nlbServerStatusShowOutput     `json:"healthcheck_status"`
	State             string                          `json:"state"`
}

func (o *nlbServiceShowOutput) toJSON() { outputJSON(o) }
func (o *nlbServiceShowOutput) toText() { outputText(o) }
func (o *nlbServiceShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"NLB Service"})
	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Instance Pool ID", o.InstancePoolID})
	t.Append([]string{"Protocol", o.Protocol})
	t.Append([]string{"Port", fmt.Sprint(o.Port)})
	t.Append([]string{"Target Port", fmt.Sprint(o.TargetPort)})
	t.Append([]string{"Strategy", o.Strategy})
	t.Append([]string{"Healthcheck Mode", o.Healthcheck.Mode})
	t.Append([]string{"Healthcheck Port", fmt.Sprint(o.Healthcheck.Port)})
	if o.Healthcheck.Mode == "http" {
		t.Append([]string{"Healthcheck URI", o.Healthcheck.URI})
	}
	t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.Healthcheck.Interval)})
	t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.Healthcheck.Timeout)})
	t.Append([]string{"Healthcheck Retries", fmt.Sprint(o.Healthcheck.Retries)})
	t.Append([]string{"Healthcheck Status", func() string {
		if len(o.HealthcheckStatus) > 0 {
			return strings.Join(
				func() []string {
					statuses := make([]string, len(o.HealthcheckStatus))
					for i := range o.HealthcheckStatus {
						statuses[i] = fmt.Sprintf("%s | %s",
							o.HealthcheckStatus[i].InstanceIP,
							o.HealthcheckStatus[i].Status)
					}
					return statuses
				}(),
				"\n")
		}
		return "n/a"
	}()})
	t.Append([]string{"State", o.State})
}

var nlbServiceShowCmd = &cobra.Command{
	Use:   "show <NLB ID> <ID>",
	Short: "Show a Network Load Balancer service details",
	Long: fmt.Sprintf(`This command shows a Network Load Balancer service details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbServiceShowOutput{}), ", ")),
	Aliases: gShowAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		return output(showNLBService(args[0], args[1], zone))
	},
}

func showNLBService(nlbID, svcID, zone string) (outputter, error) {
	var svc *egoscale.NetworkLoadBalancerService

	ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, ""))
	nlb, err := cs.GetNetworkLoadBalancer(ctx, zone, nlbID)
	if err != nil {
		return nil, err
	}
	for _, s := range nlb.Services {
		if s.ID == svcID {
			svc = s
			break
		}
	}
	if svc == nil {
		return nil, errors.New("service not found")
	}

	out := nlbServiceShowOutput{
		ID:             svc.ID,
		Name:           svc.Name,
		Description:    svc.Description,
		InstancePoolID: svc.InstancePoolID,
		Protocol:       svc.Protocol,
		Port:           svc.Port,
		TargetPort:     svc.TargetPort,
		Strategy:       svc.Strategy,
		State:          svc.State,

		Healthcheck: nlbServiceHealthcheckShowOutput{
			Mode:     svc.Healthcheck.Mode,
			Port:     svc.Healthcheck.Port,
			Interval: svc.Healthcheck.Interval,
			Timeout:  svc.Healthcheck.Timeout,
			Retries:  svc.Healthcheck.Retries,
			URI:      svc.Healthcheck.URI,
		},

		HealthcheckStatus: func() []nlbServerStatusShowOutput {
			statuses := make([]nlbServerStatusShowOutput, len(svc.HealthcheckStatus))
			for i, st := range svc.HealthcheckStatus {
				statuses[i] = nlbServerStatusShowOutput{
					InstanceIP: st.InstanceIP.String(),
					Status:     st.Status,
				}
			}
			return statuses
		}(),
	}

	return &out, nil
}

func init() {
	nlbServiceShowCmd.Flags().StringP("zone", "z", "", "Network Load Balancer zone")
	nlbServiceCmd.AddCommand(nlbServiceShowCmd)
}
