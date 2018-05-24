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
var configFilePath = ""
var cfgFilePath string

var cs *egoscale.Client

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "exo",
	Short: "A simple CLI to use CloudStack using egoscale lib",
	//Long:  `A simple CLI to use CloudStack using egoscale lib`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFilePath, "config", "", "Specify an alternate config file (default: \"~/.cloudstack.ini\" | \"~/.exoscale/cloudstack.ini\")")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "cloudstack", "config ini file section name")

	cobra.OnInitialize(initConfig, buildClient)

}

func buildClient() {
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
	envEndpoint := os.Getenv("CLOUDSTACK_ENDPOINT")
	envKey := os.Getenv("CLOUDSTACK_KEY")
	envSecret := os.Getenv("CLOUDSTACK_SECRET")

	if envEndpoint != "" && envKey != "" && envSecret != "" {
		cs = egoscale.NewClient(envEndpoint, envKey, envSecret)
		return
	}

	envConfigFile := os.Getenv("CLOUDSTACK_CONFIG")

	if envConfigFile != "" {
		configFilePath = envConfigFile
		return
	}

	usr, _ := user.Current()
	configFolder = path.Join(usr.HomeDir, ".exoscale")

	if cfgFilePath != "" {
		configFilePath = cfgFilePath
		return
	}

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

	if configFilePath == "" {
		path, err := generateConfigFile(false)
		if err != nil {
			log.Fatal(err)
		}
		configFilePath = path
	}

	if configFilePath == "" {
		log.Fatalf("Config file not found within: %s", strings.Join(inis, ", "))
	}
}
