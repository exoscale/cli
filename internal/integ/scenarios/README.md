# Testscript Integration Test Scenarios

This directory contains testscript scenarios for integration testing the Exoscale CLI.

## Current Scenarios

- **basic_no_api.txtar**: Tests basic CLI functionality without API access (version, help commands)
- **config_isolated.txtar**: Tests config file isolation using `XDG_CONFIG_HOME`

## Running Tests

```bash
# Run all scenarios
cd internal/integ
go test -v -run TestScripts

# Run specific scenario
go test -v -run 'TestScripts/basic_no_api'
```

## Future Work

**TODO**: API-based integration tests will be added in a separate PR. These will require:
- Organization test account setup
- Proper API credentials configuration (`EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`)
- Test scenarios covering:
  - Block storage operations (create, resize, snapshot, delete)
  - Compute instance operations
  - Network resources
  - Other API-dependent features

The CI workflow is already configured to run API tests on the master branch when credentials are available.
