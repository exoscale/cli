package cmd

import (
	"fmt"
	"strings"

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

func (o *serviceOfferingListOutput) toJSON()  { outputJSON(o) }
func (o *serviceOfferingListOutput) toText()  { outputText(o) }
func (o *serviceOfferingListOutput) toTable() { outputTable(o) }

func init() {
	vmCmd.AddCommand(&cobra.Command{
		Use:   "serviceoffering",
		Short: "List available services offerings with details",
		Long: fmt.Sprintf(`This command lists available Compute service offerings.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&serviceOfferingListOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listServiceOfferings())
		},
	})
}

func listServiceOfferings() (outputter, error) {
	serviceOffering, err := cs.ListWithContext(gContext, &egoscale.ServiceOffering{})
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

	resp, err := cs.GetWithContext(gContext, so)
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
