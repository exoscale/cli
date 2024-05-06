# Egoscale v3 (Alpha)

Exoscale API Golang wrapper

**Egoscale v3** is based on a generator written from scratch with [libopenapi](https://github.com/pb33f/libopenapi).

The core base of the generator is using libopenapi to parse and read the [Exoscale OpenAPI spec](https://openapi-v2.exoscale.com/source.yaml) and then generate the code from it.

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

## Development

### Generate Egoscale v3

From the root repo
```Bash
make generate
```

### Debug generator output

```Bash
mkdir test
GENERATOR_DEBUG=client make generate > test/client.go
GENERATOR_DEBUG=schemas make generate > test/schemas.go
GENERATOR_DEBUG=operations make generate > test/operations.go
```
