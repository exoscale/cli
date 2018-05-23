package cmd

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

type configFile struct {
	APIURL    string
	APIKey    string
	SecretKey string
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate config file for this cli",
}

func configCmdRun(cmd *cobra.Command, args []string) {
	isPrint, err := configCmd.Flags().GetBool("print")
	if err != nil {
		log.Fatal(err)
	}

	_, err = generateConfigFile(isPrint)
	if err != nil {
		log.Fatal(err)
	}
}

func generateConfigFile(isPrint bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	confFile := &configFile{}

	apiURL, err := getConfig(reader, "Compute API Endpoint", "https://api.exoscale.ch/compute")
	if err != nil {
		return "", err
	}
	confFile.APIURL = apiURL

	apiKey, err := getConfig(reader, "API Key", "")
	if err != nil {
		return "", err
	}
	confFile.APIKey = apiKey

	secretKey, err := getConfig(reader, "Secret Key", "")
	if err != nil {
		return "", err
	}
	confFile.SecretKey = secretKey

	tmpl, err := template.New("config").Parse(
		`[cloudstack]
endpoint={{.APIURL}}
key={{.APIKey}}
secret={{.SecretKey}}
`)
	if err != nil {
		return "", err
	}

	var outPut *os.File
	filepath := ""

	if isPrint {
		outPut = os.Stdout
	} else {

		filepath = path.Join(configFolder, "cloudstack.ini")

		_, err := os.Stat(filepath)

		if os.IsNotExist(err) {
			file, err := os.Create(filepath)
			if err != nil {
				return "", err
			}
			outPut = file
			defer file.Close()
		}
	}

	if err = tmpl.Execute(outPut, confFile); err != nil {
		return "", err
	}
	return filepath, nil
}

func getConfig(reader *bufio.Reader, text, def string) (string, error) {
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

func init() {

	configCmd.Run = configCmdRun
	configCmd.Flags().BoolP("print", "p", false, "Print configuration")
	rootCmd.AddCommand(configCmd)
}
