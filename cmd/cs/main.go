package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/go-ini/ini"
	"github.com/urfave/cli"
)

var _client = new(egoscale.Client)

func main() {
	// global flags
	var debug bool
	var dryRun bool
	var dryJSON bool
	var region string
	var theme string

	app := cli.NewApp()
	app.Name = "cs"
	app.HelpName = "cs"
	app.Usage = "CloudStack at the fingerprints"
	app.Description = "Exoscale Go CloudStack cli"
	app.HideVersion = true
	app.Compiled = time.Now()
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "debug mode on",
			Destination: &debug,
		},
		cli.BoolFlag{
			Name:        "dry-run, D",
			Usage:       "produce a cURL ready URL",
			Destination: &dryRun,
			Hidden:      true,
		},
		cli.BoolFlag{
			Name:        "dry-json, j",
			Usage:       "produce a JSON preview of the query",
			Destination: &dryJSON,
			Hidden:      true,
		},
		cli.StringFlag{
			Name:        "region, r",
			Usage:       "cloudstack.ini file section name",
			Value:       "cloudstack",
			Destination: &region,
		},
		cli.StringFlag{
			Name:        "theme, t",
			Usage:       "syntax highlighting theme, see: https://xyproto.github.io/splash/docs/",
			Value:       "",
			Destination: &theme,
		},
	}

	var method egoscale.Command
	app.Commands = buildCommands(&method, methods)

	app.Run(os.Args)

	client, _ := buildClient(region)
	if theme != "" {
		client.Theme = theme
	}

	if method == nil {
		os.Exit(0)
	}

	// Show request and quit
	if debug {
		payload, err := client.Payload(method)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, client.Endpoint)
		fmt.Fprint(os.Stdout, "\\\n?")
		fmt.Fprintln(os.Stdout, strings.Replace(payload, "&", "\\\n&", -1))

		response := client.Response(method)
		resp, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(os.Stdout, "")
		printJSON(string(resp), client.Theme)
		os.Exit(0)
	}

	if dryRun {
		payload, err := client.Payload(method)
		if err != nil {
			log.Fatal(err)
		}
		signature, err := client.Sign(payload)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, client.Endpoint)
		fmt.Fprint(os.Stdout, "?")
		fmt.Fprintln(os.Stdout, signature)
		os.Exit(0)
	}

	if dryJSON {
		request, err := json.MarshalIndent(method, "", "  ")
		if err != nil {
			log.Panic(err)
		}

		printJSON(string(request), client.Theme)
		os.Exit(0)
	}

	resp, err := client.Request(method)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(&resp, "", "  ")
	printJSON(string(out), client.Theme)
}

func buildClient(region string) (*Client, error) {
	usr, _ := user.Current()
	localConfig, _ := filepath.Abs("cloudstack.ini")
	inis := []string{
		localConfig,
		filepath.Join(usr.HomeDir, ".cloudstack.ini"),
	}
	config := ""
	for _, i := range inis {
		if _, err := os.Stat(i); err != nil {
			continue
		}
		config = i
		break
	}

	if config == "" {
		log.Fatalf("Config file not found within: %s", strings.Join(inis, ", "))
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, config)
	if err != nil {
		log.Fatal(err)
	}

	section, err := cfg.GetSection(region)
	if err != nil {
		log.Fatalf("Section %q not found in the config file %s", region, config)
	}
	endpoint := "https://api.exoscale.ch/compute"
	ep, err := section.GetKey("endpoint")
	if err == nil {
		endpoint = ep.String()
	}

	key, errKey := section.GetKey("key")
	secret, errSecret := section.GetKey("secret")

	if errKey != nil || errSecret != nil {
		log.Fatalf("Section %q is missing key or secret", region)
	}

	cs := egoscale.NewClient(endpoint, key.String(), secret.String())

	client := &Client{cs, ""}

	section, err = cfg.GetSection("exoscale")
	if err == nil {
		theme, _ := section.GetKey("theme")
		client.Theme = theme.String()
	}

	return client, nil
}

func buildCommands(out *egoscale.Command, methods map[string][]cmd) []cli.Command {
	commands := make([]cli.Command, 0)

	for category, ms := range methods {
		for i := range ms {
			s := ms[i]
			commands = append(commands, buildCommand(out, category, s.command, s.hidden))
		}
	}

	return commands
}

func buildCommand(out *egoscale.Command, category string, method egoscale.Command, hidden bool) cli.Command {
	command := cli.Command{
		Name:     _client.APIName(method),
		Category: category,
		HideHelp: hidden,
		Hidden:   hidden,
		Action: func(c *cli.Context) error {
			*out = method
			return nil
		},
	}

	command.Flags = buildFlags(method)

	return command
}

