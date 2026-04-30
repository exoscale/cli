package apikey

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyListItemOutput struct {
	ID         v3.UUID     `json:"id" outputWidth:"36"`
	Name       string      `json:"name" outputWidth:"40"`
	Zone       v3.ZoneName `json:"zone" outputWidth:"8"`
	CreationAt string      `json:"created_at" outputWidth:"20"`
	RevokedAt  string      `json:"revoked_at" outputWidth:"20"`
}

type AIAPIKeyListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *AIAPIKeyListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *AIAPIKeyListCmd) CmdShort() string     { return "List AI API keys" }
func (c *AIAPIKeyListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI API keys across all zones.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&AIAPIKeyListItemOutput{}), ", "))
}
func (c *AIAPIKeyListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runAIAPIKeyList(c, os.Stdout, os.Stderr)
}

func runAIAPIKeyList(c *AIAPIKeyListCmd, stdout, stderr io.Writer) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	var (
		mu    sync.Mutex
		items []AIAPIKeyListItemOutput
	)

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListAIAPIKeys(ctx)
			if err != nil {
				return err
			}
			for _, key := range resp.AIAPIKeys {
				revokedAt := ""
				if key.RevokedAT != nil {
					revokedAt = key.RevokedAT.Format(time.RFC3339)
				}
				mu.Lock()
				items = append(items, AIAPIKeyListItemOutput{
					ID:         key.ID,
					Name:       key.Name,
					Zone:       zone.Name,
					CreationAt: key.CreatedAT.Format(time.RFC3339),
					RevokedAt:  revokedAt,
				})
				mu.Unlock()
			}
			return nil
		})

	sort.Slice(items, func(i, j int) bool {
		ri := items[i].RevokedAt != ""
		rj := items[j].RevokedAt != ""
		if ri != rj {
			return !ri // revoked keys go to the bottom
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].Zone < items[j].Zone
	})

	streamer := output.NewStreamer(AIAPIKeyListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	for _, item := range items {
		if err := streamer.Push(item); err != nil {
			return err
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d zone(s) failed", failed)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
