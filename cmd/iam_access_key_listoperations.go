package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type iamAccessKeyListOperationsItemOutput struct {
	Operation string   `json:"operation"`
	Tags      []string `json:"tags"`
}

type iamAccessKeyListOperationsOutput []iamAccessKeyListOperationsItemOutput

func (o *iamAccessKeyListOperationsOutput) ToJSON() { output.JSON(o) }
func (o *iamAccessKeyListOperationsOutput) ToText() { output.Text(o) }
func (o *iamAccessKeyListOperationsOutput) ToTable() {
	operationByTag := make(map[string][]string)

	for _, op := range *o {
		for _, tag := range op.Tags {
			if _, ok := operationByTag[tag]; !ok {
				operationByTag[tag] = make([]string, 0)
			}
			operationByTag[tag] = append(operationByTag[tag], op.Operation)
		}
	}

	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"Tag", "Operations"})

	sortedTags := make([]string, 0)
	for tag := range operationByTag {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	for _, tag := range sortedTags {
		operations := operationByTag[tag]
		sort.Strings(operations)
		t.Append([]string{tag, strings.Join(operations, "\n")})
	}
}

type iamAccessKeyListOperationsCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list-operations"`

	Mine bool `cli-usage:"only report operations available to the IAM access key used to perform the API request"`
}

func (c *iamAccessKeyListOperationsCmd) cmdAliases() []string { return gListAlias }

func (c *iamAccessKeyListOperationsCmd) cmdShort() string { return "List IAM access keys operations" }

func (c *iamAccessKeyListOperationsCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists operations available to IAM access keys.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&iamAccessKeyListOperationsOutput{}), ", "))
}

func (c *iamAccessKeyListOperationsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyListOperationsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var (
		iamAccessKeyOperations []*egoscale.IAMAccessKeyOperation
		err                    error
	)

	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	if c.Mine {
		iamAccessKeyOperations, err = globalstate.GlobalEgoscaleClient.ListMyIAMAccessKeyOperations(ctx, zone)
	} else {
		iamAccessKeyOperations, err = globalstate.GlobalEgoscaleClient.ListIAMAccessKeyOperations(ctx, zone)
	}
	if err != nil {
		return err
	}

	out := make(iamAccessKeyListOperationsOutput, 0)

	for _, o := range iamAccessKeyOperations {
		out = append(out, iamAccessKeyListOperationsItemOutput{
			Operation: o.Name,
			Tags:      o.Tags,
		})
	}

	return c.outputFunc(&out, err)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAccessKeyCmd, &iamAccessKeyListOperationsCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
