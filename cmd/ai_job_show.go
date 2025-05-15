package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exossh "github.com/exoscale/cli/pkg/ssh"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var (
	// Job status command
	jobStatusCommand = "sudo kubectl get job %s"
	defaultAIJobName = "ai-job"
)

type aiJobShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	// SSH options
	SshKey  string `cli-short:"k" cli-flag:"ssh-key" cli-usage:"instance ssh private key"`
	SshUser string `cli-short:"u" cli-flag:"ssh-user" cli-usage:"instance ssh user"`
	SshPort string `cli-short:"p" cli-flag:"ssh-port" cli-usage:"instance ssh port"`

	// AI job options
	JobName string `cli-short:"j" cli-flag:"job-name" cli-usage:"name of the AI job to show"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *aiJobShowCmd) cmdAliases() []string { return gShowAlias }

func (c *aiJobShowCmd) cmdShort() string { return "Show an AI job status" }

func (c *aiJobShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an AI job status.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *aiJobShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *aiJobShowCmd) cmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	// Assign the job command
	if c.JobName != "" {
		jobStatusCommand = fmt.Sprintf(jobStatusCommand, c.JobName)
	} else {
		jobStatusCommand = fmt.Sprintf(jobStatusCommand, defaultAIJobName)
	}

	// SSH Port
	if c.SshPort == "" {
		c.SshPort = "22"
	}
	// SSH User
	if c.SshUser == "" {
		c.SshUser = "debian"
	}

	// Connect via the SSH tunnel and issue the command to check the job
	cmdResponse, err := exossh.RunCmd(instance, c.SshUser, c.SshPort, c.SshKey, jobStatusCommand)
	if err != nil {
		return err
	}

	out := instanceShowOutput{
		AntiAffinityGroups: make([]string, 0),
		CreationDate:       instance.CreatedAt.String(),
		DiskSize:           humanize.IBytes(uint64(*instance.DiskSize << 30)),
		ElasticIPs:         make([]string, 0),
		ID:                 *instance.ID,
		IPAddress:          utils.DefaultIP(instance.PublicIPAddress, "-"),
		IPv6Address:        utils.DefaultIP(instance.IPv6Address, "-"),
		Labels: func() (v map[string]string) {
			if instance.Labels != nil {
				v = *instance.Labels
			}
			return
		}(),
		Name:            *instance.Name,
		PrivateNetworks: make([]string, 0),
		SSHKey:          utils.DefaultString(instance.SSHKey, "-"),
		SecurityGroups:  make([]string, 0),
		State:           *instance.State,
		Zone:            c.Zone,
		AIJobStatus:     cmdResponse,
	}

	out.PrivateInstance = "No"
	if instance.PublicIPAssignment != nil && *instance.PublicIPAssignment == "none" {
		out.PrivateInstance = "Yes"
	}

	if instance.AntiAffinityGroupIDs != nil {
		for _, id := range *instance.AntiAffinityGroupIDs {
			antiAffinityGroup, err := globalstate.EgoscaleClient.GetAntiAffinityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Anti-Affinity Group: %w", err)
			}
			out.AntiAffinityGroups = append(out.AntiAffinityGroups, *antiAffinityGroup.Name)
		}
	}

	out.DeployTarget = "-"
	if instance.DeployTargetID != nil {
		DeployTarget, err := globalstate.EgoscaleClient.GetDeployTarget(ctx, c.Zone, *instance.DeployTargetID)
		if err != nil {
			return fmt.Errorf("error retrieving Deploy Target: %w", err)
		}
		out.DeployTarget = *DeployTarget.Name
	}

	if instance.ElasticIPIDs != nil {
		for _, id := range *instance.ElasticIPIDs {
			elasticIP, err := globalstate.EgoscaleClient.GetElasticIP(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Elastic IP: %w", err)
			}
			out.ElasticIPs = append(out.ElasticIPs, elasticIP.IPAddress.String())
		}
	}

	instanceType, err := globalstate.EgoscaleClient.GetInstanceType(ctx, c.Zone, *instance.InstanceTypeID)
	if err != nil {
		return err
	}
	out.InstanceType = fmt.Sprintf("%s.%s", *instanceType.Family, *instanceType.Size)

	if instance.PrivateNetworkIDs != nil {
		for _, id := range *instance.PrivateNetworkIDs {
			privateNetwork, err := globalstate.EgoscaleClient.GetPrivateNetwork(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Private Network: %w", err)
			}
			out.PrivateNetworks = append(out.PrivateNetworks, *privateNetwork.Name)
		}
	}

	if instance.SecurityGroupIDs != nil {
		for _, id := range *instance.SecurityGroupIDs {
			securityGroup, err := globalstate.EgoscaleClient.GetSecurityGroup(ctx, c.Zone, id)
			if err != nil {
				return fmt.Errorf("error retrieving Security Group: %w", err)
			}
			out.SecurityGroups = append(out.SecurityGroups, *securityGroup.Name)
		}
	}

	template, err := globalstate.EgoscaleClient.GetTemplate(ctx, c.Zone, *instance.TemplateID)
	if err != nil {
		return err
	}
	out.Template = *template.Name

	rdns, err := globalstate.EgoscaleClient.GetInstanceReverseDNS(ctx, c.Zone, *instance.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			out.ReverseDNS = ""
		} else {
			return err
		}
	}

	out.ReverseDNS = rdns

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(aiJobCmd, &aiJobShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
