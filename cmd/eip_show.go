package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type eipHealthcheckShowOutput struct {
	Mode        string `json:"mode,omitempty"`
	Path        string `json:"path,omitempty"`
	Port        int64  `json:"port,omitempty"`
	Interval    int64  `json:"interval,omitempty"`
	Timeout     int64  `json:"timeout,omitempty"`
	StrikesOk   int64  `json:"strikes_ok,omitempty"`
	StrikesFail int64  `json:"strikes_fail,omitempty"`
}

type eipShowOutput struct {
	ID          string                    `json:"id"`
	Zone        string                    `json:"zone"`
	IPAddress   string                    `json:"ip_address"`
	Description string                    `json:"description"`
	Healthcheck *eipHealthcheckShowOutput `json:"healthcheck"`
	Instances   []string                  `json:"instances"`
}

func (o *eipShowOutput) toJSON() { outputJSON(o) }

func (o *eipShowOutput) toText() { outputText(o) }

func (o *eipShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Elastic IP"})

	t.Append([]string{"Description", o.Description})
	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"IP Address", o.IPAddress})

	if o.Healthcheck != nil {
		t.Append([]string{"Healthcheck Mode", o.Healthcheck.Mode})
		t.Append([]string{"Healthcheck Port", fmt.Sprint(o.Healthcheck.Port)})
		if o.Healthcheck.Mode == "http" {
			t.Append([]string{"Healthcheck Path", o.Healthcheck.Path})
		}
		t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.Healthcheck.Interval)})
		t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.Healthcheck.Timeout)})
		t.Append([]string{"Healthcheck Strikes OK", fmt.Sprint(o.Healthcheck.StrikesOk)})
		t.Append([]string{"Healthcheck Strikes Fail", fmt.Sprint(o.Healthcheck.StrikesFail)})
	}

	if len(o.Instances) > 0 {
		t.Append([]string{"Instances", strings.Join(o.Instances, "\n")})
	}

	t.Render()
}

func init() {
	eipCmd.AddCommand(&cobra.Command{
		Use:   "show <ip address | eip id>",
		Short: "Show an Elastic IP details",
		Long: fmt.Sprintf(`This command shows an Elastic IP details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&eipShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			return output(showEIP(args[0]))
		},
	})
}

func showEIP(eip string) (outputter, error) {
	id, err := egoscale.ParseUUID(eip)
	if err != nil {
		id, err = getEIPIDByIP(eip)
		if err != nil {
			return nil, err
		}
	}

	ip, vms, err := eipDetails(id)
	if err != nil {
		return nil, err
	}

	out := eipShowOutput{
		Description: ip.Description,
		ID:          id.String(),
		Zone:        ip.ZoneName,
		IPAddress:   ip.IPAddress.String(),
	}

	if ip.Healthcheck != nil {
		out.Healthcheck = &eipHealthcheckShowOutput{
			Mode:        ip.Healthcheck.Mode,
			Path:        ip.Healthcheck.Path,
			Port:        ip.Healthcheck.Port,
			Interval:    ip.Healthcheck.Interval,
			Timeout:     ip.Healthcheck.Timeout,
			StrikesOk:   ip.Healthcheck.StrikesOk,
			StrikesFail: ip.Healthcheck.StrikesFail,
		}
	}

	instances := make([]string, len(vms))
	for i := range vms {
		instances[i] = vms[i].Name
	}
	out.Instances = instances

	return &out, nil
}

func eipDetails(eip *egoscale.UUID) (*egoscale.IPAddress, []egoscale.VirtualMachine, error) {
	var eipID = eip

	query := &egoscale.IPAddress{ID: eipID, IsElastic: true}
	resp, err := cs.GetWithContext(gContext, query)
	if err != nil {
		return nil, nil, err
	}

	addr := resp.(*egoscale.IPAddress)
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{ZoneID: addr.ZoneID})
	if err != nil {
		return nil, nil, err
	}

	vmAssociated := []egoscale.VirtualMachine{}

	for _, value := range vms {
		vm := value.(*egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic == nil {
			continue
		}
		for _, sIP := range nic.SecondaryIP {
			if sIP.IPAddress.Equal(addr.IPAddress) {
				vmAssociated = append(vmAssociated, *vm)
			}
		}
	}

	return addr, vmAssociated, nil
}
