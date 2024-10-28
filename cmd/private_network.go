package cmd

import (
	"fmt"
	"net"
	"strings"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var privateNetworkCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Private Networks management",
	Aliases: []string{"privnet"},
}

func processPrivateNetworkOptions(options []string) (*v3.PrivateNetworkOptions, error) {
	opts := &v3.PrivateNetworkOptions{}
	optionsMap := make(map[string][]string)

	// Process each option flag
	for _, opt := range options {
		keyValue := strings.SplitN(opt, "=", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("malformed option %q: must be in format key=\"value1 value2\"", opt)
		}
		key := keyValue[0]
		values := strings.Split(keyValue[1], " ")
		optionsMap[key] = append(optionsMap[key], values...)
	}

	// Process collected values
	for key, values := range optionsMap {
		switch key {
		case "dns-servers":
			for _, v := range values {
				if ip := net.ParseIP(v); ip != nil {
					opts.DNSServers = append(opts.DNSServers, ip)
				}
			}
		case "ntp-servers":
			for _, v := range values {
				if ip := net.ParseIP(v); ip != nil {
					opts.NtpServers = append(opts.NtpServers, ip)
				}
			}
		case "routers":
			for _, v := range values {
				if ip := net.ParseIP(v); ip != nil {
					opts.Routers = append(opts.Routers, ip)
				}
			}
		case "domain-search":
			opts.DomainSearch = values
		default:
			return nil, fmt.Errorf("unrecognized option key %q: supported keys are: 'dns-servers', 'ntp-servers', 'routers', 'domain-search'", key)
		}
	}
	return opts, nil
}

func init() {
	computeCmd.AddCommand(privateNetworkCmd)
}
