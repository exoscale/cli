package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const (
	serviceOfferingHelp = "service offering NAME (micro|tiny|small|medium|large|extra-large|huge|mega|titan|jumbo)"
)

type serviceOfferingListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	CPU  string `json:"cpu"`
	RAM  string `json:"ram"`
}

type serviceOfferingListOutput []serviceOfferingListItemOutput

func (o *serviceOfferingListOutput) toJSON()  { output.JSON(o) }
func (o *serviceOfferingListOutput) toText()  { output.Text(o) }
func (o *serviceOfferingListOutput) toTable() { output.Table(o) }

func init() {
	vmCmd.AddCommand(&cobra.Command{
		Use:   "serviceoffering",
		Short: "List available services offerings with details",
		Long: fmt.Sprintf(`This command lists available Compute service offerings.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&serviceOfferingListOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listServiceOfferings())
		},
	})
}

func listServiceOfferings() (output.Outputter, error) {
	serviceOffering, err := globalstate.GlobalEgoscaleClient.ListWithContext(gContext, &egoscale.ServiceOffering{})
	if err != nil {
		return nil, err
	}

	out := serviceOfferingListOutput{}

	for _, key := range serviceOffering {
		so := key.(*egoscale.ServiceOffering)

		ram := ""
		if so.Memory > 1000 {
			ram = fmt.Sprintf("%d GB", so.Memory>>10)
		} else if so.Memory < 1000 {
			ram = fmt.Sprintf("%d MB", so.Memory)
		}

		out = append(out, serviceOfferingListItemOutput{
			ID:   so.ID.String(),
			Name: so.Name,
			CPU:  fmt.Sprintf("%d × %d MHz", so.CPUNumber, so.CPUSpeed),
			RAM:  ram,
		})
	}

	return &out, nil
}

func getServiceOfferingByNameOrID(v string) (*egoscale.ServiceOffering, error) {
	so := &egoscale.ServiceOffering{}

	id, err := egoscale.ParseUUID(v)
	if err != nil {
		so.Name = v
	} else {
		so.ID = id
	}

	resp, err := globalstate.GlobalEgoscaleClient.GetWithContext(gContext, so)
	switch err {
	case nil:
		return resp.(*egoscale.ServiceOffering), nil

	case egoscale.ErrNotFound:
		return nil, fmt.Errorf("unknown Service Offering %q", v)

	case egoscale.ErrTooManyFound:
		return nil, fmt.Errorf("multiple Service Offerings match %q", v)

	default:
		return nil, err
	}
}
