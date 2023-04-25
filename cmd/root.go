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

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gContext context.Context

var gConfig *viper.Viper
var gConfigFolder string
var gConfigFilePath string

// current Account information
var gAccountName string
var gCurrentAccount = &account.Account{
	DefaultZone:     defaultZone,
	DefaultTemplate: defaultTemplate,
	Endpoint:        defaultEndpoint,
	Environment:     defaultEnvironment,
	SosEndpoint:     defaultSosEndpoint,
}

var csRunstatus *egoscale.Client

// Aliases
var gListAlias = []string{"ls"}
var gRemoveAlias = []string{"rm"}
var gDeleteAlias = []string{"del"}
var gRevokeAlias = []string{"rvk"}
var gShowAlias = []string{"get"}
var gCreateAlias = []string{"add"}
var gUploadAlias = []string{"up"}
var gDissociateAlias = []string{"disassociate", "dissoc"}
var gAssociateAlias = []string{"assoc"}

var RootCmd = &cobra.Command{
	Use:           "exo",
	Short:         "Manage your Exoscale infrastructure easily",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var gVersion string
var gCommit string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of exo",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Printf("%s %s %s (egoscale %s)\n", cmd.Parent().Name(), gVersion, gCommit, egoscale.Version)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute(version, commit string) {
	gVersion = version
	gCommit = commit

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

	gContext = ctx

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	gConfig = viper.New()

	RootCmd.PersistentFlags().StringVarP(&gConfigFilePath, "config", "C", "", "Specify an alternate config file [env EXOSCALE_CONFIG]")
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
func initConfig() {
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

	endpointFromEnv := readFromEnv(
		"EXOSCALE_API_ENDPOINT",
		"EXOSCALE_COMPUTE_API_ENDPOINT",
		"EXOSCALE_ENDPOINT",
		"EXOSCALE_COMPUTE_ENDPOINT",
		"CLOUDSTACK_ENDPOINT")

	sosEndpointFromEnv := readFromEnv(
		"EXOSCALE_STORAGE_API_ENDPOINT",
		"EXOSCALE_SOS_ENDPOINT",
	)

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
		gCurrentAccount.Name = "<environment variables>"
		gConfigFilePath = "<environment variables>"
		gCurrentAccount.Account = "unknown"
		gCurrentAccount.Key = apiKeyFromEnv
		gCurrentAccount.Secret = apiSecretFromEnv

		if apiEnvironmentFromEnv != "" {
			gCurrentAccount.Environment = apiEnvironmentFromEnv
		}
		if endpointFromEnv != "" {
			gCurrentAccount.Endpoint = endpointFromEnv
		}
		if sosEndpointFromEnv != "" {
			gCurrentAccount.SosEndpoint = sosEndpointFromEnv
		}
		gCurrentAccount.DNSEndpoint = buildDNSAPIEndpoint(gCurrentAccount.Endpoint)

		if gCurrentAccount.ClientTimeout == 0 {
			gCurrentAccount.ClientTimeout = defaultClientTimeout
		}

		account.GAllAccount = &account.AccountConfig{
			DefaultAccount: gCurrentAccount.Name,
			Accounts:       []account.Account{*gCurrentAccount},
		}

		buildClient()

		return
	}

	config := &account.AccountConfig{}

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
	gConfigFolder = path.Join(cfgdir, "exoscale")

	// Snap packages use $HOME/.exoscale (as negotiated with the snap store)
	if _, snap := os.LookupEnv("SNAP_USER_COMMON"); snap {
		gConfigFolder = path.Join(usr.HomeDir, ".exoscale")
	}

	if gConfigFilePath != "" {
		// Use config file from the flag.
		gConfig.SetConfigFile(gConfigFilePath)
	} else {
		gConfig.SetConfigName("exoscale")
		gConfig.AddConfigPath(gConfigFolder)
		// Retain backwards compatibility
		gConfig.AddConfigPath(path.Join(usr.HomeDir, ".exoscale"))
		gConfig.AddConfigPath(usr.HomeDir)
		gConfig.AddConfigPath(".")
	}

	nonCredentialCmds := []string{"config", "version", "status"}

	if err := gConfig.ReadInConfig(); err != nil {
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
	gConfigFilePath = gConfig.ConfigFileUsed()
	gConfigFolder = filepath.Dir(gConfigFilePath)

	if err := gConfig.Unmarshal(config); err != nil {
		log.Fatal(fmt.Errorf("couldn't read config: %s", err))
	}

	if len(config.Accounts) == 0 {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			return
		}

		log.Fatalf("no accounts were found into %q", gConfig.ConfigFileUsed())
		return
	}

	if config.DefaultAccount == "" && gAccountName == "" {
		log.Fatalf("default account not defined")
	}

	if gAccountName == "" {
		gAccountName = config.DefaultAccount
	}

	account.GAllAccount = config
	account.GAllAccount.DefaultAccount = gAccountName

	for i, acc := range config.Accounts {
		if acc.Name == gAccountName {
			gCurrentAccount = &config.Accounts[i]
			break
		}
	}

	if gCurrentAccount.Name == "" {
		log.Fatalf("error: could't find any configured account named %q", gAccountName)
	}

	if gCurrentAccount.Endpoint == "" {
		if gCurrentAccount.ComputeEndpoint != "" {
			gCurrentAccount.Endpoint = gCurrentAccount.ComputeEndpoint
		} else {
			gCurrentAccount.Endpoint = defaultEndpoint
		}
	}

	if gCurrentAccount.Environment == "" {
		gCurrentAccount.Environment = defaultEnvironment
	}

	if gCurrentAccount.DefaultZone == "" {
		gCurrentAccount.DefaultZone = defaultZone
	}

	// if an output format isn't specified via cli argument, use
	// the current account default format
	if globalstate.OutputFormat == "" {
		if gCurrentAccount.DefaultOutputFormat != "" {
			globalstate.OutputFormat = gCurrentAccount.DefaultOutputFormat
		} else {
			globalstate.OutputFormat = defaultOutputFormat
		}
	}

	if gCurrentAccount.DNSEndpoint == "" {
		gCurrentAccount.DNSEndpoint = buildDNSAPIEndpoint(gCurrentAccount.Endpoint)
	}

	if gCurrentAccount.DefaultTemplate == "" {
		gCurrentAccount.DefaultTemplate = defaultTemplate
	}

	if gCurrentAccount.SosEndpoint == "" {
		gCurrentAccount.SosEndpoint = defaultSosEndpoint
	}

	if gCurrentAccount.RunstatusEndpoint == "" {
		gCurrentAccount.RunstatusEndpoint = defaultRunstatusEndpoint
	}

	if gCurrentAccount.ClientTimeout == 0 {
		gCurrentAccount.ClientTimeout = defaultClientTimeout
	}
	clientTimeoutFromEnv := readFromEnv("EXOSCALE_API_TIMEOUT")
	if clientTimeoutFromEnv != "" {
		if t, err := strconv.Atoi(clientTimeoutFromEnv); err == nil {
			gCurrentAccount.ClientTimeout = t
		}
	}

	gCurrentAccount.Endpoint = strings.TrimRight(gCurrentAccount.Endpoint, "/")
	gCurrentAccount.DNSEndpoint = strings.TrimRight(gCurrentAccount.DNSEndpoint, "/")
	gCurrentAccount.SosEndpoint = strings.TrimRight(gCurrentAccount.SosEndpoint, "/")
	gCurrentAccount.RunstatusEndpoint = strings.TrimRight(gCurrentAccount.RunstatusEndpoint, "/")
}

func isNonCredentialCmd(cmds ...string) bool {
	for _, cmd := range cmds {
		if getCmdPosition(cmd) == 1 {
			return true
		}
	}
	return false
}

func buildDNSAPIEndpoint(defaultEndpoint string) string {
	dnsEndpoint := strings.Replace(defaultEndpoint, "/"+apiVersion, "/dns", 1)
	if strings.Contains(dnsEndpoint, "/"+legacyAPIVersion) {
		dnsEndpoint = strings.Replace(defaultEndpoint, "/"+legacyAPIVersion, "/dns", 1)
	}

	return dnsEndpoint
}

// getCmdPosition returns a command position by fetching os.args and ignoring flags
//
// example: "$ exo -r preprod vm create" vm position is 1 and create is 2
func getCmdPosition(cmd string) int {
	count := 1

	isFlagParam := false

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			flag := RootCmd.Flags().Lookup(strings.Trim(arg, "-"))
			if flag == nil {
				flag = RootCmd.Flags().ShorthandLookup(strings.Trim(arg, "-"))
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

// readFromEnv is a os.Getenv on steroids
func readFromEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
	}
	return ""
}
