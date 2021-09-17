package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"
)

type instanceSCPCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	scpInfo struct {
		ipAddress string
		keyFile   string
	} `cli-cmd:"-"`
	_ bool `cli-cmd:"scp"`

	Instance string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	Source   string `cli-arg:"#"`
	Target   string `cli-arg:"#"`

	IPv6      bool   `cli-flag:"ipv6" cli-short:"6" cli-help:"connect to the instance via its IPv6 address"`
	Login     string `cli-short:"l" cli-help:"SCP username to use for logging in (default: instance template default username)"`
	PrintCmd  bool   `cli-flag:"print-command" cli-usage:"print the SCP command that would be executed instead of executing it"`
	Recursive bool   `cli-short:"r" cli-usage:"recursively copy entire directories"`
	ReplStr   string `cli-flag:"replace-str" cli-short:"i" cli-usage:"string to replace with the actual Compute instance information (i.e. username@IP-ADDRESS)"`
	SCPOpts   string `cli-flag:"scp-options" cli-short:"o" cli-usage:"additional options to pass to the scp(1) command"`
	Zone      string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSCPCmd) buildSCPCommand() []string {
	cmd := []string{"scp"}

	if _, err := os.Stat(c.scpInfo.keyFile); err == nil {
		cmd = append(cmd, "-i", c.scpInfo.keyFile)
	}

	if c.Recursive {
		cmd = append(cmd, "-r")
	}

	if c.SCPOpts != "" {
		opts, err := shellquote.Split(c.SCPOpts)
		if err == nil {
			cmd = append(cmd, opts...)
		}
	}

	// Parse arguments to find the replacement string to interpolate with <[username@]target instance IP address>
	for _, arg := range []string{c.Source, c.Target} {
		if strings.Contains(arg, c.ReplStr) {
			remote := c.scpInfo.ipAddress
			if c.Login != "" {
				remote = fmt.Sprintf("%s@%s", c.Login, remote)
			}

			arg = strings.Replace(arg, c.ReplStr, remote, 1)
		}
		cmd = append(cmd, arg)
	}

	return cmd
}

func (c *instanceSCPCmd) cmdAliases() []string { return nil }

func (c *instanceSCPCmd) cmdShort() string { return "SCP files to/from a Compute instance" }

func (c *instanceSCPCmd) cmdLong() string {
	return `This command executes the scp(1) command to send or receive files to/from the
specified Compute instance. TARGET (or SOURCE, depending on the direction
of the transfer) must contain the "{}" marker which will be interpolated at
run time with the actual Compute instance IP address, similar to xargs(1).
This marker can be replaced by another string via the --replace-str|-i flag.

Example:

    exo compute instance scp my-instance hello-world.txt {}:
    exo compute instance scp -i%% my-instance %%:/etc/motd .
`
}

func (c *instanceSCPCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSCPCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if c.Login == "" {
		instanceTemplate, err := cs.GetTemplate(ctx, c.Zone, *instance.TemplateID)
		if err != nil {
			return fmt.Errorf("error retrieving instance template: %s", err)
		}
		if instanceTemplate.DefaultUser != nil {
			c.Login = *instanceTemplate.DefaultUser
		}
	}

	c.scpInfo.keyFile = getInstanceSSHKeyPath(*instance.ID)

	c.scpInfo.ipAddress = instance.PublicIPAddress.String()
	if c.IPv6 {
		if instance.IPv6Address == nil {
			return fmt.Errorf("instance %q has no IPv6 address", c.Instance)
		}
		c.scpInfo.ipAddress = instance.IPv6Address.String()
	}

	scpCmd := c.buildSCPCommand()

	if c.PrintCmd {
		fmt.Println(strings.Join(scpCmd, " "))
		return nil
	}

	cmd := exec.Command("scp", scpCmd[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceSCPCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		ReplStr: "{}",
	}))
}
