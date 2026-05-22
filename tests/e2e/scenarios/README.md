# E2E Testscript Scenarios

Testscript scenarios for end-to-end testing the Exoscale CLI.

## Directory Structure

- **without-api/**: No API access required, runs by default
- **with-api/compute/**: Compute API scenarios (`-tags=api`, `-run TestScriptsAPICompute`)
- **with-api/dbaas/**: DBaaS API scenarios (`-tags=api`, `-run TestScriptsAPIDBaaS`)

## Running Tests

Build the binary first:

```bash
make build
```

### Without API (default)

```bash
cd tests/e2e
go test -v
```

### With API (CI — env vars)

```bash
cd tests/e2e
EXOSCALE_API_KEY=... EXOSCALE_API_SECRET=... \
  go test -v -tags=api -timeout 30m -run TestScriptsAPICompute
```

```bash
cd tests/e2e
EXOSCALE_API_KEY=... EXOSCALE_API_SECRET=... \
  go test -v -tags=api -timeout 30m -run TestScriptsAPIDBaaS
```

### With API (local — reads from exoscale.toml)

```bash
cd tests/e2e
go test -v -tags=local_integration -timeout 30m \
  -run TestAPIComputeLocal -account=<account-name>
```

```bash
cd tests/e2e
go test -v -tags=local_integration -timeout 30m \
  -run TestAPIDBaaSLocal -account=<account-name>
```

`-account` matches a substring of the account name in
`~/.config/exoscale/exoscale.toml`. Defaults to `owner-production`.
