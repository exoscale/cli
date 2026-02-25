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
				// Handle cancellation gracefully
				if errors.Is(err, context.Canceled) {
					fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
					os.Exit(exocmd.ExitCodeInterrupt)
				}
				if err == io.EOF {
					fmt.Fprintln(os.Stderr, "")
					os.Exit(0)
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
					if errors.Is(err, promptui.ErrInterrupt) {
						fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
						os.Exit(exocmd.ExitCodeInterrupt)
					}
					if err == promptui.ErrEOF {
						fmt.Fprintln(os.Stderr, "")
						os.Exit(0)
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
		// Handle cancellation gracefully
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
			os.Exit(exocmd.ExitCodeInterrupt)
		}
		if err == io.EOF {
			fmt.Fprintln(os.Stderr, "")
			os.Exit(0)
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
		// Handle prompt cancellation
		if err == promptui.ErrInterrupt {
			fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
			os.Exit(exocmd.ExitCodeInterrupt)
		}
		if err == promptui.ErrEOF {
			fmt.Fprintln(os.Stderr, "")
			os.Exit(0)
		}
		// API error - try with fallback zones
		for {
			defaultZone, err := chooseZone(globalstate.EgoscaleV3Client, utils.AllZones)
			if err != nil {
				// Handle prompt cancellation in fallback
				if err == promptui.ErrInterrupt {
					fmt.Fprintln(os.Stderr, "Error: Operation Cancelled")
					os.Exit(exocmd.ExitCodeInterrupt)
				}
				if err == promptui.ErrEOF {
					fmt.Fprintln(os.Stderr, "")
					os.Exit(0)
				}
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

// askSetDefault asks whether the new account should become the default.
// Returns true for "y/Y/yes", false for anything else (empty = default No).
//
// This deliberately uses plain bufio line-based I/O rather than promptui.Prompt
// (readline). promptui.Prompt defers its initial PTY render until the first
// keystroke, which deadlocks the settle-based PTY test harness: the test waits
// for output before sending input, but the prompt waits for input before
// producing output. Plain bufio avoids readline entirely; the PTY line
// discipline (cooked mode, restored by the preceding promptui.Select) handles
// '\r' → '\n' translation for us.
func askSetDefault(name string) (bool, error) {
	fmt.Printf("[?] Set [%s] as default account? [y/N]: ", name)

	ctx := exocmd.GContext
	reader := bufio.NewReader(os.Stdin)

	resultCh := make(chan struct {
		value string
		err   error
	}, 1)
	go func() {
		value, err := reader.ReadString('\n')
		resultCh <- struct {
			value string
			err   error
		}{value, err}
	}()

	select {
	case result := <-resultCh:
		if result.err == io.EOF {
			return false, io.EOF
		}
		if result.err != nil {
			return false, result.err
		}
		lower := strings.ToLower(strings.TrimSpace(result.value))
		return lower == "y" || lower == "yes", nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}
