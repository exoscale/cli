package cmd

import (
	"fmt"
	"time"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasKafkaUserShowOutput struct {
	AccessKey        string    `json:"access-key,omitempty"`
	AccessCert       string    `json:"access-cert,omitempty"`
	AccessCertExpiry time.Time `json:"access-cert-expiry,omitempty"`
}

func (o *dbaasKafkaUserShowOutput) formatUser(t *table.Table) {
	t.Append([]string{"Access Key", o.AccessKey})
	t.Append([]string{"Access Cert", o.AccessCert})
	t.Append([]string{"Access Cert Expiry", o.AccessCertExpiry.String()})
}

func (c *dbaasUserShowCmd) showKafka(cmd *cobra.Command, _ []string) (output.Outputter, error) {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	s, err := client.GetDBAASServiceKafka(ctx, c.Name)
	if err != nil {
		return &dbaasUserShowOutput{}, err
	}

	for _, u := range s.Users {
		if u.Username == c.Username {
			return &dbaasUserShowOutput{
				Username: c.Username,
				Password: u.Password,
				Type:     u.Type,
				Kafka: &dbaasKafkaUserShowOutput{
					AccessKey:        u.AccessKey,
					AccessCert:       u.AccessCert,
					AccessCertExpiry: u.AccessCertExpiry,
				},
			}, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
