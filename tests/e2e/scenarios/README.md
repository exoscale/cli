# E2E Testscript Scenarios

This directory contains testscript scenarios for end-to-end testing the Exoscale CLI.

## Directory Structure

- **without-api/**: Test scenarios that don't require API access (run by default)
- **with-api/**: Test scenarios that require API access (run with `-tags=api`)

## Current Scenarios

### Tests Without API (scenarios/without-api/)

- **basic_no_api.txtar**: Tests basic CLI functionality without API access (version, help commands)
- **config_isolated.txtar**: Tests config file isolation using `XDG_CONFIG_HOME`

### Tests With API (scenarios/with-api/)

No API test scenarios yet. These will be added in a future PR.

## Running Tests

Tests use the pre-built binary from the existing build pipeline. Build it first:

```bash
# Build the CLI binary (from repository root)
make build

# Run local tests only (default - no build tag needed)
cd tests/e2e
go test -v

# Run API tests only (requires API credentials)
cd tests/e2e
go test -v -tags=api

# Run all tests (local + API)
cd tests/e2e
go test -v -tags=api
```

## Using Build Tags

The test suite uses Go build tags to separate local tests from API tests:

- **Local tests** (`testscript_local_test.go`): No build tag, runs by default
- **API tests** (`testscript_api_test.go`): Requires `-tags=api` build tag

This approach is more maintainable than regex filtering and follows Go best practices.

## Future Work

**TODO**: API-based tests will be added in a separate PR to `scenarios/with-api/`. These will require:
- Organization test account setup
- Proper API credentials configuration (`EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`)
- Test scenarios covering:
  - Block storage operations (create, resize, snapshot, delete)
  - Compute instance operations
  - Network resources
  - Other API-dependent features

The CI workflow is already configured to run API tests on the master branch when credentials are available.
