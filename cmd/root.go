package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

var GContext context.Context

var GConfig *viper.Viper
var GConfigFilePath string

// current Account information
var gAccountName string

// Aliases
var GListAlias = []string{"ls"}
var GRemoveAlias = []string{"rm"}
var GDeleteAlias = []string{"del"}
var GShowAlias = []string{"get"}
var GCreateAlias = []string{"add"}

var RootCmd = &cobra.Command{
	Use:           "exo",
	Short:         "Manage your Exoscale infrastructure easily",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of exo",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Printf("%s %s %s (egoscale %s)\n", cmd.Parent().Name(), globalstate.GitVersion, globalstate.GitCommit, v3.Version)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute(version, commit string) {

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	GContext = ctx

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)

		os.Exit(1) //nolint:gocritic
	}
}

func init() {
	account.CurrentAccount = &account.Account{
		DefaultZone: DefaultZone,
		Environment: DefaultEnvironment,
		SosEndpoint: DefaultSosEndpoint,
	}

	GConfig = viper.New()

	RootCmd.PersistentFlags().StringVarP(&GConfigFilePath, "config", "C", "", "Specify an alternate config file [env EXOSCALE_CONFIG]")
	RootCmd.PersistentFlags().StringVarP(&gAccountName, "use-account", "A", "", "Account to use in config file [env EXOSCALE_ACCOUNT]")
	RootCmd.PersistentFlags().StringVarP(&globalstate.OutputFormat, "output-format", "O", "", "Output format (table|json|text), see \"exo output --help\" for more information")
	RootCmd.PersistentFlags().StringVar(&output.GOutputTemplate, "output-template", "", "Template to use if output format is \"text\"")
	RootCmd.PersistentFlags().BoolVarP(&globalstate.Quiet, "quiet", "Q", false, "Quiet mode (disable non-essential command output)")
	RootCmd.AddCommand(versionCmd)

	// Don't attempt to load client configuration in testing mode.
	// FIXME: stop using global configurations, see if this can be replaced
	//   with rootCmd.PersistentPreRun or something.
	if !strings.HasSuffix(os.Args[0], ".test") {
		cobra.OnInitialize(initConfig, buildClient)
	}
}

var ignoreClientBuild = false