func buildFlags(method egoscale.Command) []cli.Flag {
	flags := make([]cli.Flag, 0)

	val := reflect.ValueOf(method)
	// we've got a pointer
	value := val.Elem()

	if value.Kind() != reflect.Struct {
		log.Fatalf("struct was expected")
		return flags
	}

	ty := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := ty.Field(i)

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
				argName = strings.ToLower(field.Name)
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
			flags = append(flags, cli.BoolFlag{
				Name:        argName,
				Usage:       description,
				Destination: addr.(*bool),
			})
		case reflect.Int:
			flags = append(flags, cli.IntFlag{
				Name:        argName,
				Usage:       description,
				Destination: addr.(*int),
			})
		case reflect.Int64:
			if argName == "resourcetype" {
				flags = append(flags, cli.GenericFlag{
					Name:  argName,
					Usage: description,
					Value: &resourceTypeGeneric{
						value: addr.(*egoscale.ResourceType),
					},
				})
			} else {
				flags = append(flags, cli.Int64Flag{
					Name:        argName,
					Usage:       description,
					Destination: addr.(*int64),
				})
			}
		case reflect.Uint:
			flags = append(flags, cli.UintFlag{
				Name:        argName,
				Usage:       description,
				Destination: addr.(*uint),
			})
		case reflect.Uint64:
			flags = append(flags, cli.Uint64Flag{
				Name:        argName,
				Usage:       description,
				Destination: addr.(*uint64),
			})
		case reflect.Float64:
			flags = append(flags, cli.Float64Flag{
				Name:        argName,
				Usage:       description,
				Destination: addr.(*float64),
			})
		case reflect.Int16:
			flag := cli.GenericFlag{
				Name:  argName,
				Usage: description,
			}
			if argName == "accounttype" {
				flag.Value = &accountTypeGeneric{
					value: addr.(*egoscale.AccountType),
				}
			} else {
				flag.Value = &int16Generic{
					value: addr.(*int16),
				}
			}
			flags = append(flags, flag)
		case reflect.Uint8:
			flags = append(flags, cli.GenericFlag{
				Name:  argName,
				Usage: description,
				Value: &uint8Generic{
					value: addr.(*uint8),
				},
			})
		case reflect.Uint16:
			flags = append(flags, cli.GenericFlag{
				Name:  argName,
				Usage: description,
				Value: &uint16Generic{
					value: addr.(*uint16),
				},
			})
		case reflect.String:
			if argName == "resourcetypename" {
				flags = append(flags, cli.GenericFlag{
					Name:  argName,
					Usage: description,
					Value: &resourceTypeNameGeneric{
						value: addr.(*egoscale.ResourceTypeName),
					},
				})

			} else {
				flags = append(flags, cli.StringFlag{
					Name:        argName,
					Usage:       description,
					Destination: addr.(*string),
				})
			}
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				ip := addr.(*net.IP)
				if *ip == nil || (*ip).Equal(net.IPv4zero) || (*ip).Equal(net.IPv6zero) {
					flags = append(flags, cli.GenericFlag{
						Name:  argName,
						Usage: description,
						Value: &ipGeneric{
							value: ip,
						},
					})
				}
			case reflect.String:
				flags = append(flags, cli.StringSliceFlag{
					Name:  argName,
					Usage: description,
					Value: (*cli.StringSlice)(addr.(*[]string)),
				})
			default:
				switch field.Type.Elem() {
				case reflect.TypeOf(egoscale.ResourceTag{}):
					flags = append(flags, cli.GenericFlag{
						Name:  argName,
						Usage: description,
						Value: &tagGeneric{
							value: addr.(*[]egoscale.ResourceTag),
						},
					})
				default:
					//log.Printf("[SKIP] Slice of %s is not supported!", field.Name)
				}
			}
		case reflect.Map:
			key := reflect.TypeOf(val.Interface()).Key()
			switch key.Kind() {
			case reflect.String:
				flags = append(flags, cli.GenericFlag{
					Name:  argName,
					Usage: description,
					Value: &mapGeneric{
						value: addr.(*map[string]string),
					},
				})
			default:
				log.Printf("[SKIP] Type map for %s is not supported!", field.Name)
			}
		case reflect.Ptr:
			switch field.Type.Elem().Kind() {
			case reflect.Bool:
				flags = append(flags, cli.GenericFlag{
					Name:  argName,
					Usage: description,
					Value: &boolPtrGeneric{
						value: addr.(**bool),
					},
				})
			default:
				log.Printf("[SKIP] Ptr type of %s is not supported!", field.Name)
			}
		default:
			log.Printf("[SKIP] Type of %s is not supported! %v", field.Name, val.Kind())
		}
	}

	return flags
}

// Client holds the internal meta information for the cli
type Client struct {
	*egoscale.Client
	Theme string
}
