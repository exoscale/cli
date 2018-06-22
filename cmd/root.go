package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/client"

	"github.com/spf13/cobra"
)

var region string
var configFolder string
var configFilePath string
var cfgFilePath string

var ignoreClientBuild = false

var cs *egoscale.Client

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "exo",
	Short: "A simple CLI to use CloudStack using egoscale lib",
	//Long:  `A simple CLI to use CloudStack using egoscale lib`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", "", "Specify an alternate config file [env CLOUDSTACK_CONFIG]")
	RootCmd.PersistentFlags().StringVarP(&region, "region", "r", "cloudstack", "config ini file section name [env CLOUDSTACK_REGION]")

	cobra.OnInitialize(initConfig, buildClient)

}

func buildClient() {
	if ignoreClientBuild {
		return
	}

	if cs != nil {
		return
	}

	var err error
	cs, err = client.BuildClient(configFilePath, region)
	if err != nil {
		log.Fatal(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	envs := map[string]string{
		"CLOUDSTACK_CONFIG": "config",
		"CLOUDSTACK_REGION": "region",
	}

	for env, flag := range envs {
		flag := RootCmd.Flags().Lookup(flag)
		if value := os.Getenv(env); value != "" {
			flag.Value.Set(value)
		}
	}

	envEndpoint := os.Getenv("CLOUDSTACK_ENDPOINT")
	envKey := os.Getenv("CLOUDSTACK_KEY")
	envSecret := os.Getenv("CLOUDSTACK_SECRET")

	if envEndpoint != "" && envKey != "" && envSecret != "" {
		cs = egoscale.NewClient(envEndpoint, envKey, envSecret)
		return
	}

	usr, _ := user.Current()
	configFolder = path.Join(usr.HomeDir, ".exoscale")

	localConfig, _ := filepath.Abs("cloudstack.ini")
	inis := []string{
		localConfig,
		filepath.Join(usr.HomeDir, ".cloudstack.ini"),
		filepath.Join(configFolder, "cloudstack.ini"),
	}

	for _, i := range inis {
		if _, err := os.Stat(i); err != nil {
			continue
		}
		configFilePath = i
		break
	}

	if cfgFilePath != "" {
		configFilePath = cfgFilePath
		return
	}

	if getCmdPosition("config") == 1 {
		ignoreClientBuild = true
		return
	}

	if configFilePath == "" {
		log.Fatalf("Config file not found within: %s", strings.Join(inis, ", "))
	}
}

// return a command position by fetching os.args and ignoring flags
//
//example: "$ exo -r preprod vm create" vm position is 1 and create is 2
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
