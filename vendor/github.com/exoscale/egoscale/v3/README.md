# Egoscale v3

Exoscale API Golang wrapper, generated from the [Exoscale OpenAPI spec](https://api-ch-gva-2.exoscale.com/v2/openapi.json).

## Installation

Install the following dependencies:

```shell
go get "github.com/exoscale/egoscale/v3"
```

Add the following import:

```golang
import "github.com/exoscale/egoscale/v3"
```
## Examples

```Golang
package main

import (
	"context"
	"log"

	"github.com/davecgh/go-spew/spew"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

func main() {
	creds := credentials.NewEnvCredentials()
	// OR
	creds = credentials.NewStaticCredentials("EXOxxx..", "...")

	client, err := v3.NewClient(creds)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	op, err := client.CreateInstance(ctx, v3.CreateInstanceRequest{
		Name:     "egoscale-v3",
		DiskSize: 50,
		// Ubuntu 24.04 LTS
		Template: &v3.Template{ID: v3.UUID("cbd89eb1-c66c-4637-9483-904d7e36c318")},
		// Medium type
		InstanceType: &v3.InstanceType{ID: v3.UUID("b6e9d1e8-89fc-4db3-aaa4-9b4c5b1d0844")},
	})
	if err != nil {
		log.Fatal(err)
	}

	op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	if err != nil {
		log.Fatal(err)
	}

	instance, err := client.GetInstance(ctx, op.Reference.ID)
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(instance)
}
```

### Logging

Egoscale supports debug logs which are off by default. To enable logs, use `ClientOptWithLogHandler` when initializing a client or `WithLogHandler` on existing client and provide `log/slog` compatible log handler to process them. For example, using the built-in JSON handler from `log/slog`, we can setup logs like so:

```go
	...
	opts := []v3.ClientOpt{
		v3.ClientOptWithLogHandler(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		)),
	}
	c, err := v3.NewClient(credentials.NewEnvCredentials(), opts...)
	...
```

Note that all logs printed by the client are of level Debug. A reloadable credentials provider
(see below) logs with its own handler, set with `ReloadableOptWithLogHandler`: its reloads at level
Info, and at level Error when it keeps the last credentials because new ones can't be read or are
invalid, so that a broken Secret is seen without debug logs.

### Findable

Most of the list request `ListX()` return a type containing the list of the resource requested and a method `FindX()` to be able to retrieve a resource by its `name` or `id` most of the time.
It returns an error wrapping `ErrNotFound` when no resource matches, and `ErrConflict` when several do.

```Golang
pools, err := client.ListInstancePools(ctx)
if err != nil {
	log.Fatal(err)
}
pool, err := pools.FindInstancePool("my-pool-example")
if err != nil {
	log.Fatal(err)
}

fmt.Println(pool.Name)
```

### Reloadable credentials

The client reads its credentials for every request. With a reloadable provider, a long-running
program (a Kubernetes controller, a daemon…) uses a rotated API key without being restarted or
creating a new client.

`credentials.NewReloadableCredentials` re-reads its source every 30 seconds, and right away when the
API rejects the key a request was signed with: the request is then signed again and retried once,
if the key changed.
New credentials are used as soon as they are read. If the source can't be read, is incomplete or
doesn't have the format of an Exoscale API key and secret, the last credentials are kept and the
error is logged on every re-read.

Two sources are provided:
- `credentials.KubernetesSecretSource(mountPath, apiKeyFile, apiSecretFile)`: the files of a
  Kubernetes Secret mounted as a volume, named after the keys of the Secret.
- `credentials.JSONFileSource(path)`: a JSON file with `api_key` and `api_secret` fields.

Any `func() (credentials.Value, error)` works as a source too.

The pieces can also be used separately: `credentials.NewReloadableProvider(source)` only re-reads
the source when the API rejects the key, and `credentials.NewPeriodicReloadProvider(ctx, p, interval)`
adds a periodic re-read to any `credentials.Reloader`, so other reload strategies can be built the
same way.

```Golang
// The periodic re-read stops when ctx is done.
creds, err := credentials.NewReloadableCredentials(ctx,
	credentials.KubernetesSecretSource("/etc/exoscale", "EXOSCALE_API_KEY", "EXOSCALE_API_SECRET"),
	// Optional: logs the reloads and the errors (nothing is logged by default).
	credentials.ReloadableOptWithLogHandler(slog.Default().Handler()),
)
if err != nil {
	log.Fatal(err)
}

client, err := v3.NewClient(creds, v3.ClientOptWithEndpoint(endpoint))
```

The Secret is mounted as a volume, not as environment variables, which a running process can't see
change. Don't use `subPath` either: the kubelet doesn't update `subPath` mounts.

```yaml
      containers:
        - name: controller
          volumeMounts:
            - name: exoscale-credentials
              mountPath: /etc/exoscale
              readOnly: true
      volumes:
        - name: exoscale-credentials
          secret:
            secretName: exoscale-credentials # keys: EXOSCALE_API_KEY, EXOSCALE_API_SECRET
```

The kubelet propagates a Secret update to the volume within about a minute: if the old key is
deleted before that, requests fail until the new key reaches the volume.

## Development

The SDK is generated: see the [generator](https://github.com/exoscale/egoscale/tree/master/v3/generator)
to regenerate it or change the generated code.
