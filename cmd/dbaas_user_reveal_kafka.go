package cmd

import (
	"context"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasKafkaUserRevealOutput struct {
	AccessKey string `json:"access-key,omitempty"`
}

func (o *dbaasKafkaUserRevealOutput) formatUser(t *table.Table) {
	t.Append([]string{"Access Key", o.AccessKey})
}

func (c *dbaasUserRevealCmd) revealKafka(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
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
			AccessKey: s.AccessKey,
		},
	}, nil
}
