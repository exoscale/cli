package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exossh "github.com/exoscale/cli/pkg/ssh"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var labAIJobCmd = &cobra.Command{
	Use:   "job",
	Short: "AI jobs management",
}

func init() {
	labAICmd.AddCommand(labAIJobCmd)
}

// Create command

type labAIJobCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	ContainerImage    string `cli-flag:"container-image" cli-usage:"container image to use for the AI job"`
	HuggingFaceSecret string `cli-flag:"hf-secret" cli-usage:"HuggingFace secret to use for the AI job"`
	Model             string `cli-flag:"model" cli-usage:"model to use for the AI job"`
	JobName           string `cli-flag:"job-name" cli-usage:"name of the AI job to create"`

	InstanceCreateCmd
}

func (c *labAIJobCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *labAIJobCreateCmd) cmdShort() string { return "Create an AI job Instance" }

var gpuInstanceTypeFamilies = []string{
	"gpu",
	"gpu2",
	"gpu3",
}

var gpuInstanceTypeSizes = []string{
	"medium",
	"large",
	"huge",
}

func (c *labAIJobCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates an AI Job Instance.

Supported Compute instance type families: %s

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(gpuInstanceTypeFamilies, ", "),
		strings.Join(gpuInstanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *labAIJobCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	if err := c.InstanceCreateCmd.cmdPreRun(cmd, args); err != nil {
		return err
	}

	return cmdCheckRequiredFlags(cmd, []string{"hf-secret"})
}

const (
	defaultAIJobName = "ai-job"
	// Creates a PVC to store the fine-tuning check points
	aiJobPVCTemplate = `
#cloud-config
write_files:
 - path: /var/lib/rancher/k3s/server/manifests/ai-job.yaml
   content: |
     apiVersion: ai.re-cinq.com/v1
     kind: Job
     metadata:
       name: ` + defaultAIJobName + `
     spec:
       image: %s
       model: %s
       diskSize: %d
       huggingFaceSecret: %s
   owner: 'root:root'
   permissions: '0640'
`
)

func (c *labAIJobCreateCmd) cmdRun(cmd *cobra.Command, args []string) error { //nolint:gocyclo
	aiJobConfig := fmt.Sprintf(aiJobPVCTemplate,
		c.ContainerImage,
		c.Model,
		c.DiskSize/2,
		c.HuggingFaceSecret,
	)

	cloudInitFile, err := os.CreateTemp("", "cloud-init")
	if err != nil {
		return err
	}
	defer os.Remove(cloudInitFile.Name())

	_, err = cloudInitFile.Write([]byte(aiJobConfig))
	if err != nil {
		return err
	}
	c.CloudInitFile = cloudInitFile.Name()

	return c.InstanceCreateCmd.cmdRun(cmd, args)
}

func init() {
	cobra.CheckErr(registerCLICommand(labAIJobCmd, &labAIJobCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		// Here is some default values.
		InstanceCreateCmd: InstanceCreateCmd{
			cliCommandSettings: defaultCLICmdSettings(),
			DiskSize:           50,
			InstanceType:       fmt.Sprintf("%s.%s", "gpu3", "small"), // Default to gpu3.small
			TemplateVisibility: "private",                             // TODO change it to defaultTemplateVisibility (public) once this template is published
			Template:           "linux-debian-12-gpu",
		},
	}))
}

// Delete command

type labAIJobDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	InstanceDeleteCmd
}

func (c *labAIJobDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *labAIJobDeleteCmd) cmdShort() string { return "Delete an AI job Instance" }

func (c *labAIJobDeleteCmd) cmdLong() string { return "" }

func (c *labAIJobDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return c.InstanceDeleteCmd.cmdPreRun(cmd, args)
}

func (c *labAIJobDeleteCmd) cmdRun(cmd *cobra.Command, args []string) error {
	return c.InstanceDeleteCmd.cmdRun(cmd, args)
}

func init() {
	cobra.CheckErr(registerCLICommand(labAIJobCmd, &labAIJobDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
		InstanceDeleteCmd: InstanceDeleteCmd{
			cliCommandSettings: defaultCLICmdSettings(),
		},
	}))
}

// Show command

type labAIJobShowOutput struct {
	Instance    string
	InstanceID  string
	AIJobStatus string `json:"ai_job_status"`
}

func (o *labAIJobShowOutput) Type() string { return "AI Job instance" }
func (o *labAIJobShowOutput) ToJSON()      { output.JSON(o) }
func (o *labAIJobShowOutput) ToText()      { output.Text(o) }
func (o *labAIJobShowOutput) ToTable()     { output.Table(o) }

var (
	// Job status command
	jobStatusCommand = "sudo kubectl get job %s"
)

type labAIJobShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	AIJobInstance string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone          v3.ZoneName `cli-short:"z" cli-usage:"AI job Instance zone"`

	// SSH options
	SshKey  string `cli-short:"k" cli-flag:"ssh-key" cli-usage:"instance ssh private key"`
	SshUser string `cli-short:"u" cli-flag:"ssh-user" cli-usage:"instance ssh user"`
	SshPort string `cli-short:"p" cli-flag:"ssh-port" cli-usage:"instance ssh port"`

	// AI job options
	JobName string `cli-flag:"job" cli-usage:"job name of the AI job to show"`
}

func (c *labAIJobShowCmd) cmdAliases() []string { return gShowAlias }

func (c *labAIJobShowCmd) cmdShort() string { return "Show an AI job status" }

func (c *labAIJobShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an AI job status.
Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *labAIJobShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *labAIJobShowCmd) cmdRun(_ *cobra.Command, _ []string) error { //nolint:gocyclo
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	i, err := instances.FindListInstancesResponseInstances(c.AIJobInstance)
	if err != nil {
		return err
	}

	instance, err := client.GetInstance(ctx, i.ID)
	if err != nil {
		return err
	}

	jobStatusCommand = fmt.Sprintf(jobStatusCommand, defaultAIJobName)

	// Connect via the SSH tunnel and issue the command to check the job
	cmdResponse, err := exossh.RunCmd(instance.PublicIP, c.SshUser, c.SshPort, c.SshKey, jobStatusCommand)
	if err != nil {
		return err
	}

	out := labAIJobShowOutput{
		Instance:    instance.Name,
		InstanceID:  instance.ID.String(),
		AIJobStatus: cmdResponse,
	}

	return c.outputFunc(&out, nil)

}

func init() {
	cobra.CheckErr(registerCLICommand(labAIJobCmd, &labAIJobShowCmd{
		JobName:            defaultAIJobName,
		SshPort:            "22",
		SshUser:            "debian",
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
