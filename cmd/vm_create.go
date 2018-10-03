package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/exoscale/cli/utils"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// vmCreateCmd represents the create command
var vmCreateCmd = &cobra.Command{
	Use:     "create <vm name>+",
	Short:   "Create and deploy a virtual machine",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		userDataPath, err := cmd.Flags().GetString("cloud-init-file")
		if err != nil {
			return err
		}

		userData := ""

		if userDataPath != "" {
			userData, err = getUserData(userDataPath)
			if err != nil {
				return err
			}
		}

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zoneName == "" {
			zoneName = gCurrentAccount.DefaultZone
		}

		zone, err := getZoneIDByName(zoneName)
		if err != nil {
			return err
		}

		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		if templateName == "" {
			templateName = gCurrentAccount.DefaultTemplate
		}

		diskSize, err := cmd.Flags().GetInt64("disk")
		if err != nil {
			return err
		}

		template, err := getTemplateByName(zone, templateName)
		if err != nil {
			return err
		}

		keypair, err := cmd.Flags().GetString("keypair")
		if err != nil {
			return err
		}

		sg, err := cmd.Flags().GetString("security-group")
		if err != nil {
			return err
		}

		sgs, err := getSecurityGroups(sg)
		if err != nil {
			return err
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		privnet, err := cmd.Flags().GetString("privnet")
		if err != nil {
			return err
		}

		pvs, err := getPrivnetList(privnet, zone)
		if err != nil {
			return err
		}

		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		servOffering, err := getServiceOfferingByName(so)
		if err != nil {
			return err
		}

		affinitygroup, err := cmd.Flags().GetString("anti-affinity-group")
		if err != nil {
			return err
		}

		affinitygroups, err := getAffinityGroup(affinitygroup)
		if err != nil {
			return err
		}

		tasks := make([]egoscale.DeployVirtualMachine, len(args))

		for i, name := range args {

			vmInfo := &egoscale.DeployVirtualMachine{
				Name:              name,
				UserData:          userData,
				ZoneID:            zone,
				TemplateID:        template.ID,
				RootDiskSize:      diskSize,
				KeyPair:           keypair,
				SecurityGroupIDs:  sgs,
				IP6:               &ipv6,
				NetworkIDs:        pvs,
				ServiceOfferingID: servOffering.ID,
				AffinityGroupIDs:  affinitygroups,
			}

			tasks[i] = *vmInfo

		}

		r, errs := createVM(tasks)
		if len(errs) > 0 {
			return errs[0]
		}

		if len(r) > 1 {
			return nil
		}

		if r[0].ID == nil {
			return fmt.Errorf("virtual machine ID is (nil)")
		}
		sshinfo, err := getSSHInfo(r[0].ID.String(), ipv6)
		if err != nil {
			return err
		}

		fmt.Printf(`
What to do now?

1. Connect to the machine

> exo ssh %s
`, r[0].Name)

		printSSHConnectSTR(sshinfo)

		fmt.Printf(`
2. Put the SSH configuration into ".ssh/config"

> exo ssh %s --info
`, r[0].Name)

		printSSHInfo(sshinfo)

		fmt.Print(`
Tip of the day:
	You're the sole owner of the private key.
	Be cautious with it.
`)

		return nil
	},
}

func getCommaflag(p string) []string {
	if p == "" {
		return nil
	}

	p = strings.Trim(p, ",")
	args := strings.Split(p, ",")

	res := []string{}

	for _, arg := range args {
		if arg == "" {
			continue
		}
		res = append(res, strings.TrimSpace(arg))
	}

	return res
}

func getSecurityGroups(commaParameter string) ([]egoscale.UUID, error) {

	sgs := getCommaflag(commaParameter)
	ids := make([]egoscale.UUID, len(sgs))

	for i, sg := range sgs {
		s, err := getSecurityGroupByNameOrID(sg)
		if err != nil {
			return nil, err
		}

		ids[i] = *s.ID
	}

	return ids, nil
}

