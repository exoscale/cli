package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/go-ini/ini"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	exoConfigFileName = "exoscale"
	computeEndpoint   = "https://api.exoscale.ch/compute"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config file for this cli",
}

func configCmdRun(cmd *cobra.Command, args []string) {

	if viper.ConfigFileUsed() != "" {
		println("Good day! exo is already configured with accounts:")
		listAccounts()
		if err := addNewAccount(false); err != nil {
			log.Fatal(err)
		}
		return
	}
	csPath, ok := isCloudstackINIFileExist()
	if ok {
		resp, ok, err := askCloudstackINIMigration(csPath)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			if err := addNewAccount(true); err != nil {
				log.Fatal(err)
			}
			return
		}

		cfgPath, err := createConfigFile(exoConfigFileName)
		if err != nil {
			log.Fatal(err)
		}
		if err := importCloudstackINI(resp, csPath, cfgPath); err != nil {
			log.Fatal(err)
		}
		addNewAccount(false)
		return
	}
	println("Hi happy Exoscalian, some configuration is required to use exo")
	println(`
We now need some very important informations, find them there.
https://portal.exoscale.com/account/profile/api
`)
	addNewAccount(true)
}

func addNewAccount(firstRun bool) error {

	config := &config{}

	if firstRun {
		filePath, err := createConfigFile(exoConfigFileName)
		if err != nil {
			return err
		}

		viper.SetConfigFile(filePath)

		newAccount, err := getAccount()
		if err != nil {
			return err
		}
		config.DefaultAccount = newAccount.Name
		config.Accounts = []account{*newAccount}
		viper.Set("defaultAccount", newAccount.Name)
	}

	for askQuestion("Do you wish to add another account?") {
		newAccount, err := getAccount()
		if err != nil {
			return err
		}
		config.Accounts = append(config.Accounts, *newAccount)
		if askQuestion("Make [" + newAccount.Name + "] your default profile?") {
			config.DefaultAccount = newAccount.Name
			viper.Set("defaultAccount", newAccount.Name)
		}
	}

	if len(config.Accounts) == 0 {
		return nil
	}

	return addAccount(viper.ConfigFileUsed(), config)
}

func getAccount() (*account, error) {
	reader := bufio.NewReader(os.Stdin)

	account := &account{}

	name, err := readInput(reader, "Account name", "")
	if err != nil {
		return nil, err
	}

	for name == "" {
		fmt.Printf("Must be not empty\n")
		name, err = readInput(reader, "Account name", "")
		if err != nil {
			return nil, err
		}
	}

	for isAccountExist(name) {
		fmt.Printf("Account name [%s] already exist\n", name)
		name, err = readInput(reader, "Account name", "")
		if err != nil {
			return nil, err
		}
	}

	account.Name = name

	account.Endpoint = computeEndpoint

	apiKey, err := readInput(reader, "API Key", "")
	if err != nil {
		return nil, err
	}
	account.Key = apiKey

	secretKey, err := readInput(reader, "Secret Key", "")
	if err != nil {
		return nil, err
	}
	account.Secret = secretKey

	accountResp, err := checkCredentials(account)
	if err != nil {
		return nil, fmt.Errorf("Account [%s]: unable to verify user credentials", account.Name)
	}

	account.Account = accountResp.Name

	cs := egoscale.NewClient(account.Endpoint, account.Key, account.Secret)

	defaultZone, err := chooseZone(account.Name, cs)
	if err != nil {
		return nil, err
	}

	account.DefaultZone = defaultZone

	return account, nil
}

func addAccount(filePath string, newAccounts *config) error {

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

	accounts := make([]map[string]string, accountsSize+newAccountsSize)

	conf := &config{}

	for i, acc := range currentAccounts {

		accounts[i] = map[string]string{}

		accounts[i]["name"] = acc.Name
		accounts[i]["endpoint"] = acc.Endpoint
		accounts[i]["key"] = acc.Key
		accounts[i]["secret"] = acc.Secret
		accounts[i]["defaultZone"] = acc.DefaultZone
		accounts[i]["account"] = acc.Account

		conf.Accounts = append(conf.Accounts, acc)
	}

	if newAccounts != nil {

		for i, acc := range newAccounts.Accounts {

			accounts[accountsSize+i] = map[string]string{}

			accounts[accountsSize+i]["name"] = acc.Name
			accounts[accountsSize+i]["endpoint"] = acc.Endpoint
			accounts[accountsSize+i]["key"] = acc.Key
			accounts[accountsSize+i]["secret"] = acc.Secret
			accounts[accountsSize+i]["defaultZone"] = acc.DefaultZone
			accounts[accountsSize+i]["account"] = acc.Account
			conf.Accounts = append(conf.Accounts, acc)
		}
	}

	viper.SetConfigType("toml")
	viper.SetConfigFile(filePath)

	viper.Set("accounts", accounts)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	conf.DefaultAccount = viper.Get("defaultAccount").(string)
	gAllAccount = conf

	return nil

}

func isCloudstackINIFileExist() (string, bool) {

	envConfigPath := os.Getenv("CLOUDSTACK_CONFIG")

	usr, _ := user.Current()

	localConfig, _ := filepath.Abs("cloudstack.ini")
	inis := []string{
		localConfig,
		filepath.Join(usr.HomeDir, ".cloudstack.ini"),
		filepath.Join(gConfigFolder, "cloudstack.ini"),
		envConfigPath,
	}

	cfgPath := ""

	for _, i := range inis {
		if _, err := os.Stat(i); err != nil {
			continue
		}
		cfgPath = i
		break
	}

	if cfgPath == "" {
		return "", false
	}
	return cfgPath, true
}

