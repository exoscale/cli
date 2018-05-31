package client

import (
	"log"

	"github.com/go-ini/ini"

	"github.com/exoscale/egoscale"
)

//BuildClient get cs client with a cfg file path and ini file region
func BuildClient(config, region string) (*egoscale.Client, error) {

	if config == "" {
		log.Fatalf("Config file not found")
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	section, err := cfg.GetSection(region)
	if err != nil {
		log.Fatalf("Section %q not found in the config file %s", region, config)
	}
	endpoint, _ := section.GetKey("endpoint")
	key, _ := section.GetKey("key")
	secret, _ := section.GetKey("secret")

	client := egoscale.NewClient(endpoint.String(), key.String(), secret.String())

	return client, nil
}