func getPrivnetList(commaParameter string, zoneID *egoscale.UUID) ([]egoscale.UUID, error) {

	sgs := getCommaflag(commaParameter)
	ids := make([]egoscale.UUID, len(sgs))

	for i, sg := range sgs {
		n, err := getNetworkByNameAndZone(sg, zoneID)
		if err != nil {
			return nil, err
		}

		ids[i] = *n.ID
	}

	return ids, nil
}

func getAffinityGroup(commaParameter string) ([]egoscale.UUID, error) {
	affs := getCommaflag(commaParameter)
	ids := make([]egoscale.UUID, len(affs))

	for i, aff := range affs {
		s, err := getAffinityGroupByName(aff)

		if err != nil {
			return nil, err
		}

		ids[i] = *s.ID
	}

	return ids, nil
}

func getUserData(userDataPath string) (string, error) {
	buff, err := ioutil.ReadFile(userDataPath)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buff), nil
}

func createVM(deploys []egoscale.DeployVirtualMachine) ([]egoscale.VirtualMachine, []error) {
	isDefaultKeyPair := false
	var keyPairs *egoscale.SSHKeyPair

	var keypairsName string
	if deploys[0].KeyPair == "" {
		isDefaultKeyPair = true
		fmt.Println("Creating private SSH key")
		sshKeyName, err := utils.RandStringBytes(64)
		if err != nil {
			return nil, []error{err}
		}
		keyPairs, err = createSSHKey(sshKeyName)
		if err != nil {
			r := err.(*egoscale.ErrorResponse)
			if r.ErrorCode != egoscale.ParamError && r.CSErrorCode != egoscale.InvalidParameterValueException {
				return nil, []error{err}
			}
			return nil, []error{fmt.Errorf("an SSH key with that name %q already exists, please choose a different name", sshKeyName)}
		}
		defer deleteSSHKey(keyPairs.Name) // nolint: errcheck

		keypairsName = keyPairs.Name

	}

	tasks := make([]task, len(deploys))

	for i := range deploys {
		tasks[i].string = fmt.Sprintf("Deploying %q", deploys[i].Name)
		if keypairsName != "" {
			deploys[i].KeyPair = keypairsName
		}
		tasks[i].AsyncCommand = deploys[i]
	}

	resps := asyncTasks(tasks)
	errors := filterErrors(resps)
	if len(errors) > 0 {
		return nil, errors
	}

	vmResp := make([]egoscale.VirtualMachine, len(resps))

	for i, vm := range resps {
		v := vm.resp.(*egoscale.VirtualMachine)
		vmResp[i] = *v
		if isDefaultKeyPair {
			saveKeyPair(keyPairs, *v.ID)
		}
	}
	return vmResp, nil
}

func saveKeyPair(keyPairs *egoscale.SSHKeyPair, vmID egoscale.UUID) {
	filePath := path.Join(gConfigFolder, "instances", vmID.String())

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	filePath = path.Join(filePath, "id_rsa")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filePath, []byte(keyPairs.PrivateKey), 0600); err != nil {
			log.Fatalf("SSH private key could not be written. %s", err)
		}
	}
}

func init() {
	vmCreateCmd.Flags().StringP("cloud-init-file", "f", "", "Deploy instance with a cloud-init file")
	vmCreateCmd.Flags().StringP("zone", "z", "", "<zone name | id | keyword> (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1)")
	vmCreateCmd.Flags().StringP("template", "t", "", fmt.Sprintf("<template name | id> (default: %s)", defaultTemplate))
	vmCreateCmd.Flags().Int64P("disk", "d", 50, "<disk size>")
	vmCreateCmd.Flags().StringP("keypair", "k", "", "<ssh keys name>")
	vmCreateCmd.Flags().StringP("security-group", "s", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().StringP("privnet", "p", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().StringP("anti-affinity-group", "a", "", "<name | id, name | id, ...>")
	vmCreateCmd.Flags().BoolP("ipv6", "6", false, "enable ipv6")
	vmCreateCmd.Flags().StringP("service-offering", "o", "Small", "<name | id> (micro|tiny|small|medium|large|extra-large|huge|mega|titan")
	vmCmd.AddCommand(vmCreateCmd)
}
