package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type rsyslogOutput struct {
	ID       string                                `json:"id"`
	Name     string                                `json:"name"`
	Type     string                                `json:"type"`
	Settings v3.DBAASEndpointRsyslogOptionalFields `json:"settings"`
}

func (o *rsyslogOutput) ToJSON() { output.JSON(o) }
func (o *rsyslogOutput) ToText() { output.Text(o) }
func (o *rsyslogOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Rsyslog External Endpoint"})
	defer t.Render()

	t.Append([]string{"Endpoint ID", o.ID})
	t.Append([]string{"Endpoint Name", o.Name})
	t.Append([]string{"Endpoint Type", o.Type})

	settings := o.Settings
	tls := "false"

	if settings.Tls != nil && *settings.Tls {
		tls = "true"
	}

	t.Append([]string{"Server", settings.Server})
	t.Append([]string{"Port", strconv.FormatInt(settings.Port, 10)})
	t.Append([]string{"Tls", tls})
	t.Append([]string{"Max Message Size", strconv.FormatInt(settings.MaxMessageSize, 10)})
	t.Append([]string{"Structured data block", settings.SD})
	t.Append([]string{"Custom logline format", settings.Logline})
}

func (c *dbaasExternalEndpointShowCmd) showRsyslog() (output.Outputter, error) {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	endpointUUID, err := v3.ParseUUID(c.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint ID: %w", err)
	}

	endpointResponse, err := client.GetDBAASExternalEndpointRsyslog(ctx, endpointUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting Rsyslog external endpoint: %w", err)
	}

	output := &rsyslogOutput{
		ID:       endpointResponse.ID.String(),
		Name:     endpointResponse.Name,
		Type:     string(endpointResponse.Type),
		Settings: *endpointResponse.Settings,
	}

	return output, nil
}
