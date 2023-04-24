package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceTemplateRegisterCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"register"`

	Name     string `cli-arg:"#"`
	URL      string `cli-arg:"#"`
	Checksum string `cli-arg:"#"`

	BootMode        string `cli-usage:"template boot mode (legacy|uefi)"`
	Description     string `cli-usage:"template description"`
	Build           string `cli-usage:"template build"`
	Version         string `cli-usage:"template version"`
	Maintainer      string `cli-usage:"template maintainer"`
	DisablePassword bool   `cli-usage:"disable password-based authentication"`
	DisableSSHKey   bool   `cli-flag:"disable-ssh-key" cli-usage:"disable SSH key-based authentication"`
	FromSnapshot    string `cli-usage:"ID of a Compute instance snapshot to register as template"`
	Timeout         int64  `cli-usage:"registration timeout duration in seconds"`
	Username        string `cli-usage:"template default username"`
	Zone            string `cli-short:"z" cli-usage:"zone to register the template into (default: current account's default zone)"`
}

func (c *instanceTemplateRegisterCmd) cmdAliases() []string { return gCreateAlias }

func (c *instanceTemplateRegisterCmd) cmdShort() string {
	return "Register a new Compute instance template"
}

func (c *instanceTemplateRegisterCmd) cmdLong() string {
	return fmt.Sprintf(`This command registers a new Compute instance template.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceTemplateShowOutput{}), ", "))
}

func (c *instanceTemplateRegisterCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)

	// In case the user specified a snapshot ID using the `--from-snapshot` flag,
	// we add empty positional argument placeholders in order to trick the
	// cliCommandDefaultPreRun() wrapper into believing URL/Checksum args were provided,
	// but the actual command function won't use them since it will dynamically retrieve
	// this information from the specified snapshot export information.

	snapshotID, err := cmd.Flags().GetString(mustCLICommandFlagName(c, &c.FromSnapshot))
	if err != nil {
		return err
	}
	if snapshotID != "" {
		args = append(args, "", "")
	}

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateRegisterCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var (
		template *egoscale.Template
		err      error
	)

	cs.Client.SetTimeout(time.Duration(c.Timeout) * time.Second)

	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	passwordEnabled := !c.DisablePassword
	sshKeyEnabled := !c.DisableSSHKey

	template = &egoscale.Template{
		Checksum:        utils.NonEmptyStringPtr(c.Checksum),
		DefaultUser:     utils.NonEmptyStringPtr(c.Username),
		Description:     utils.NonEmptyStringPtr(c.Description),
		Build:           utils.NonEmptyStringPtr(c.Build),
		Version:         utils.NonEmptyStringPtr(c.Version),
		Maintainer:      utils.NonEmptyStringPtr(c.Maintainer),
		Name:            &c.Name,
		PasswordEnabled: &passwordEnabled,
		SSHKeyEnabled:   &sshKeyEnabled,
		URL:             utils.NonEmptyStringPtr(c.URL),
	}

	if c.FromSnapshot != "" {
		snapshot, err := cs.GetSnapshot(ctx, c.Zone, c.FromSnapshot)
		if err != nil {
			return fmt.Errorf("error retrieving snapshot: %w", err)
		}

		snapshotExport, err := cs.ExportSnapshot(ctx, c.Zone, snapshot)
		if err != nil {
			return fmt.Errorf("error retrieving snapshot export information: %w", err)
		}

		template.URL = snapshotExport.PresignedURL
		template.Checksum = snapshotExport.MD5sum

		// Pre-setting the new template properties from the source template.
		instance, err := cs.GetInstance(ctx, c.Zone, *snapshot.InstanceID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance from snapshot: %w", err)
		}

		srcTemplate, err := cs.GetTemplate(ctx, c.Zone, *instance.TemplateID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance template from snapshot: %w", err)
		}

		template.BootMode = srcTemplate.BootMode

		// Above properties are inherited from snapshot source template, unless otherwise specified
		// by the user from the command line
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DisablePassword)) {
			template.PasswordEnabled = &passwordEnabled
		} else {
			template.PasswordEnabled = srcTemplate.PasswordEnabled
		}

		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DisableSSHKey)) {
			template.SSHKeyEnabled = &sshKeyEnabled
		} else {
			template.SSHKeyEnabled = srcTemplate.SSHKeyEnabled
		}

		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Username)) {
			template.DefaultUser = utils.NonEmptyStringPtr(c.Username)
		} else {
			template.DefaultUser = srcTemplate.DefaultUser
		}
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.BootMode)) {
		template.BootMode = &c.BootMode
	}

	decorateAsyncOperation(fmt.Sprintf("Registering template %q...", c.Name), func() {
		template, err = cs.RegisterTemplate(ctx, c.Zone, template)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return c.outputFunc(&instanceTemplateShowOutput{
			ID:              *template.ID,
			Family:          utils.DefaultString(template.Family, ""),
			Name:            *template.Name,
			Description:     utils.DefaultString(template.Description, ""),
			CreationDate:    template.CreatedAt.String(),
			Visibility:      *template.Visibility,
			Size:            *template.Size,
			Version:         utils.DefaultString(template.Version, ""),
			Build:           utils.DefaultString(template.Build, ""),
			Maintainer:      utils.DefaultString(template.Maintainer, ""),
			Checksum:        *template.Checksum,
			DefaultUser:     utils.DefaultString(template.DefaultUser, ""),
			SSHKeyEnabled:   *template.SSHKeyEnabled,
			PasswordEnabled: *template.PasswordEnabled,
			BootMode:        *template.BootMode,
		}, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTemplateCmd, &instanceTemplateRegisterCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		BootMode: "legacy",

		// Template registration can take a _long time_, raising
		// the Exoscale API client timeout to 1h by default as a precaution.
		Timeout: 3600,
	}))
}
