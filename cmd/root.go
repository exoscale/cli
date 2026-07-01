package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

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
var GUpdateAlias = []string{"set"}

var RootCmd = &cobra.Command{
	Use:           "exo",
	Short:         "Manage your Exoscale infrastructure easily",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if globalstate.RequestTimeout != -time.Second && globalstate.RequestTimeout <= 0 {
			return fmt.Errorf("--timeout must be a positive duration (e.g. 15s), or -1s to disable")
		}
		return nil
	},
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

	// Trap Ctrl+C (and SIGHUP, which the kernel delivers when the PTY session
	// leader exits before we do) and cancel the context.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP)
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
		fmt.Fprintf(os.Stderr, "error: %s\n", formatError(err))

		os.Exit(1) //nolint:gocritic
	}
}

var jsonPathFieldRe = regexp.MustCompile(`\$\['([^']+)'\]`)

// formatError renders a user-friendly message from an error. When the chain
// contains a *v3.APIError it formats the structured server response; otherwise
// it falls back to the raw error string.
func formatError(err error) string {
	var apiErr *v3.APIError
	if !errors.As(err, &apiErr) {
		return err.Error()
	}

	// Prefer the specific "detail" message, fall back to "title".
	lead := apiErr.Detail
	if lead == "" {
		lead = apiErr.Title
	}
	if lead == "" {
		// Simple message/error format — just the status text and message.
		if apiErr.Message != "" {
			return apiErr.Unwrap().Error() + ": " + apiErr.Message
		}
		return apiErr.Unwrap().Error()
	}

	msg := apiErr.Unwrap().Error() + ": " + apiErr.Title + ": " + lead
	for _, e := range apiErr.Errors {
		field := formatFieldName(e.Location)
		detail := formatDetail(e.Detail)
		if field != "" {
			msg += fmt.Sprintf("\n  - %s: %s", field, detail)
		} else {
			msg += fmt.Sprintf("\n  - %s", detail)
		}
	}
	return msg
}

// formatFieldName turns a JSONPath location like $['inference-engine-version']
// into a CLI flag name like --inference-engine-version.
func formatFieldName(location string) string {
	m := jsonPathFieldRe.FindStringSubmatch(location)
	if len(m) < 2 {
		return location
	}
	return "--" + m[1]
}

// formatDetail rewrites technical validation messages into plain English.
func formatDetail(detail string) string {
	const enumPrefix = "does not have a value in the enumeration "
	if after, ok := strings.CutPrefix(detail, enumPrefix); ok {
		after = strings.TrimSpace(after)
		var values []string
		if err := json.Unmarshal([]byte(after), &values); err == nil && len(values) > 0 {
			return "invalid value; valid values are: " + strings.Join(values, ", ")
		}
	}
	return detail
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
	RootCmd.PersistentFlags().DurationVar(&globalstate.RequestTimeout, "timeout", 15*time.Second, "Per-zone timeout for list operations; -1s disables timeout [env EXOSCALE_TIMEOUT]")
	RootCmd.AddCommand(versionCmd)

	// Don't attempt to load client configuration in testing mode.
	// FIXME: stop using global configurations, see if this can be replaced
	//   with rootCmd.PersistentPreRun or something.
	if !strings.HasSuffix(os.Args[0], ".test") {
		cobra.OnInitialize(initConfig, buildClient)
	}
}

var ignoreClientBuild = false

