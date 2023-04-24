package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type eipHealthcheckShowOutput struct {
	Mode          string `json:"mode,omitempty"`
	Path          string `json:"path,omitempty"`
	Port          int64  `json:"port,omitempty"`
	Interval      int64  `json:"interval,omitempty"`
	Timeout       int64  `json:"timeout,omitempty"`
	StrikesOk     int64  `json:"strikes_ok,omitempty"`
	StrikesFail   int64  `json:"strikes_fail,omitempty"`
	TLSSkipVerify bool   `json:"tls_skip_verify"`
	TLSSNI        string `json:"tls_sni,omitempty"`
}

type eipShowOutput struct {
	ID          string                    `json:"id"`
	Zone        string                    `json:"zone"`
	IPAddress   string                    `json:"ip_address"`
	Description string                    `json:"description"`
	Healthcheck *eipHealthcheckShowOutput `json:"healthcheck"`
	Instances   []string                  `json:"instances"`
}

func (o *eipShowOutput) toJSON() { output.JSON(o) }

func (o *eipShowOutput) toText() { output.Text(o) }

func (o *eipShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Elastic IP"})

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"IP Address", o.IPAddress})
	t.Append([]string{"Description", o.Description})

	if o.Healthcheck != nil {
		t.Append([]string{"Healthcheck Mode", o.Healthcheck.Mode})
		t.Append([]string{"Healthcheck Port", fmt.Sprint(o.Healthcheck.Port)})
		if strings.HasPrefix(o.Healthcheck.Mode, "http") {
			t.Append([]string{"Healthcheck Path", o.Healthcheck.Path})
		}
		t.Append([]string{"Healthcheck Interval", fmt.Sprint(o.Healthcheck.Interval)})
		t.Append([]string{"Healthcheck Timeout", fmt.Sprint(o.Healthcheck.Timeout)})
		t.Append([]string{"Healthcheck Strikes OK", fmt.Sprint(o.Healthcheck.StrikesOk)})
		t.Append([]string{"Healthcheck Strikes Fail", fmt.Sprint(o.Healthcheck.StrikesFail)})
		if o.Healthcheck.Mode == "https" {
			t.Append([]string{"Healthcheck TLS Skip Verification", fmt.Sprintf("%t", o.Healthcheck.TLSSkipVerify)})
			t.Append([]string{"Healthcheck TLS SNI", fmt.Sprint(o.Healthcheck.TLSSNI)})
		}
	}

	if len(o.Instances) > 0 {
		t.Append([]string{"Instances", strings.Join(o.Instances, "\n")})
	}

	t.Render()
}

func init() {
	eipCmd.AddCommand(&cobra.Command{
		Use:   "show IP-ADDRESS|ID",
		Short: "Show an Elastic IP details",
		Long: fmt.Sprintf(`This command shows an Elastic IP details.

Supported output template annotations: %s`,
			strings.Join(output.output.OutputterTemplateAnnotations(&eipShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			return printOutput(showEIP(args[0]))
		},
	})
}

func showEIP(v string) (output.Outputter, error) {
	eip, err := getElasticIPByAddressOrID(v)
	if err != nil {
		return nil, err
	}

	out := eipShowOutput{
		ID:          eip.ID.String(),
		Zone:        eip.ZoneName,
		Description: eip.Description,
		IPAddress:   eip.IPAddress.String(),
	}

	if eip.Healthcheck != nil {
		out.Healthcheck = &eipHealthcheckShowOutput{
			Mode:          eip.Healthcheck.Mode,
			Path:          eip.Healthcheck.Path,
			Port:          eip.Healthcheck.Port,
			Interval:      eip.Healthcheck.Interval,
			Timeout:       eip.Healthcheck.Timeout,
			StrikesOk:     eip.Healthcheck.StrikesOk,
			StrikesFail:   eip.Healthcheck.StrikesFail,
			TLSSkipVerify: eip.Healthcheck.TLSSkipVerify,
			TLSSNI:        eip.Healthcheck.TLSSNI,
		}
	}

	res, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{ZoneID: eip.ZoneID})
	if err != nil {
		return nil, err
	}

	for _, item := range res {
		vm := item.(*egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic == nil {
			continue
		}

		for _, sIP := range nic.SecondaryIP {
			if sIP.IPAddress.Equal(eip.IPAddress) {
				out.Instances = append(out.Instances, vm.Name)
			}
		}
	}

	return &out, nil
}
