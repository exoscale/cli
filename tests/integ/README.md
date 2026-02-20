# Integration Tests

This directory contains integration tests for the Exoscale CLI.

## Directory Structure

```
tests/integ/
├── local/          # Tests that don't require API credentials
├── api/            # Tests that require real API credentials (build tag: integration_api)
└── suite.go        # Shared test utilities
```

## Running Tests

### Local Tests (No API Required)

These tests verify CLI behavior without making actual API calls:

```bash
cd tests/integ/local
go test -v
```

Or from the root:
```bash
go test -v ./tests/integ/local/...
```

**Tests in local/:**
- `config_panic_test.go` - Tests config command behavior with missing default account

### API Tests (API Credentials Required)

These tests make real API calls and require valid Exoscale credentials:

```bash
cd tests/integ/api
go test -v -tags=integration_api
```

Or from the root:
```bash
go test -v -tags=integration_api ./tests/integ/api/...
```

**Tests in api/:**
- `blockstorage_test.go` - Tests block storage volume operations (creates/deletes real resources)

**Note:** API tests require:
- Valid Exoscale API credentials in `~/.config/exoscale/exoscale.toml`
- Or credentials via environment variables
- Tests will fail with "no accounts configured" if credentials are missing

## CI/CD Integration

- **Local tests** are run automatically in CI/CD as they don't require credentials
- **API tests** are NOT run in CI/CD by default (require the `integration_api` build tag)
- API tests can be run manually or in a separate CI job with secrets configured

## Adding New Tests

### Local Test (No API)
1. Create test file in `tests/integ/local/`
2. Use package `integ_local_test`
3. No build tags needed

### API Test (Requires API)
1. Create test file in `tests/integ/api/`
2. Use package `integ_api_test`
3. Add build tags at the top:
   ```go
   //go:build integration_api
   // +build integration_api
   ```
4. Import the suite if needed: `import "github.com/exoscale/cli/internal/integ"`
