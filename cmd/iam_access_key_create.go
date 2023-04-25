package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type iamAccessKeyCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Operations []string `cli-flag:"operation" cli-usage:"API operation to restrict the access key to. Can be repeated multiple times."`
	Resources  []string `cli-flag:"resource" cli-usage:"API resource to restrict the access key to (format: DOMAIN/TYPE:NAME). Can be repeated multiple times."`
	Tags       []string `cli-flag:"tag" cli-usage:"API operations tag to restrict the access key to. Can be repeated multiple times."`
}

func (c *iamAccessKeyCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *iamAccessKeyCreateCmd) cmdShort() string {
	return "Create an IAM access key"
}

func (c *iamAccessKeyCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an IAM access key.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&iamAccessKeyShowOutput{}), ", "))
}

func (c *iamAccessKeyCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	opts := make([]egoscale.CreateIAMAccessKeyOpt, 0)

	if len(c.Operations) > 0 {
		opts = append(opts, egoscale.CreateIAMAccessKeyWithOperations(c.Operations))
	}

	if len(c.Resources) > 0 {
		resources := make([]egoscale.IAMAccessKeyResource, len(c.Resources))
		for i, rs := range c.Resources {
			r, err := parseIAMAccessKeyResource(rs)
			if err != nil {
				return fmt.Errorf("invalid API resource %q", rs)
			}
			resources[i] = *r
		}
		opts = append(opts, egoscale.CreateIAMAccessKeyWithResources(resources))
	}

	if len(c.Tags) > 0 {
		opts = append(opts, egoscale.CreateIAMAccessKeyWithTags(c.Tags))
	}

	iamAccessKey, err := globalstate.GlobalEgoscaleClient.CreateIAMAccessKey(ctx, zone, c.Name, opts...)
	if err != nil {
		return fmt.Errorf("unable to create a new IAM access key: %w", err)
	}

	out := iamAccessKeyShowOutput{
		Name:       *iamAccessKey.Name,
		APIKey:     *iamAccessKey.Key,
		APISecret:  iamAccessKey.Secret,
		Operations: iamAccessKey.Operations,
		Resources: func() *[]string {
			if iamAccessKey.Resources != nil {
				list := make([]string, len(*iamAccessKey.Resources))
				for i, r := range *iamAccessKey.Resources {
					list[i] = fmt.Sprintf("%s/%s:%s", r.Domain, r.ResourceType, r.ResourceName)
				}
				return &list
			}
			return nil
		}(),
		Tags: iamAccessKey.Tags,
		Type: *iamAccessKey.Type,
	}

	if !globalstate.Quiet {
		return c.outputFunc(&out, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAccessKeyCmd, &iamAccessKeyCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
