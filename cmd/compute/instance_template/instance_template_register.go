package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceTemplateRegisterCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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

func (c *instanceTemplateRegisterCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *instanceTemplateRegisterCmd) CmdShort() string {
	return "Register a new Compute instance template"
}

func (c *instanceTemplateRegisterCmd) CmdLong() string {
	return fmt.Sprintf(`This command registers a new Compute instance template.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTemplateShowOutput{}), ", "))
}

func (c *instanceTemplateRegisterCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)

	// In case the user specified a snapshot ID using the `--from-snapshot` flag,
	// we add empty positional argument placeholders in order to trick the
	// exocmd.CliCommandDefaultPreRun() wrapper into believing URL/Checksum args were provided,
	// but the actual command function won't use them since it will dynamically retrieve
	// this information from the specified snapshot export information.

	snapshotID, err := cmd.Flags().GetString(exocmd.MustCLICommandFlagName(c, &c.FromSnapshot))
	if err != nil {
		return err
	}
	if snapshotID != "" {
		args = append(args, "", "")
	}

	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateRegisterCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var (
		templateRequest v3.RegisterTemplateRequest
		err             error
	)

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	passwordEnabled := !c.DisablePassword
	sshKeyEnabled := !c.DisableSSHKey

	templateRequest = v3.RegisterTemplateRequest{
		Checksum:        c.Checksum,
		DefaultUser:     c.Username,
		Description:     c.Description,
		Build:           c.Build,
		Version:         c.Version,
		Maintainer:      c.Maintainer,
		Name:            c.Name,
		PasswordEnabled: &passwordEnabled,
		SSHKeyEnabled:   &sshKeyEnabled,
		URL:             c.URL,
	}

	if c.FromSnapshot != "" {

		snapshots, err := client.ListSnapshots(ctx)
		if err != nil {
			return err
		}
		snapshot, err := snapshots.FindSnapshot(c.FromSnapshot)
		if err != nil {
			return err
		}

		op, err := client.ExportSnapshot(ctx, snapshot.ID)
		utils.DecorateAsyncOperation("exporting snapshot...", func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return fmt.Errorf("error retrieving snapshot export information: %w", err)
		}

		snapshotExport, err := client.GetSnapshot(ctx, snapshot.ID)
		if err != nil {
			return err
		}

		templateRequest.URL = snapshotExport.Export.PresignedURL
		templateRequest.Checksum = snapshotExport.Export.Md5sum

		// Pre-setting the new template properties from the source template.
		instance, err := client.GetInstance(ctx, snapshot.Instance.ID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance from snapshot: %w", err)
		}

		srcTemplate, err := client.GetTemplate(ctx, instance.Template.ID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance template from snapshot: %w", err)
		}

		templateRequest.BootMode = v3.RegisterTemplateRequestBootMode(srcTemplate.BootMode)

		// Above properties are inherited from snapshot source template, unless otherwise specified
		// by the user from the command line
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DisablePassword)) {
			templateRequest.PasswordEnabled = &passwordEnabled
		} else {
			templateRequest.PasswordEnabled = srcTemplate.PasswordEnabled
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DisableSSHKey)) {
			templateRequest.SSHKeyEnabled = &sshKeyEnabled
		} else {
			templateRequest.SSHKeyEnabled = srcTemplate.SSHKeyEnabled
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Username)) {
			templateRequest.DefaultUser = c.Username
		} else {
			templateRequest.DefaultUser = srcTemplate.DefaultUser
		}
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.BootMode)) {
		templateRequest.BootMode = v3.RegisterTemplateRequestBootMode(c.BootMode)
	}

	op, err := client.RegisterTemplate(ctx, templateRequest)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Registering template %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	template, err := client.GetTemplate(ctx, op.Reference.ID)
	if err != nil {
		return fmt.Errorf("error retrieving newly registered template: %w", err)
	}

	if !globalstate.Quiet {
		return c.OutputFunc(&instanceTemplateShowOutput{
			ID:              template.ID.String(),
			Zone:            c.Zone,
			Family:          template.Family,
			Name:            template.Name,
			Description:     template.Description,
			CreationDate:    template.CreatedAT.String(),
			Visibility:      string(template.Visibility),
			Size:            template.Size,
			Version:         template.Version,
			Build:           template.Build,
			Maintainer:      template.Maintainer,
			Checksum:        template.Checksum,
			DefaultUser:     template.DefaultUser,
			SSHKeyEnabled:   utils.DefaultBool(template.SSHKeyEnabled, false),
			PasswordEnabled: utils.DefaultBool(template.PasswordEnabled, false),
			BootMode:        string(template.BootMode),
		}, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTemplateCmd, &instanceTemplateRegisterCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		BootMode: "legacy",

		// Template registration can take a _long time_, raising
		// the Exoscale API client timeout to 1h by default as a precaution.
		Timeout: 3600,
	}))
}