func askCloudstackINIMigration(csFilePath string) (string, bool, error) {

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, csFilePath)
	if err != nil {
		return "", false, err
	}

	if len(cfg.Sections()) <= 0 {
		return "", false, nil
	}

	fmt.Printf("We've found a %q configuration file with the following configurations:\n", "cloudstack.ini")
	for i, acc := range cfg.Sections() {
		if i == 0 {
			continue
		}
		fmt.Printf("- [%s] %s\n", acc.Name(), acc.Key("key").String())
	}

	reader := bufio.NewReader(os.Stdin)

	resp, err := readInput(reader, "Do you wish to import them automagically?", "All, some, none")
	if err != nil {
		return "", false, err
	}

	resp = strings.ToLower(resp)

	return resp, (resp == "all" || resp == "some"), nil
}

func importCloudstackINI(option, csPath, cfgPath string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, csPath)
	if err != nil {
		return err
	}

	config := &config{}

	for i, acc := range cfg.Sections() {
		if i == 0 {
			continue
		}

		if option == "some" {
			if !askQuestion(fmt.Sprintf("Importing %s %s?", acc.Name(), acc.Key("key").String())) {
				continue
			}
		}

		account := account{
			Name:     acc.Name(),
			Endpoint: acc.Key("endpoint").String(),
			Key:      acc.Key("key").String(),
			Secret:   acc.Key("secret").String(),
		}

		accountResp, err := checkCredentials(&account)
		if err != nil {
			fmt.Printf("Account [%s]: unable to verify user credentials\n", acc.Name())
			if !askQuestion("Do you want to keep this account?") {
				continue
			}
		}

		cs := egoscale.NewClient(account.Endpoint, account.Key, account.Secret)

		defaultZone, err := chooseZone(acc.Name(), cs)
		if err != nil {
			return err
		}

		account.DefaultZone = defaultZone

		isDefault := false
		if askQuestion("Make [" + acc.Name() + "] your default profile?") {
			isDefault = true
		}

		account.Account = accountResp.Name

		config.Accounts = append(config.Accounts, account)

		if i == 1 || isDefault {
			config.DefaultAccount = acc.Name()
			viper.Set("defaultAccount", acc.Name())
		}
	}

	addAccount(cfgPath, config)

	return nil
}

func isAccountExist(name string) bool {

	if gAllAccount == nil {
		return false
	}

	for _, acc := range gAllAccount.Accounts {
		if acc.Name == name {
			return true
		}
	}

	return false
}

func createConfigFile(fileName string) (string, error) {
	if _, err := os.Stat(gConfigFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(gConfigFolder, os.ModePerm); err != nil {
			return "", err
		}
	}

	filepath := path.Join(gConfigFolder, fileName+".toml")

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		return "", fmt.Errorf("File %q already exist", filepath)
	}
	return filepath, nil
}

func readInput(reader *bufio.Reader, text, def string) (string, error) {
	if def == "" {
		fmt.Printf("[+] %s [%s]: ", text, "none")
	} else {
		fmt.Printf("[+] %s [%s]: ", text, def)
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)

	if input != "" {
		return input, nil
	}
	return def, nil
}

func askQuestion(text string) bool {

	reader := bufio.NewReader(os.Stdin)

	resp, err := readInput(reader, text, "Yn")
	if err != nil {
		log.Fatal(err)
	}

	return (strings.ToLower(resp) == "y" || strings.ToLower(resp) == "yes")
}

func checkCredentials(account *account) (*egoscale.Account, error) {
	cs := egoscale.NewClient(account.Endpoint, account.Key, account.Secret)

	resp, err := cs.Request(&egoscale.ListAccounts{})
	if err != nil {
		return nil, err
	}

	accountsResp := resp.(*egoscale.ListAccountsResponse)

	if accountsResp.Count == 1 {
		return &accountsResp.Account[0], nil
	}

	return nil, fmt.Errorf("more than one account found")
}

func listAccounts() {
	if gAllAccount == nil {
		return
	}
	for _, acc := range gAllAccount.Accounts {
		print("- ", acc.Name)
		if acc.Name == gAllAccount.DefaultAccount {
			print(" [Default]")
		}
		println("")
	}
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

func getSelectedZone(number string, zones map[string]string) (string, bool) {
	zName, ok := zones[number]
	if !ok {
		return "", false
	}
	return zName, true
}

func chooseZone(accountName string, cs *egoscale.Client) (string, error) {

	reader := bufio.NewReader(os.Stdin)

	zonesResp, err := cs.List(&egoscale.Zone{})
	if err != nil {
		return "", err
	}

	zones := map[string]string{}

	// XXX if no zone is found like in preprod bug
	if len(zonesResp) == 0 {
		println(`No zones found: take "ch-dk-2" by default`)
		return "ch-dk-2", nil
	}

	fmt.Printf("Choose [%s] default zone:\n", accountName)

	for i, z := range zonesResp {
		zone := z.(*egoscale.Zone)

		zName := strings.ToLower(zone.Name)

		n := fmt.Sprintf("%d", i+1)

		zones[n] = zName

		fmt.Printf("%d: %s\n", i+1, zName)
	}

	zoneNumber, err := readInput(reader, "Select", "1")
	if err != nil {
		return "", err
	}

	defaultZone, ok := getSelectedZone(zoneNumber, zones)
	for !ok {
		println("Error: Invalid zone number")
		defaultZone, err = chooseZone(accountName, cs)
		if err == nil {
			break
		}
	}
	return defaultZone, nil
}

func init() {

	configCmd.Run = configCmdRun
	RootCmd.AddCommand(configCmd)
}
