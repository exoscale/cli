package keystore

import (
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type KeyStoreShowOutput struct {
	ID              v3.UUID   `json:"id"`
	Name            string    `json:"name"`
	KeyStoreType    string    `json:"type" outputLabel:"Type"`
	Status          string    `json:"status"`
	StatusSince     time.Time `json:"status-since"`
	Endpoint        string    `json:"endpoint"`
	AccessKey       string    `json:"access-key"`
	Description     string    `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created-at"`
	Health          string    `json:"health,omitempty"`
	HealthReason    string    `json:"health-reason,omitempty"`
	HealthError     string    `json:"health-error,omitempty"`
	HealthCheckedAt time.Time `json:"health-checked-at,omitempty"`
}

func (o *KeyStoreShowOutput) Type() string { return "Key store" }
func (o *KeyStoreShowOutput) ToJSON()      { output.JSON(o) }
func (o *KeyStoreShowOutput) ToText()      { output.Text(o) }
func (o *KeyStoreShowOutput) ToTable()     { output.Table(o) }

func newKeyStoreShowOutput(ks *v3.GetKeyStoreResponse) *KeyStoreShowOutput {
	out := &KeyStoreShowOutput{
		ID:           ks.ID,
		Name:         ks.Name,
		KeyStoreType: string(ks.Type),
		Status:       string(ks.Status),
		StatusSince:  ks.StatusSince,
		Description:  ks.Description,
		CreatedAt:    ks.CreatedAT,
	}
	if ks.Proxy != nil {
		out.Endpoint = ks.Proxy.Endpoint
		if ks.Proxy.Auth != nil {
			out.AccessKey = ks.Proxy.Auth.Key
		}
	}
	if ks.Health != nil {
		out.Health = string(ks.Health.Status)
		out.HealthReason = ks.Health.StatusReason
		out.HealthError = ks.Health.ErrorDetail
		out.HealthCheckedAt = ks.Health.CheckedAT
	}
	return out
}

type KeyStoreShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	KeyStore string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *KeyStoreShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *KeyStoreShowCmd) CmdShort() string     { return "Retrieve key store details" }
func (c *KeyStoreShowCmd) CmdLong() string {
	return "This command retrieves an external key store's details, including the latest health check of its XKS proxy."
}
func (c *KeyStoreShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *KeyStoreShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	id, err := parseKeyStoreID(c.KeyStore)
	if err != nil {
		return err
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	ks, err := client.GetKeyStore(ctx, id)
	if err != nil {
		return err
	}

	return c.OutputFunc(newKeyStoreShowOutput(ks), nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &KeyStoreShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
