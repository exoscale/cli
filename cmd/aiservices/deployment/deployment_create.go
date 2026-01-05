package deployment

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type DeploymentCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"?" cli-usage:"NAME"`
	GPUType  string `cli-flag:"gpu-type" cli-usage:"GPU type family (e.g., gpua5000, gpu3080ti)"`
	GPUCount int64  `cli-flag:"gpu-count" cli-usage:"Number of GPUs (1-8)"`
	Replicas int64  `cli-flag:"replicas" cli-usage:"Number of replicas (>=1)"`

	ModelID                   string      `cli-flag:"model-id" cli-usage:"Model ID (UUID)"`
	ModelName                 string      `cli-flag:"model-name" cli-usage:"Model name (as created)"`
	InferenceEngineParameters string      `cli-flag:"inference-engine-params" cli-usage:"Space-separated inference engine server CLI arguments (e.g., \"--gpu-memory-usage=0.8 --max-tokens=4096\")"`
	InferenceEngineHelp       bool        `cli-flag:"inference-engine-parameter-help" cli-usage:"Show inference engine parameters help"`
	Zone                      v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *DeploymentCreateCmd) CmdShort() string     { return "Create AI deployment" }
func (c *DeploymentCreateCmd) CmdLong() string {
	return "This command creates an AI deployment on dedicated inference servers."
}
func (c *DeploymentCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *DeploymentCreateCmd) showInferenceEngineParameterHelp(ctx context.Context, client *v3.Client) error {
	resp, err := client.GetInferenceEngineHelp(ctx)
	if err != nil {
		return err
	}

	sections := make(map[string][]v3.InferenceEngineParameterEntry)
	var sectionNames []string
	for _, p := range resp.Parameters {
		if _, ok := sections[p.Section]; !ok {
			sectionNames = append(sectionNames, p.Section)
		}
		sections[p.Section] = append(sections[p.Section], p)
	}
	sort.Strings(sectionNames)

	for i, section := range sectionNames {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s:\n", section)
		for _, p := range sections[section] {
			flags := strings.Join(p.Flags, ", ")
			if p.Type != "boolean" && p.Type != "enum" {
				flags += " " + strings.ToUpper(strings.ReplaceAll(p.Name, "-", "_"))
			}

			fmt.Printf("  %s\n", flags)

			desc := p.Description
			if p.Default != "" {
				// The description in example sometimes already includes default, but the example output
				// shows (default: ...) at the end.
				if !strings.Contains(desc, fmt.Sprintf("(default: %s)", p.Default)) {
					desc += fmt.Sprintf(" (default: %s)", p.Default)
				}
			}

			// Simple wrapping - adapt to terminal width
			maxWidth := 80 // default fallback
			if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
				maxWidth = width
			}

			words := strings.Fields(desc)
			if len(words) > 0 {
				line := "                        "
				for _, word := range words {
					if len(line)+len(word) > maxWidth {
						fmt.Println(line)
						line = "                        " + word
					} else {
						if line == "                        " {
							line += word
						} else {
							line += " " + word
						}
					}
				}
				fmt.Println(line)
			}
		}
	}

	return nil
}

func (c *DeploymentCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if c.InferenceEngineHelp {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
		if err != nil {
			return err
		}
		return c.showInferenceEngineParameterHelp(ctx, client)
	}

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("NAME is required")
	}

	if c.GPUType == "" || c.GPUCount == 0 {
		return fmt.Errorf("--gpu-type and --gpu-count are required")
	}
	if c.ModelID == "" && c.ModelName == "" {
		return fmt.Errorf("--model-id or --model-name is required")
	}

	// Parse inference engine parameters from space-separated string
	var inferenceParams []string
	if c.InferenceEngineParameters != "" {
		inferenceParams = strings.Fields(c.InferenceEngineParameters)
	}

	req := v3.CreateDeploymentRequest{
		Name:                      c.Name,
		GpuType:                   c.GPUType,
		GpuCount:                  c.GPUCount,
		Replicas:                  c.Replicas,
		InferenceEngineParameters: inferenceParams,
	}
	if c.ModelID != "" || c.ModelName != "" {
		req.Model = &v3.ModelRef{}
		if c.ModelID != "" {
			if id, err := v3.ParseUUID(c.ModelID); err == nil {
				req.Model.ID = id
			} else {
				return fmt.Errorf("invalid --model-id: %w", err)
			}
		}
		if c.ModelName != "" {
			req.Model.Name = c.ModelName
		}
	}

	if err := utils.RunAsync(ctx, client, fmt.Sprintf("Creating deployment %q...", c.Name), func(ctx context.Context, c *v3.Client) (*v3.Operation, error) {
		return c.CreateDeployment(ctx, req)
	}); err != nil {
		return err
	}
	if !globalstate.Quiet {
		fmt.Fprintln(os.Stdout, "Deployment created.")
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