// initConfig reads in config file and ENV variables if set.
func initConfig() { //nolint:gocyclo
	envs := map[string]string{
		"EXOSCALE_CONFIG":  "config",
		"EXOSCALE_ACCOUNT": "use-account",
	}

	for env, flag := range envs {
		pflag := RootCmd.Flags().Lookup(flag)
		if pflag == nil {
			panic(fmt.Sprintf("unknown flag '%s'", flag))
		}

		if value, ok := os.LookupEnv(env); ok {
			if err := pflag.Value.Set(value); err != nil {
				log.Fatal(err)
			}
		}
	}

	sosEndpointFromEnv := readFromEnv(
		"EXOSCALE_STORAGE_API_ENDPOINT",
		"EXOSCALE_SOS_ENDPOINT",
	)

	apiEndpoint := os.Getenv("EXOSCALE_API_ENDPOINT")

	apiKeyFromEnv := readFromEnv(
		"EXOSCALE_API_KEY",
		"EXOSCALE_KEY",
		"CLOUDSTACK_KEY",
		"CLOUDSTACK_API_KEY",
	)

	apiSecretFromEnv := readFromEnv(
		"EXOSCALE_API_SECRET",
		"EXOSCALE_SECRET",
		"EXOSCALE_SECRET_KEY",
		"CLOUDSTACK_SECRET",
		"CLOUDSTACK_SECRET_KEY",
	)

	apiEnvironmentFromEnv := readFromEnv("EXOSCALE_API_ENVIRONMENT")

	if apiKeyFromEnv != "" && apiSecretFromEnv != "" {
		account.CurrentAccount.Name = "<environment variables>"
		GConfigFilePath = "<environment variables>"
		account.CurrentAccount.Key = apiKeyFromEnv
		account.CurrentAccount.Secret = apiSecretFromEnv

		if apiEndpoint != "" {
			account.CurrentAccount.Endpoint = apiEndpoint
		}

		if apiEnvironmentFromEnv != "" {
			account.CurrentAccount.Environment = apiEnvironmentFromEnv
		}
		if sosEndpointFromEnv != "" {
			account.CurrentAccount.SosEndpoint = sosEndpointFromEnv
		}

		clientTimeoutFromEnv := readFromEnv("EXOSCALE_API_TIMEOUT")
		if clientTimeoutFromEnv != "" {
			if t, err := strconv.Atoi(clientTimeoutFromEnv); err == nil {
				account.CurrentAccount.ClientTimeout = t
			}
		}

		account.GAllAccount = &account.Config{
			DefaultAccount: account.CurrentAccount.Name,
			Accounts:       []account.Account{*account.CurrentAccount},
		}

		buildClient()

		return
	}

	config := &account.Config{}

	usr, err := user.Current()
	if err != nil {
		log.Println(`current user cannot be read, using "root"`)
		usr = &user.User{
			Uid:      "0",
			Gid:      "0",
			Username: "root",
			Name:     "root",
			HomeDir:  "/root",
		}
	}

	cfgdir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("could not find configuration directory: %s", err)
	}
	globalstate.ConfigFolder = path.Join(cfgdir, "exoscale")

	// Snap packages use $HOME/.exoscale (as negotiated with the snap store)
	if _, snap := os.LookupEnv("SNAP_USER_COMMON"); snap {
		globalstate.ConfigFolder = path.Join(usr.HomeDir, ".exoscale")
	}

	if GConfigFilePath != "" {
		configFileStat, err := os.Stat(GConfigFilePath)
		if err != nil {
			log.Fatal(err)
		}

		if configFileStat.IsDir() {
			log.Fatalf("%q is a directory but but should be configuration file", GConfigFilePath)
		}

		// Use config file from the flag.
		GConfig.SetConfigFile(GConfigFilePath)
	} else {
		GConfig.SetConfigName("exoscale")
		GConfig.SetConfigType("toml")
		GConfig.AddConfigPath(globalstate.ConfigFolder)
		// Retain backwards compatibility
		GConfig.AddConfigPath(path.Join(usr.HomeDir, ".exoscale"))
		GConfig.AddConfigPath(usr.HomeDir)
		GConfig.AddConfigPath(".")
	}

	nonCredentialCmds := []string{"config", "version", "status"}

	if err := GConfig.ReadInConfig(); err != nil {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			return
		}

		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal(`error: the exo CLI must be configured before usage, please run "exo config"`)
		}

		log.Fatal(err)
	}

	// All the stored data (e.g. ssh keys) will be put next to the config file.
	GConfigFilePath = GConfig.ConfigFileUsed()
	globalstate.ConfigFolder = filepath.Dir(GConfigFilePath)

	if err := GConfig.Unmarshal(config); err != nil {
		log.Fatal(fmt.Errorf("couldn't read config: %s", err))
	}

	if len(config.Accounts) == 0 {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			// Set GAllAccount so config commands can handle the empty state gracefully
			account.GAllAccount = config
			return
		}

		log.Fatalf("no accounts were found into %q", GConfig.ConfigFileUsed())
		return
	}

	// Allow config management commands to run without a default account
	// This fixes the circular dependency where 'exo config set' couldn't run
	// to set a default account because it required a default account to exist
	configManagementCmds := []string{"list", "set", "show"}
	isConfigManagementCmd := getCmdPosition("config") == 1
	if isConfigManagementCmd && len(os.Args) > 2 {
		// Check if the subcommand is a config management command
		// Need to find the actual subcommand by skipping flags
		for i := 2; i < len(os.Args); i++ {
			if !strings.HasPrefix(os.Args[i], "-") {
				isConfigManagementCmd = contains(configManagementCmds, os.Args[i])
				break
			}
		}
	} else {
		isConfigManagementCmd = false
	}

	if config.DefaultAccount == "" && gAccountName == "" {
		// Allow config management commands to proceed without default account
		if isConfigManagementCmd {
			ignoreClientBuild = true
			// Set GAllAccount so config commands can access the account list
			account.GAllAccount = config
			return
		}

		// Provide helpful error message with available accounts
		var availableAccounts []string
		for _, acc := range config.Accounts {
			availableAccounts = append(availableAccounts, acc.Name)
		}
		if len(availableAccounts) > 0 {
			log.Fatalf("default account not defined\n\nSet a default account with: exo config set <account-name>\nAvailable accounts: %s\n\nOr specify an account for this command with: --use-account <account-name>",
				strings.Join(availableAccounts, ", "))
		} else {
			log.Fatalf("default account not defined")
		}
	}

	if gAccountName == "" {
		gAccountName = config.DefaultAccount
	}

	account.GAllAccount = config
	account.GAllAccount.DefaultAccount = gAccountName

	for i, acc := range config.Accounts {
		if acc.Name == gAccountName {
			account.CurrentAccount = &config.Accounts[i]
			break
		}
	}

	if account.CurrentAccount.Name == "" {
		log.Fatalf("error: could't find any configured account named %q", gAccountName)
	}

	if account.CurrentAccount.Environment == "" {
		account.CurrentAccount.Environment = DefaultEnvironment
	}

	if account.CurrentAccount.DefaultZone == "" {
		account.CurrentAccount.DefaultZone = DefaultZone
	}

	// if an output format isn't specified via cli argument, use
	// the current account default format
	if globalstate.OutputFormat == "" {
		if account.CurrentAccount.DefaultOutputFormat != "" {
			globalstate.OutputFormat = account.CurrentAccount.DefaultOutputFormat
		} else {
			globalstate.OutputFormat = DefaultOutputFormat
		}
	}

	if account.CurrentAccount.SosEndpoint == "" {
		account.CurrentAccount.SosEndpoint = DefaultSosEndpoint
	}

	clientTimeoutFromEnv := readFromEnv("EXOSCALE_API_TIMEOUT")
	if clientTimeoutFromEnv != "" {
		if t, err := strconv.Atoi(clientTimeoutFromEnv); err == nil {
			account.CurrentAccount.ClientTimeout = t
		}
	}

	account.CurrentAccount.SosEndpoint = strings.TrimRight(account.CurrentAccount.SosEndpoint, "/")
}

func isNonCredentialCmd(cmds ...string) bool {
	for _, cmd := range cmds {
		if getCmdPosition(cmd) == 1 {
			return true
		}
	}
	return false
}

// getCmdPosition returns a command position by fetching os.args and ignoring flags
//
// example: "$ exo -r preprod vm create" vm position is 1 and create is 2
func getCmdPosition(cmd string) int {
	count := 1

	isFlagParam := false

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			trimmedArg := strings.Trim(arg, "-")

			flag := RootCmd.Flags().Lookup(trimmedArg)
			if flag == nil && len(trimmedArg) < 2 {
				flag = RootCmd.Flags().ShorthandLookup(trimmedArg)
			}

			if flag != nil && (flag.Value.Type() != "bool") {
				isFlagParam = true
			}
			continue
		}

		if isFlagParam {
			isFlagParam = false
			continue
		}

		if arg == cmd {
			break
		}
		count++
	}

	return count
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// readFromEnv is a os.Getenv on steroids
func readFromEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
	}
	return ""
}
