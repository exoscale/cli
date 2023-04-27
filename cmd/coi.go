package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const coiTemplateName = "Container-Optimized Instance"

var coiCmd = &cobra.Command{
	Use:   "coi NAME",
	Short: "Deploy a Container-Optimized Instance",
	Long: fmt.Sprintf(`This command creates a Compute instance running one or several Docker
containers based on the Docker image specified or a Docker Compose
configuration provided. It can be invoked in 2 ways:

* Specifying a Docker image and optionally ports to publish:

    $ exo lab coi nginx \
        --zone ch-gva-2 \
        --image nginxdemos/hello \
        --port 80:80 \
        ...

* Specifying a Docker Compose file:

    $ exo lab coi my-app \
        --zone ch-gva-2 \
        --docker-compose /path/to/docker-compose.yml \
        # or --docker-compose https://my.app/docker-compose.yml
        ...

Once created, the Compute instance can be managed like any other standard
instance using the "exo vm" commands.

The Docker Compose configuration reference can be found at this address:
https://docs.docker.com/compose/compose-file/

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vmShowOutput{}), ", ")),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var cloudInitUserdata string

		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return err
		}

		diskSize, err := cmd.Flags().GetInt64("disk-size")
		if err != nil {
			return err
		}

		sshKey, err := cmd.Flags().GetString("ssh-key")
		if err != nil {
			return err
		}
		if sshKey == "" {
			sshKey = account.CurrentAccount.DefaultSSHKey
		}

		sg, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}

		securityGroups, err := getSecurityGroupIDs(sg)
		if err != nil {
			return err
		}

		privnet, err := cmd.Flags().GetStringSlice("private-network")
		if err != nil {
			return err
		}

		privnets, err := getPrivnetIDs(privnet, zone.ID)
		if err != nil {
			return err
		}

		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		serviceOffering, err := getServiceOfferingByNameOrID(so)
		if err != nil {
			return err
		}

		template, err := getTemplateByNameOrID(zone.ID, coiTemplateName, "community")
		if err != nil {
			return err
		}

		dockerImage, err := cmd.Flags().GetString("image")
		if err != nil {
			return err
		}

		dockerPorts, err := cmd.Flags().GetStringSlice("port")
		if err != nil {
			return err
		}

		dockerComposeSource, err := cmd.Flags().GetString("docker-compose")
		if err != nil {
			return err
		}

		if (dockerImage == "" && dockerComposeSource == "") ||
			(dockerImage != "" && dockerComposeSource != "") {
			return errors.New("either --image or --docker-compose must be specified")
		}

		if dockerImage != "" {
			cloudInitUserdata, err = coiGenerateUserdataWithImage(dockerImage, dockerPorts)
		} else {
			cloudInitUserdata, err = coiGenerateUserdataWithDockerCompose(dockerComposeSource)
		}
		if err != nil {
			return fmt.Errorf("unable to generate cloud-init userdata: %v", err)
		}

		cloudInitUserdataEncoded := base64.StdEncoding.EncodeToString([]byte(cloudInitUserdata))
		if len(cloudInitUserdataEncoded) >= maxUserDataLength {
			return fmt.Errorf("maximum allowed length for Docker Compose configuration is %d bytes",
				maxUserDataLength)
		}

		resp := asyncTasks([]task{{
			egoscale.DeployVirtualMachine{
				Name:              args[0],
				ZoneID:            zone.ID,
				ServiceOfferingID: serviceOffering.ID,
				TemplateID:        template.ID,
				KeyPair:           sshKey,
				RootDiskSize:      diskSize,
				SecurityGroupIDs:  securityGroups,
				NetworkIDs:        privnets,
				UserData:          cloudInitUserdataEncoded,
			},
			fmt.Sprintf("Creating Compute instance %q", args[0]),
		}})
		errors := filterErrors(resp)
		if len(errors) > 0 {
			return errors[0]
		}
		vm := resp[0].resp.(*egoscale.VirtualMachine)

		if !globalstate.Quiet {
			return printOutput(showVM(vm.Name))
		}

		return nil
	},
}

func coiGenerateUserdataWithImage(image string, ports []string) (string, error) {
	dockerCompose, err := yaml.Marshal(map[string]interface{}{
		"version": "3",
		"services": map[string]interface{}{
			"coi": map[string]interface{}{
				"image": image,
				"ports": ports,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`#cloud-config
write_files:
- path: /etc/docker-compose.yml
  content: %s
  encoding: b64
  owner: root:root
  permissions: '0600'
`, base64.StdEncoding.EncodeToString(dockerCompose)), nil
}

func coiGenerateUserdataWithDockerCompose(source string) (string, error) {
	if strings.HasPrefix(source, "http") {
		return fmt.Sprintf(`#cloud-config
bootcmd:
- curl -sL "%s" > /etc/docker-compose.yml
- chmod 600 /etc/docker-compose.yml
`, source), nil
	}

	file, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`#cloud-config
write_files:
- path: /etc/docker-compose.yml
  content: %s
  encoding: b64
  owner: root:root
  permissions: '0600'
`, base64.StdEncoding.EncodeToString(data)), nil
}

func init() {
	coiCmd.Flags().Int64P("disk-size", "d", 20, "disk size")
	coiCmd.Flags().StringP("image", "i", "", "Docker image to run")
	coiCmd.Flags().StringP("docker-compose", "c", "",
		"Docker Compose configuration file (local path/URL)")
	coiCmd.Flags().StringSliceP("port", "p", nil,
		"publish container port(s) (format \"[HOST-PORT:]CONTAINER-PORT\")")
	coiCmd.Flags().StringP("ssh-key", "k", "", "SSH key name")
	coiCmd.Flags().StringSlice("private-network", nil, "Private Network name/ID")
	coiCmd.Flags().StringSlice("security-group", nil, "Security Group name/ID")
	coiCmd.Flags().StringP("service-offering", "o", "medium", serviceOfferingHelp)
	coiCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneHelp)
	labCmd.AddCommand(coiCmd)
}
