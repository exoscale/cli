package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gContext context.Context

var gConfigFolder string
var gConfigFilePath string

//current Account information
var gAccountName string
var gCurrentAccount = &account{
	DefaultZone:     defaultZone,
	DefaultTemplate: defaultTemplate,
	Endpoint:        defaultEndpoint,
	SosEndpoint:     defaultSosEndpoint,
}

var gAllAccount *config

//egoscale client
var cs *egoscale.Client
var csDNS *egoscale.Client
var csRunstatus *egoscale.Client

//Aliases
var gListAlias = []string{"ls"}
var gRemoveAlias = []string{"rm"}
var gDeleteAlias = []string{"del"}
var gRevokeAlias = []string{"rvk"}
var gShowAlias = []string{"get"}
var gCreateAlias = []string{"add"}
var gUploadAlias = []string{"up"}
var gDissociateAlias = []string{"disassociate", "dissoc"}
var gAssociateAlias = []string{"assoc"}

// RootCmd represents the base command when called without any subcommands
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

var (
	gOutputFormat   string
	gOutputTemplate string

	gQuiet bool
)

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
	RootCmd.PersistentFlags().StringVarP(&gConfigFilePath, "config", "C", "", "Specify an alternate config file [env EXOSCALE_CONFIG]")
	RootCmd.PersistentFlags().StringVarP(&gAccountName, "use-account", "A", "", "Account to use in config file [env EXOSCALE_ACCOUNT]")
	RootCmd.PersistentFlags().StringVarP(&gOutputFormat, "output-format", "O", "", "Output format (table|json|text)")
	RootCmd.PersistentFlags().StringVar(&gOutputTemplate, "output-template", "", "Template to use if output format is \"text\"")
	RootCmd.PersistentFlags().BoolVarP(&gQuiet, "quiet", "Q", false, "Quiet mode (disable non-essential command output")
	RootCmd.AddCommand(versionCmd)

	cobra.OnInitialize(initConfig, buildClient)
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

	// an attempt to mimic existing behaviours

	envEndpoint := readFromEnv(
		"EXOSCALE_ENDPOINT",
		"EXOSCALE_COMPUTE_ENDPOINT",
		"CLOUDSTACK_ENDPOINT")

	envKey := readFromEnv(
		"EXOSCALE_KEY",
		"EXOSCALE_API_KEY",
		"CLOUDSTACK_KEY",
		"CLOUDSTACK_API_KEY",
	)

	envSecret := readFromEnv(
		"EXOSCALE_SECRET",
		"EXOSCALE_API_SECRET",
		"EXOSCALE_SECRET_KEY",
		"CLOUDSTACK_SECRET",
		"CLOUDSTACK_SECRET_KEY",
	)

	envSosEndpoint := readFromEnv(
		"EXOSCALE_SOS_ENDPOINT",
	)

	if envKey != "" && envSecret != "" {
		gCurrentAccount.Name = "environment variables"
		gCurrentAccount.Account = "unknown"
		gCurrentAccount.Key = envKey
		gCurrentAccount.Secret = envSecret

		if envEndpoint != "" {
			gCurrentAccount.Endpoint = envEndpoint
		}
		if envSosEndpoint != "" {
			gCurrentAccount.SosEndpoint = envSosEndpoint
		}
		gCurrentAccount.DNSEndpoint = strings.Replace(gCurrentAccount.Endpoint, "/compute", "/dns", 1)

		gAllAccount = &config{
			DefaultAccount: gCurrentAccount.Name,
			Accounts:       []account{*gCurrentAccount},
		}

		buildClient()

		return
	}

	config := &config{}

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

	xdgHome, found := os.LookupEnv("XDG_CONFIG_HOME")
	if found {
		gConfigFolder = path.Join(xdgHome, "exoscale")
	} else {
		home, found := os.LookupEnv("HOME")
		if found {
			gConfigFolder = path.Join(home, ".config", "exoscale")
		} else {
			// The XDG spec specifies a default XDG_CONFIG_HOME in $HOME/.config
			gConfigFolder = path.Join(usr.HomeDir, ".config", "exoscale")
		}
	}

	// Snap packages use $HOME/.exoscale (as negotiated with the snap store)
	if _, snap := os.LookupEnv("SNAP_USER_COMMON"); snap {
		gConfigFolder = path.Join(usr.HomeDir, ".exoscale")
	}

	if gConfigFilePath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(gConfigFilePath)
	} else {
		viper.SetConfigName("exoscale")
		viper.AddConfigPath(gConfigFolder)
		// Retain backwards compatibility
		viper.AddConfigPath(path.Join(usr.HomeDir, ".exoscale"))
		viper.AddConfigPath(usr.HomeDir)
		viper.AddConfigPath(".")
	}

	nonCredentialCmds := []string{"config", "version", "status"}

	if err := viper.ReadInConfig(); err != nil {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			return
		}

		log.Fatal(err)
	}

	// All the stored data (e.g. ssh keys) will be put next to the config file.
	gConfigFilePath = viper.ConfigFileUsed()
	gConfigFolder = path.Dir(gConfigFilePath)

	if err := viper.Unmarshal(config); err != nil {
		log.Fatal(fmt.Errorf("couldn't read config: %s", err))
	}

	if len(config.Accounts) == 0 {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			return
		}

		log.Fatalf("no accounts were found into %q", viper.ConfigFileUsed())
		return
	}

	if config.DefaultAccount == "" && gAccountName == "" {
		log.Fatalf("default account not defined")
	}

	if gOutputFormat == "" {
		if gOutputFormat = config.DefaultOutputFormat; gOutputFormat == "" {
			gOutputFormat = defaultOutputFormat
		}
	}

	if gAccountName == "" {
		gAccountName = config.DefaultAccount
	}

	gAllAccount = config
	gAllAccount.DefaultAccount = gAccountName

	for i, acc := range config.Accounts {
		if acc.Name == gAccountName {
			gCurrentAccount = &config.Accounts[i]
			break
		}
	}

	if gCurrentAccount == nil {
		log.Fatalf("could't find any account with name: %q", gAccountName)
	}

	if gCurrentAccount.Endpoint == "" {
		if gCurrentAccount.ComputeEndpoint != "" {
			gCurrentAccount.Endpoint = gCurrentAccount.ComputeEndpoint
		} else {
			gCurrentAccount.Endpoint = defaultEndpoint
		}
	}

	if gCurrentAccount.DefaultZone == "" {
		gCurrentAccount.DefaultZone = defaultZone
	}

	if gCurrentAccount.DNSEndpoint == "" {
		gCurrentAccount.DNSEndpoint = strings.Replace(gCurrentAccount.Endpoint, "/compute", "/dns", 1)
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

	gCurrentAccount.Endpoint = strings.TrimRight(gCurrentAccount.Endpoint, "/")
	gCurrentAccount.DNSEndpoint = strings.TrimRight(gCurrentAccount.DNSEndpoint, "/")
	gCurrentAccount.SosEndpoint = strings.TrimRight(gCurrentAccount.SosEndpoint, "/")
	gCurrentAccount.RunstatusEndpoint = strings.TrimRight(gCurrentAccount.RunstatusEndpoint, "/")

	egoscale.UserAgent = fmt.Sprintf("Exoscale-CLI/%s (%s) %s", gVersion, gCommit, egoscale.UserAgent)
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
//
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
