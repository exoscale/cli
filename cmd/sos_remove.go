package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove <bucket name> [object name]+",
	Short:   "Remove object(s) from a bucket",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		bucket := args[0]

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		if err = validateArgs(args); err != nil {
			return err
		}

		objects := []string{}
		if len(args) < 2 {
			if !recursive {
				// Cannot remove bucket objects when not invoked recursively
				return cmd.Usage()
			}
			objects = append(objects, "")
		} else {
			objects = args[1:]
		}

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objectsCh := make(chan string)

		// Send object names that are needed to be removed to objectsCh
		go func() {
			defer close(objectsCh)
			// List all objects from a bucket-name with a matching prefix.
			errors := make([]error, 0, len(objects))

			for _, keyPrefix := range objects {
				nbFile := 0
				for object := range sosClient.ListObjects(bucket, keyPrefix, true, gContext.Done()) {
					if object.Err != nil {
						log.Fatalln(object.Err)
					}

					obj := filepath.ToSlash(object.Key)
					keyPrefix = filepath.ToSlash(keyPrefix)
					keyPrefix = strings.Trim(keyPrefix, "/")

					if strings.HasPrefix(obj, "/") {
						keyPrefix = fmt.Sprintf("/%s", keyPrefix)
					}

					if (strings.HasPrefix(obj, fmt.Sprintf("%s/", keyPrefix)) && obj != keyPrefix) || keyPrefix == "" {
						if !recursive {
							errors = append(errors, fmt.Errorf("%s: is a directory", keyPrefix)) // nolint: errcheck
							nbFile = 1
							break
						}
						objectsCh <- obj
					} else if obj == keyPrefix {
						objectsCh <- obj
					}
					nbFile++
				}
				if nbFile == 0 {
					errors = append(errors, fmt.Errorf("cannot remove '%s': No such object or directory", keyPrefix))
				}
				nbFile = 0
			}
			if len(errors) > 0 {
				for _, err := range errors {
					fmt.Fprintf(os.Stderr, "%v\n", err) // nolint: errcheck
				}
				os.Exit(1)
			}
		}()

		for objectErr := range sosClient.RemoveObjectsWithContext(gContext, bucket, objectsCh) {
			return fmt.Errorf("error detected during deletion: %v", objectErr)
		}

		return nil
	},
}

func validateArgs(args []string) error {
	for _, arg := range args {
		if arg == "" {
			return fmt.Errorf("invalid arg: must be not empty")
		}
	}
	return nil
}

func init() {
	sosCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolP("recursive", "r", false, "Attempt to remove the file hierarchy rooted in each file argument")
}
