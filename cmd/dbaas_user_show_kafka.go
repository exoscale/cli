package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
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

func (c *dbaasUserShowCmd) showKafka(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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
				Type:     u.Type,
				Kafka: &dbaasKafkaUserShowOutput{
					AccessCert:       u.AccessCert,
					AccessCertExpiry: u.AccessCertExpiry,
					AccessKey:        "xxxxxx",
				},
			}, nil
		}

	}

	return &dbaasUserShowOutput{}, fmt.Errorf("user %q not found for service %q", c.Username, c.Name)
}
