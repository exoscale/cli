package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/exoscale/cli/utils"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Exoscale api",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintf(os.Stderr, `/!\  WARNING  /!\  WARNING  /!\ WARNING  /!\ WARNING  /!\  WARNING  /!\
/!\
/!\    The "exo api" command is deprecated and will be removed in a near
/!\    future, please stop using it and prefer CLI command equivalents.
/!\
/!\  WARNING  /!\  WARNING  /!\  WARNING  /!\  WARNING  /!\  WARNING  /!\
`)
		time.Sleep(3 * time.Second)
	},
}

const userDocumentationURL = "http://cloudstack.apache.org/api/apidocs-4.4/user/%s.html"

// global flags
var apiDebug bool
var apiDryRun bool

func init() {
	RootCmd.AddCommand(apiCmd)
	buildCommands(methods)
	apiCmd.PersistentFlags().BoolVarP(&apiDebug, "debug", "d", false, "debug mode on")
	apiCmd.PersistentFlags().BoolVarP(&apiDryRun, "dry-run", "D", false, "produce a cURL ready URL")
	if err := apiCmd.PersistentFlags().MarkHidden("dry-run"); err != nil {
		log.Fatal(err)
	}
}

func buildCommands(methods []category) {
	for _, category := range methods {
		cmd := cobra.Command{
			Use:     category.name,
			Aliases: category.alias,
			Short:   category.doc,
		}

		apiCmd.AddCommand(&cmd)

		for i := range category.cmd {
			s := category.cmd[i]

			realName := cs.APIName(s.command)
			description := cs.APIDescription(s.command)

			url := userDocumentationURL

			name := realName
			if s.name != "" {
				name = s.name
			}

			hiddenCMD := cobra.Command{
				Use:    realName,
				Short:  description,
				Long:   fmt.Sprintf("%s <%s>", description, fmt.Sprintf(url, realName)),
				Hidden: true,
			}

			buildFlags(s.command, &hiddenCMD)

			subCMD := cobra.Command{
				Use:     name,
				Short:   description,
				Long:    fmt.Sprintf("%s <%s>", description, fmt.Sprintf(url, realName)),
				Aliases: append(s.alias, realName),
				Hidden:  s.hidden,
			}

			buildFlags(s.command, &subCMD)

			runCMD := func(cmd *cobra.Command, args []string) error {
				if len(args) > 0 {
					return fmt.Errorf("raw arguments are not supported. Did you mean?\n\n%s --%s", cmd.CommandPath(), strings.Join(args, " --"))
				}

				// Show request and quit DEBUG
				if apiDebug {
					payload, err := cs.Payload(s.command)
					if err != nil {
						log.Fatal(err)
					}
					qs := payload.Encode()
					if _, err = fmt.Fprintf(os.Stdout, "%s\\\n?%s", cs.Endpoint, strings.Replace(qs, "&", "\\\n&", -1)); err != nil {
						log.Fatal(err)
					}

					if _, err := fmt.Fprintln(os.Stdout); err != nil {
						log.Fatal(err)
					}
					os.Exit(0)
				}

				if apiDryRun {
					payload, err := cs.Payload(s.command)
					if err != nil {
						log.Fatal(err)
					}
					signature, err := cs.Sign(payload)
					if err != nil {
						log.Fatal(err)
					}

					payload.Add("signature", signature)

					if _, err := fmt.Fprintf(os.Stdout, "%s?%s\n", cs.Endpoint, payload.Encode()); err != nil {
						log.Fatal(err)
					}
					os.Exit(0)
				}

				// End debug section

				resp, err := cs.RequestWithContext(gContext, s.command)
				if err != nil {
					return err
				}

				data, err := json.MarshalIndent(&resp, "", "  ")
				if err != nil {
					return err
				}

				utils.PrintJSON(string(data), "")

				return nil
			}

			subCMD.RunE = runCMD
			hiddenCMD.RunE = runCMD

			subCMD.Flags().SortFlags = false
			hiddenCMD.Flags().SortFlags = false

			cmd.AddCommand(&subCMD)
			apiCmd.AddCommand(&hiddenCMD)
		}
	}
}

func buildFlags(method egoscale.Command, cmd *cobra.Command) {
	val := reflect.ValueOf(method)
	// we've got a pointer
	value := val.Elem()

	buildFlag(value, cmd)
}

