package dns

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

var dnsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add record to domain",
}

func init() {
	dnsCmd.AddCommand(dnsAddCmd)
}

func addDomainRecord(domainIdent, name, rType, content string, ttl int64, priority *int64) error {

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(domainIdent)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found")
		}
		return err
	}

	req := v3.CreateDNSDomainRecordRequest{
		Name:    name,
		Type:    v3.CreateDNSDomainRecordRequestType(rType),
		Content: content,
		Ttl:     ttl,
	}
	if priority != nil {
		req.Priority = *priority
	}

	op, err := client.CreateDNSDomainRecord(ctx, domain.ID, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Adding DNS record %q to %q...", rType, domain.UnicodeName), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was created successfully to %q\n", rType, domain.UnicodeName)
	}

	return nil
}
