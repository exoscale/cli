package x

import (
	"os"
	"path"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	logger "github.com/izumin5210/gentleman-logger"
	"github.com/spf13/cobra"
	"gopkg.in/h2non/gentleman.v2/context"
	"gopkg.in/h2non/gentleman.v2/plugin"
)

// InitCommand initializes the code-generated CLI instance and returns the resulting Cobra command
// to the caller so it can be added as subcommand to a higher level CLI.
func InitCommand() *cobra.Command {
	cli.Init(&cli.Config{AppName: "x"})
	cli.AddGlobalFlag("zone", "z", "Exoscale zone", "")
	cli.AddGlobalFlag("environment", "e", "Exoscale API environment", "")

	if _, ok := os.LookupEnv("EXOSCALE_TRACE"); ok {
		cli.Client.Use(logger.New(os.Stderr))
	}

	xRegister(false)

	// We attempts to clean up the unused configuration cache file "$HOME/.x/cache.json" created by
	// cli.Init(). Since we don't report errors as the `exo x` command is not advertised, this is
	// best effort.
	cacheFile := cli.Cache.ConfigFileUsed()
	if err := os.Remove(cacheFile); err == nil {
		os.Remove(path.Dir(cacheFile)) // golint:noerr
	}

	return cli.Root
}

// SetClientCredentials adds a pre-request hook to sign outgoing requests using specified API credentials.
func SetClientCredentials(apiKey, apiSecret string) error {
	security, err := apiv2.NewSecurityProviderExoscale(apiKey, apiSecret)
	if err != nil {
		return err
	}

	// Intercept the outgoing API request and sign it
	cli.Client.Use(plugin.NewPhasePlugin("before dial",
		func(ctx *context.Context, h context.Handler) {
			if err := security.Intercept(ctx, ctx.Request); err != nil {
				panic(err)
			}

			h.Next(ctx)
		}),
	)

	return nil
}

// SetClientUserAgent adds a pre-request hook to set the User-Agent header value on outgoing HTTP requests.
func SetClientUserAgent(ua string) {
	cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Header.Set("User-Agent", ua)
		h.Next(ctx)
	})
}
