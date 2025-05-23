package cmd

import (
	"context"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasKafkaUserRevealOutput struct {
	AccessKey        string    `json:"access-key,omitempty"`
	AccessCert       string    `json:"access-cert,omitempty"`
	AccessCertExpiry time.Time `json:"access-cert-expiry,omitempty"`
}

func (o *dbaasKafkaUserRevealOutput) formatUser(t *table.Table) {
	t.Append([]string{"Access Cert", o.AccessCert})
	t.Append([]string{"Access Key", o.AccessKey})
	t.Append([]string{"Access Cert Expiry", o.AccessCertExpiry.String()})

}

func (c *dbaasUserRevealCmd) revealKafka(ctx context.Context) (output.Outputter, error) {

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	s, err := client.RevealDBAASKafkaUserPassword(ctx, c.Name, c.Username)
	if err != nil {
		return &dbaasUserRevealOutput{}, err
	}

	return &dbaasUserRevealOutput{
		Password: s.Password,
		Kafka: &dbaasKafkaUserRevealOutput{
			AccessKey:        s.AccessKey,
			AccessCert:       s.AccessCert,
			AccessCertExpiry: s.AccessCertExpiry,
		},
	}, nil
}
