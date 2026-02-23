package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new account to configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if this is the first account
			isFirstAccount := account.GAllAccount == nil || len(account.GAllAccount.Accounts) == 0

			if isFirstAccount {
				printNoConfigMessage()
			}

			newAccount, err := promptAccountInformation()
			if err != nil {
				// Handle cancellation gracefully without showing error
				if errors.Is(err, context.Canceled) || err == io.EOF {
					fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
					os.Exit(130) // Standard exit code for SIGINT
				}
				return err
			}

			config := &account.Config{Accounts: []account.Account{*newAccount}}

			// Get config file path, creating if this is the first account
			filePath := exocmd.GConfig.ConfigFileUsed()
			if isFirstAccount && filePath == "" {
				if filePath, err = createConfigFile(exocmd.DefaultConfigFileName); err != nil {
					return err
				}
				exocmd.GConfig.SetConfigFile(filePath)
			}

			if isFirstAccount {
				// First account: automatically set as default
				config.DefaultAccount = newAccount.Name
				exocmd.GConfig.Set("defaultAccount", newAccount.Name)
				fmt.Printf("Set [%s] as default account (first account)\n", newAccount.Name)
			} else {
				// Additional account: ask user if it should be the new default
				setDefault, err := askSetDefault(newAccount.Name)
				if err != nil {
					if errors.Is(err, promptui.ErrInterrupt) || err == io.EOF {
						fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
						os.Exit(130)
					}
					return err
				}
				if setDefault {
					config.DefaultAccount = newAccount.Name
					exocmd.GConfig.Set("defaultAccount", newAccount.Name)
					fmt.Printf("Set [%s] as default account\n", newAccount.Name)
				}
			}

			return saveConfig(filePath, config)
		},
	})
}

func addConfigAccount(firstRun bool) error {
	var (
		config account.Config
		err    error
	)

	filePath := exocmd.GConfig.ConfigFileUsed()

	if firstRun {
		if filePath, err = createConfigFile(exocmd.DefaultConfigFileName); err != nil {
			return err
		}

		exocmd.GConfig.SetConfigFile(filePath)
	}

	newAccount, err := promptAccountInformation()
	if err != nil {
		// Handle cancellation gracefully with message
		if errors.Is(err, context.Canceled) || err == io.EOF {
			fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
			os.Exit(130) // Standard exit code for SIGINT
		}
		return err
	}
	config.DefaultAccount = newAccount.Name
	config.Accounts = []account.Account{*newAccount}
	exocmd.GConfig.Set("defaultAccount", newAccount.Name)

	if len(config.Accounts) == 0 {
		return nil
	}

	return saveConfig(filePath, &config)
}

// readInputWithContext reads a line from stdin with context cancellation support.
// Returns io.EOF if Ctrl+C or Ctrl+D is pressed, allowing graceful cancellation.
// Silent exit behavior matches promptui.Select's interrupt handling.
func readInputWithContext(ctx context.Context, reader *bufio.Reader, prompt string) (string, error) {
	fmt.Printf("[+] %s: ", prompt)

	inputCh := make(chan struct {
		value string
		err   error
	}, 1)

	go func() {
		value, err := reader.ReadString('\n')
		inputCh <- struct {
			value string
			err   error
		}{value, err}
	}()

	select {
	case result := <-inputCh:
		if result.err != nil {
			return "", result.err
		}
		return strings.TrimSpace(result.value), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func promptAccountInformation() (*account.Account, error) {
	var client *v3.Client

	ctx := exocmd.GContext

	reader := bufio.NewReader(os.Stdin)
	account := &account.Account{}

	// Prompt for API Key with validation
	apiKey, err := readInputWithContext(ctx, reader, "API Key")
	if err != nil {
		return nil, err
	}
	for apiKey == "" {
		fmt.Println("API Key cannot be empty")
		apiKey, err = readInputWithContext(ctx, reader, "API Key")
		if err != nil {
			return nil, err
		}
	}
	account.Key = apiKey

	// Prompt for Secret Key with validation
	secretKey, err := readInputWithContext(ctx, reader, "Secret Key")
	if err != nil {
		return nil, err
	}
	for secretKey == "" {
		fmt.Println("Secret Key cannot be empty")
		secretKey, err = readInputWithContext(ctx, reader, "Secret Key")
		if err != nil {
			return nil, err
		}
	}
	account.Secret = secretKey

	// Prompt for Name with validation
	name, err := readInputWithContext(ctx, reader, "Name")
	if err != nil {
		return nil, err
	}
	for name == "" {
		fmt.Println("Name cannot be empty")
		name, err = readInputWithContext(ctx, reader, "Name")
		if err != nil {
			return nil, err
		}
	}
	account.Name = name

	for {
		if a := getAccountByName(account.Name); a == nil {
			break
		}

		fmt.Printf("Name [%s] already exist\n", name)
		name, err = utils.ReadInput(ctx, reader, "Name", account.Name)
		if err != nil {
			return nil, err
		}

		account.Name = name
	}

	client, err = v3.NewClient(credentials.NewStaticCredentials(
		account.Key, account.APISecret(),
	))
	if err != nil {
		return nil, err
	}
	account.DefaultZone, err = chooseZone(client, nil)
	if err != nil {
		for {
			defaultZone, err := chooseZone(globalstate.EgoscaleV3Client, utils.AllZones)
			if err != nil {
				return nil, err
			}
			if defaultZone != "" {
				account.DefaultZone = defaultZone
				break
			}
		}
	}

	return account, nil
}

// askSetDefault uses a promptui Prompt (PTY-compatible) to ask whether the new
// account should become the default. Returns true for "y/Y/yes", false for anything else.
func askSetDefault(name string) (bool, error) {
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("Set [%s] as default account? [y/N]", name),
		Validate: func(input string) error {
			lower := strings.ToLower(strings.TrimSpace(input))
			if lower == "" || lower == "y" || lower == "yes" || lower == "n" || lower == "no" {
				return nil
			}
			return fmt.Errorf("please enter y or n")
		},
	}
	result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	lower := strings.ToLower(strings.TrimSpace(result))
	return lower == "y" || lower == "yes", nil
}