func buildFlag(value reflect.Value, cmd *cobra.Command) {
	if value.Kind() != reflect.Struct {
		log.Fatalf("struct was expected")
		return
	}

	ty := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := ty.Field(i)

		if field.Name == "_" {
			continue
		}

		// XXX refactor with request.go
		var argName string
		required := false
		if json, ok := field.Tag.Lookup("json"); ok {
			tags := strings.Split(json, ",")
			argName = tags[0]
			required = true
			for _, tag := range tags {
				if tag == "omitempty" {
					required = false
				}
			}
			if argName == "" || argName == "omitempty" {
				continue
			}
		}

		description := ""
		if required {
			description = "required"
		}

		if doc, ok := field.Tag.Lookup("doc"); ok {
			if description != "" {
				description = fmt.Sprintf("[%s] %s", description, doc)
			} else {
				description = doc
			}
		}

		val := value.Field(i)
		addr := val.Addr().Interface()
		switch val.Kind() {
		case reflect.Bool:
			cmd.Flags().BoolVarP(addr.(*bool), argName, "", false, description)
		case reflect.Int:
			cmd.Flags().IntVarP(addr.(*int), argName, "", 0, description)
		case reflect.Int64:
			cmd.Flags().Int64VarP(addr.(*int64), argName, "", 0, description)
		case reflect.Uint:
			cmd.Flags().UintVarP(addr.(*uint), argName, "", 0, description)
		case reflect.Uint64:
			cmd.Flags().Uint64VarP(addr.(*uint64), argName, "", 0, description)
		case reflect.Float64:
			cmd.Flags().Float64VarP(addr.(*float64), argName, "", 0, description)
		case reflect.Int16:
			typeName := field.Type.Name()
			if typeName != "int16" {
				cmd.Flags().VarP(&intTypeGeneric{addr: addr, base: 10, bitSize: 16, typ: field.Type}, argName, "", description)
			} else {
				cmd.Flags().Int16VarP(addr.(*int16), argName, "", 0, description)
			}
		case reflect.Uint8:
			cmd.Flags().Uint8VarP(addr.(*uint8), argName, "", 0, description)
		case reflect.Uint16:
			cmd.Flags().Uint16VarP(addr.(*uint16), argName, "", 0, description)
		case reflect.String:
			typeName := field.Type.Name()
			if typeName != "string" {
				cmd.Flags().VarP(&stringerTypeGeneric{addr: addr, typ: field.Type}, argName, "", description)
			} else {
				cmd.Flags().StringVarP(addr.(*string), argName, "", "", description)
			}
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				ip := addr.(*net.IP)
				if *ip == nil || (*ip).Equal(net.IPv4zero) || (*ip).Equal(net.IPv6zero) {
					cmd.Flags().IPP(argName, "", *ip, description)
				}
			case reflect.String:
				cmd.Flags().StringSliceP(argName, "", *addr.(*[]string), description)
			default:
				switch field.Type.Elem() {
				case reflect.TypeOf(egoscale.ResourceTag{}):
					cmd.Flags().VarP(&tagGeneric{addr.(*[]egoscale.ResourceTag)}, argName, "", description)
				case reflect.TypeOf(egoscale.CIDR{}):
					cmd.Flags().VarP(&cidrListGeneric{addr.(*[]egoscale.CIDR)}, argName, "", description)
				case reflect.TypeOf(egoscale.UUID{}):
					cmd.Flags().VarP(&uuidListGeneric{addr.(*[]egoscale.UUID)}, argName, "", description)
				case reflect.TypeOf(egoscale.UserSecurityGroup{}):
					cmd.Flags().VarP(&userSecurityGroupListGeneric{addr.(*[]egoscale.UserSecurityGroup)}, argName, "", description)
				default:
					//log.Printf("[SKIP] Slice of %s is not supported!", field.Name)
				}
			}
		case reflect.Map:
			key := reflect.TypeOf(val.Interface()).Key()
			switch key.Kind() {
			case reflect.String:
				cmd.Flags().VarP(&mapGeneric{addr.(*map[string]string)}, argName, "", description)
			default:
				log.Printf("[SKIP] Type map for %s is not supported!", field.Name)
			}
		case reflect.Ptr:
			switch field.Type.Elem() {
			case reflect.TypeOf(true):
				cmd.Flags().VarP(&boolFlag{(addr.(**bool))}, argName, "", description)
			case reflect.TypeOf(egoscale.CIDR{}):
				cmd.Flags().VarP(&cidr{addr.(**egoscale.CIDR)}, argName, "", description)
			case reflect.TypeOf(egoscale.UUID{}):
				cmd.Flags().VarP(&uuid{addr.(**egoscale.UUID)}, argName, "", description)

			default:
				log.Printf("[SKIP] Ptr type of %s is not supported!", field.Name)
			}
		case reflect.Struct:
			buildFlag(val, cmd)
		default:
			log.Printf("[SKIP] Type of %s is not supported! %v", field.Name, val.Kind())
		}
	}
}
