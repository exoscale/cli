package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func buildClient() {
	if ignoreClientBuild {
		return
	}

	if globalstate.EgoscaleV3Client != nil {
		return
	}

	clientTimeout := account.CurrentAccount.ClientTimeout
	if clientTimeout == 0 {
		clientTimeout = DefaultClientTimeout
	}

	creds := credentials.NewStaticCredentials(
		account.CurrentAccount.Key,
		account.CurrentAccount.APISecret(),
	)

	clientV3, err := v3.NewClient(
		creds,
		v3.ClientOptWithRequestInterceptors(func(ctx context.Context, req *http.Request) error {
			for k, v := range account.CurrentAccount.CustomHeaders {
				req.Header.Add(k, v)
			}

			return nil
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to initialize Exoscale API V3 client: %v", err))
	}

	if account.CurrentAccount.Endpoint != "" {
		clientV3 = clientV3.WithEndpoint(v3.Endpoint(account.CurrentAccount.Endpoint))
	}

	if v := os.Getenv("EXOSCALE_TRACE"); v != "" {
		clientV3 = clientV3.WithTrace()
	}
	globalstate.EgoscaleV3Client = clientV3
}

func SwitchClientZoneV3(ctx context.Context, client *v3.Client, zone v3.ZoneName) (*v3.Client, error) {
	if zone == "" {
		return client, nil
	}
	endpoint, err := client.GetZoneAPIEndpoint(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("switch client zone v3: %w", err)
	}

	return client.WithEndpoint(endpoint), nil
}
