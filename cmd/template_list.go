package cmd

import (
	"os"
	"strconv"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

func init() {
	templateListCmd.Flags().BoolP("iso", "i", false, "List ISOs")
	templateCmd.AddCommand(templateListCmd)
}

// templateListCmd represents the list command
var templateListCmd = &cobra.Command{
	Use:     "list [keyword]",
	Short:   "List all available templates",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		t := table.NewTable(os.Stdout)

		iso, err := cmd.Flags().GetBool("iso")
		if err != nil {
			return err
		}
		if iso {
			return listISOs()
		}

		err = listTemplates(t, args)
		if err == nil {
			t.Render()
		}
		return err
	},
}

func listISOs() error {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Name", "Size", "Zone", "ID"})

	resp, err := cs.ListWithContext(gContext, &egoscale.ListISOs{})
	if err != nil {
		return err
	}

	for _, i := range resp {
		iso := i.(*egoscale.ISO)
		sz := humanize.IBytes(uint64(iso.Size))
		t.Append([]string{iso.Name, sz, iso.ZoneName, iso.ID.String()})
	}
	t.Render()

	return nil
}

func listTemplates(t *table.Table, filters []string) error {
	zoneID, err := getZoneIDByName(gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	templates, err := findTemplates(zoneID, filters...)
	if err != nil {
		return err
	}

	t.SetHeader([]string{"Operating System", "Disk", "Release Date", "ID"})
	for _, template := range templates {
		sz := strconv.FormatInt(template.Size>>30, 10)
		if sz == "10" && strings.HasPrefix(template.Name, "Linux") {
			sz = ""
		}
		t.Append([]string{template.Name, sz, template.Created, template.ID.String()})
	}

	return nil
}