// applyOutputFormat sets the output format from the account default when
// --output-format was not passed on the command line.
func applyOutputFormat() {
	if globalstate.OutputFormat == "" {
		if account.CurrentAccount.DefaultOutputFormat != "" {
			globalstate.OutputFormat = account.CurrentAccount.DefaultOutputFormat
		} else {
			globalstate.OutputFormat = DefaultOutputFormat
		}
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() { //nolint:gocyclo
	// Bind meta-config env vars to flags; CLI flags take precedence.
	metaEnvFlags := map[string]string{
		"EXOSCALE_CONFIG":  "config",
		"EXOSCALE_ACCOUNT": "use-account",
		"EXOSCALE_TIMEOUT": "timeout",
	}

	for envVar, flagName := range metaEnvFlags {
		pflag := RootCmd.Flags().Lookup(flagName)
		if pflag == nil {
			panic(fmt.Sprintf("unknown flag %q", flagName))
		}

		if value, ok := os.LookupEnv(envVar); ok && !pflag.Changed {
			if err := pflag.Value.Set(value); err != nil {
				log.Fatal(err)
			}
		}
	}

	env := readEnvSources()

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

	file, err := loadFileSources(GConfig, gAccountName)
	if err != nil {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			account.GAllAccount = &account.Config{}
			return
		}
		log.Fatal(err)
	}

	// No config file: require env credentials or bail.
	if file.config == nil {
		if env.hasCredentials() {
			seed := account.Account{Name: "<environment variables>"}
			GConfigFilePath = "<environment variables>"
			resolved := resolve(env, fileSources{profile: &seed})
			account.GAllAccount = &account.Config{
				DefaultAccount: resolved.Name,
				Accounts:       []account.Account{resolved},
			}
			account.CurrentAccount = &account.GAllAccount.Accounts[0]
			applyOutputFormat()
			return
		}
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			account.GAllAccount = &account.Config{}
			return
		}
		log.Fatal(`error: the exo CLI must be configured before usage, please run "exo config"`)
	}

	// Config file found: update paths.
	GConfigFilePath = GConfig.ConfigFileUsed()
	globalstate.ConfigFolder = filepath.Dir(GConfigFilePath)

	if len(file.config.Accounts) == 0 {
		if isNonCredentialCmd(nonCredentialCmds...) {
			ignoreClientBuild = true
			account.GAllAccount = file.config
			return
		}
		log.Fatalf("no accounts were found into %q", GConfig.ConfigFileUsed())
	}

	// Allow config management commands to run without a default account.
	configManagementCmds := []string{"list", "set", "show"}
	isConfigManagementCmd := getCmdPosition("config") == 1
	if isConfigManagementCmd && len(os.Args) > 2 {
		for i := 2; i < len(os.Args); i++ {
			if !strings.HasPrefix(os.Args[i], "-") {
				isConfigManagementCmd = slices.Contains(configManagementCmds, os.Args[i])
				break
			}
		}
	} else {
		isConfigManagementCmd = false
	}

	if file.config.DefaultAccount == "" && gAccountName == "" {
		if isConfigManagementCmd {
			ignoreClientBuild = true
			account.GAllAccount = file.config
			return
		}

		var names []string
		for _, acc := range file.config.Accounts {
			names = append(names, acc.Name)
		}
		if len(names) > 0 {
			log.Fatalf("default account not defined\n\nSet a default account with: exo config set <account-name>\nAvailable accounts: %s\n\nOr specify an account for this command with: --use-account <account-name>",
				strings.Join(names, ", "))
		} else {
			log.Fatalf("default account not defined")
		}
	}

	if file.profile == nil {
		selectedName := gAccountName
		if selectedName == "" {
			selectedName = file.config.DefaultAccount
		}
		log.Fatalf("error: could't find any configured account named %q", selectedName)
	}

	if gAccountName == "" {
		gAccountName = file.config.DefaultAccount
	}

	account.GAllAccount = file.config
	account.GAllAccount.DefaultAccount = gAccountName

	resolved := resolve(env, file)
	// Update in place so account.CurrentAccount pointer stays valid.
	for i := range account.GAllAccount.Accounts {
		if account.GAllAccount.Accounts[i].Name == resolved.Name {
			account.GAllAccount.Accounts[i] = resolved
			account.CurrentAccount = &account.GAllAccount.Accounts[i]
			break
		}
	}
	applyOutputFormat()
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

// readFromEnv is a os.Getenv on steroids
func readFromEnv(keys ...string) string {
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
	}
	return ""
}
