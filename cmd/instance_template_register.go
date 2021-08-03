package cmd

import (
	"fmt"
	"strings"
	"time"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeInstanceTemplateRegisterCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"register"`

	Name     string `cli-arg:"#"`
	URL      string `cli-arg:"#"`
	Checksum string `cli-arg:"#"`

	BootMode        string `cli-usage:"template boot mode (legacy|uefi)"`
	Description     string `cli-usage:"template description"`
	DisablePassword bool   `cli-usage:"disable password-based authentication"`
	DisableSSHKey   bool   `cli-flag:"disable-ssh-key" cli-usage:"disable SSH key-based authentication"`
	FromSnapshot    string `cli-usage:"ID of a Compute instance snapshot to register as template"`
	Username        string `cli-usage:"template default username"`
	Zone            string `cli-short:"z" cli-usage:"zone to register the template into (default: current account's default zone)"`
}

func (c *computeInstanceTemplateRegisterCmd) cmdAliases() []string { return gCreateAlias }

func (c *computeInstanceTemplateRegisterCmd) cmdShort() string {
	return "Register a new Compute instance template"
}

func (c *computeInstanceTemplateRegisterCmd) cmdLong() string {
	return fmt.Sprintf(`This command registers a new Compute instance template.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&computeInstanceTemplateShowOutput{}), ", "))
}

func (c *computeInstanceTemplateRegisterCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
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

func (c *computeInstanceTemplateRegisterCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var (
		template *egoscale.Template
		err      error
	)

	// Template registration can take a _long time_, raising
	// the Exoscale API client timeout as a precaution.
	cs.Client.SetTimeout(30 * time.Minute)

	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	template = &egoscale.Template{
		Name: &c.Name,
	}

	if c.FromSnapshot != "" {
		snapshot, err := cs.GetSnapshot(ctx, c.Zone, c.FromSnapshot)
		if err != nil {
			return fmt.Errorf("error retrieving snapshot: %s", err)
		}

		snapshotExport, err := snapshot.Export(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving snapshot export information: %s", err)
		}

		c.URL = *snapshotExport.PresignedURL
		c.Checksum = *snapshotExport.MD5sum

		// Pre-setting the new template properties from the source template.
		instance, err := cs.GetInstance(ctx, c.Zone, *snapshot.InstanceID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance from snapshot: %s", err)
		}

		srcTemplate, err := cs.GetTemplate(ctx, c.Zone, *instance.TemplateID)
		if err != nil {
			return fmt.Errorf("error retrieving Compute instance template from snapshot: %s", err)
		}

		template.BootMode = srcTemplate.BootMode
		template.PasswordEnabled = srcTemplate.PasswordEnabled
		template.SSHKeyEnabled = srcTemplate.SSHKeyEnabled
		template.DefaultUser = srcTemplate.DefaultUser
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.BootMode)) {
		template.BootMode = &c.BootMode
	}

	if c.Checksum != "" {
		template.Checksum = &c.Checksum
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Username)) {
		template.DefaultUser = &c.Username
	}

	if c.Description != "" {
		template.Description = &c.Description
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DisablePassword)) {
		passwordEnabled := !c.DisablePassword
		template.PasswordEnabled = &passwordEnabled
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DisableSSHKey)) {
		sshKeyEnabled := !c.DisableSSHKey
		template.SSHKeyEnabled = &sshKeyEnabled
	}

	if c.URL != "" {
		template.URL = &c.URL
	}

	decorateAsyncOperation(fmt.Sprintf("Registering template %q...", c.Name), func() {
		template, err = cs.RegisterTemplate(ctx, c.Zone, template)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(&computeInstanceTemplateShowOutput{
			ID:              *template.ID,
			Family:          defaultString(template.Family, ""),
			Name:            *template.Name,
			Description:     defaultString(template.Description, ""),
			CreationDate:    template.CreatedAt.String(),
			Visibility:      *template.Visibility,
			Size:            *template.Size,
			Version:         defaultString(template.Version, ""),
			Build:           defaultString(template.Build, ""),
			Checksum:        *template.Checksum,
			DefaultUser:     defaultString(template.DefaultUser, ""),
			SSHKeyEnabled:   *template.SSHKeyEnabled,
			PasswordEnabled: *template.PasswordEnabled,
			BootMode:        *template.BootMode,
		}, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceTemplateCmd, &computeInstanceTemplateRegisterCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		BootMode: "legacy",
	}))
}
