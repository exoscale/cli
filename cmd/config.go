package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type config struct {
	DefaultAccount      string
	DefaultOutputFormat string
	Accounts            []account
}

type account struct {
	Name                 string
	Account              string
	Endpoint             string
	ComputeEndpoint      string // legacy config.
	DNSEndpoint          string
	SosEndpoint          string
	RunstatusEndpoint    string
	Environment          string
	Key                  string
	Secret               string
	SecretCommand        []string
	DefaultZone          string
	DefaultSSHKey        string
	DefaultTemplate      string
	DefaultRunstatusPage string
	CustomHeaders        map[string]string
}

func (a account) APISecret() string {
	if len(a.SecretCommand) != 0 {
		cmd := exec.Command(a.SecretCommand[0], a.SecretCommand[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		return strings.TrimRight(string(out), "\n")
	}

	return a.Secret
}

func (a account) AccountName() string {
	if a.Name == "" {
		resp, err := cs.GetWithContext(gContext, egoscale.Account{})
		if err != nil {
			log.Fatal(err)
		}
		acc := resp.(*egoscale.Account)
		return acc.Name
	}

	return a.Name
}

func (a account) IsDefault() bool {
	return a.Name == gAllAccount.DefaultAccount
}

const (
	legacyAPIVersion         = "compute"
	apiVersion               = "v1"
	defaultEndpoint          = "https://api.exoscale.com/" + apiVersion
	defaultEnvironment       = "api"
	defaultConfigFileName    = "exoscale"
	defaultTemplate          = "Linux Ubuntu 20.04 LTS 64-bit"
	defaultSosEndpoint       = "https://sos-{zone}.exo.io"
	defaultRunstatusEndpoint = "https://api.runstatus.com"
	defaultZone              = "ch-dk-2"
	defaultOutputFormat      = "table"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Exoscale CLI configuration management",
	RunE:  configCmdRun,
}

func configCmdRun(cmd *cobra.Command, _ []string) error {
	var (
		defaultAccountMark = promptui.Styler(promptui.FGYellow)("*")
		newAccountLabel    = "<Configure a new account>"
	)

	if gConfigFilePath == "" && gCurrentAccount.Key != "" {
		log.Fatalf("remove ENV credentials variables to use %s", cmd.CalledAs())
	}

	if gConfigFilePath != "" && gCurrentAccount.Key != "" {
		accounts := listAccounts(defaultAccountMark)
		accounts = append(accounts, newAccountLabel)
		prompt := promptui.Select{
			Label: fmt.Sprintf("Configured accounts (%s = default account)", defaultAccountMark),
			Items: accounts,
			Size:  len(accounts),
		}
		_, selectedAccount, err := prompt.Run()
		if err != nil {
			switch err {
			case promptui.ErrInterrupt:
				return nil
			default:
				return fmt.Errorf("prompt failed: %s", err)
			}
		}

		if selectedAccount == newAccountLabel {
			return addConfigAccount(false)
		}

		if strings.TrimSuffix(selectedAccount, defaultAccountMark) != gAllAccount.DefaultAccount {
			fmt.Printf("Setting default account to [%s]\n", selectedAccount)
			gConfig.Set("defaultAccount", selectedAccount)
			return saveConfig(gConfig.ConfigFileUsed(), nil)
		}

		return nil
	}

	fmt.Println("No Exoscale CLI configuration found")

	fmt.Print(`
In order to set up your configuration profile, you will need to retrieve
Exoscale API credentials from your organization's IAM:

    https://portal.exoscale.com/iam/api-keys

`)
	return addConfigAccount(true)
}

func saveConfig(filePath string, newAccounts *config) error {
	accountsSize := 0
	currentAccounts := []account{}
	if gAllAccount != nil {
		accountsSize = len(gAllAccount.Accounts)
		currentAccounts = gAllAccount.Accounts
	}

	newAccountsSize := 0

	if newAccounts != nil {
		newAccountsSize = len(newAccounts.Accounts)
	}

	accounts := make([]map[string]interface{}, accountsSize+newAccountsSize)

	conf := &config{}

	for i, acc := range currentAccounts {
		accounts[i] = map[string]interface{}{}

		accounts[i]["name"] = acc.Name
		accounts[i]["endpoint"] = acc.Endpoint
		accounts[i]["key"] = acc.Key
		accounts[i]["defaultZone"] = acc.DefaultZone
		accounts[i]["environment"] = acc.Environment
		if acc.DefaultSSHKey != "" {
			accounts[i]["defaultSSHKey"] = acc.DefaultSSHKey
		}
		if acc.DefaultRunstatusPage != "" {
			accounts[i]["defaultRunstatusPage"] = acc.DefaultRunstatusPage
		}
		if acc.DefaultTemplate != "" {
			accounts[i]["defaultTemplate"] = acc.DefaultTemplate
		}
		if len(acc.SecretCommand) != 0 {
			accounts[i]["secretCommand"] = acc.SecretCommand
		} else {
			accounts[i]["secret"] = acc.Secret
		}
		accounts[i]["account"] = acc.Account

		conf.Accounts = append(conf.Accounts, acc)
	}

	if newAccounts != nil {
		for i, acc := range newAccounts.Accounts {
			accounts[accountsSize+i] = map[string]interface{}{}

			accounts[accountsSize+i]["name"] = acc.Name
			accounts[accountsSize+i]["endpoint"] = acc.Endpoint
			accounts[accountsSize+i]["key"] = acc.Key
			accounts[accountsSize+i]["secret"] = acc.Secret
			accounts[accountsSize+i]["defaultZone"] = acc.DefaultZone
			accounts[accountsSize+i]["environment"] = acc.Environment
			if acc.DefaultSSHKey != "" {
				accounts[accountsSize+i]["defaultSSHKey"] = acc.DefaultSSHKey
			}
			if acc.DefaultRunstatusPage != "" {
				accounts[accountsSize+i]["defaultRunstatusPage"] = acc.DefaultRunstatusPage
			}
			accounts[accountsSize+i]["account"] = acc.Account
			conf.Accounts = append(conf.Accounts, acc)
		}
	}

	gConfig.SetConfigType("toml")
	gConfig.SetConfigFile(filePath)

	gConfig.Set("accounts", accounts)

	if err := gConfig.WriteConfig(); err != nil {
		return err
	}

	conf.DefaultAccount = gConfig.Get("defaultAccount").(string)
	gAllAccount = conf

	return nil
}

func createConfigFile(fileName string) (string, error) {
	if _, err := os.Stat(gConfigFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(gConfigFolder, os.ModePerm); err != nil {
			return "", err
		}
	}

	filepath := path.Join(gConfigFolder, fileName+".toml")

	if gConfig.ConfigFileUsed() == "" {
		if _, err := os.Stat(filepath); !os.IsNotExist(err) {
			return "", fmt.Errorf("%q exists already", filepath)
		}
	}

	fp, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer fp.Close() // nolint: errcheck

	return filepath, nil
}

func readInput(reader *bufio.Reader, text, def string) (string, error) {
	if def == "" {
		fmt.Printf("[+] %s [%s]: ", text, "none")
	} else {
		fmt.Printf("[+] %s [%s]: ", text, def)
	}
	c := make(chan bool)
	defer close(c)

	input := ""
	var err error
	go func() {
		input, err = reader.ReadString('\n')
		c <- true
	}()

	select {
	case <-c:
	case <-gContext.Done():
		err = fmt.Errorf("")
	}

	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}
	return input, nil
}

func askQuestion(text string) bool {
	reader := bufio.NewReader(os.Stdin)

	resp, err := readInput(reader, text, "yN")
	if err != nil {
		log.Fatal(err)
	}

	return (strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes")
}

func listAccounts(defaultAccountMark string) []string {
	if gAllAccount == nil {
		return nil
	}
	res := make([]string, len(gAllAccount.Accounts))
	for i, acc := range gAllAccount.Accounts {
		res[i] = acc.Name
		if acc.Name == gAllAccount.DefaultAccount {
			res[i] = fmt.Sprintf("%s%s", res[i], defaultAccountMark)
		}
	}
	return res
}

func getAccountByName(name string) *account {
	if gAllAccount == nil {
		return nil
	}

	for i, acc := range gAllAccount.Accounts {
		if acc.Name == name {
			return &gAllAccount.Accounts[i]
		}
	}

	return nil
}

func chooseZone(cs *egoscale.Client, zones []string) (string, error) {
	if zones == nil {
		zonesResp, err := cs.ListWithContext(gContext, &egoscale.Zone{})
		if err != nil {
			return "", err
		}

		if len(zonesResp) == 0 {
			return "", fmt.Errorf("no zones were found")
		}

		zones = make([]string, len(zonesResp))
		for i, z := range zonesResp {
			zone := z.(*egoscale.Zone)
			zName := strings.ToLower(zone.Name)
			zones[i] = zName
		}
	}

	prompt := promptui.Select{
		Label: "Default zone",
		Items: zones,
		Size:  len(zones),
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", fmt.Errorf("prompt failed %v", err)
	}

	return result, nil
}

func init() {
	RootCmd.AddCommand(configCmd)
}
