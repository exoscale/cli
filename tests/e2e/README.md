# E2E tests

Quick reference for running and writing CLI e2e tests.

## Run

```bash
make build
cd tests/e2e
go test -v                       # local only
go test -v -tags=api -run X      # api only, set EXOSCALE_API_KEY and EXOSCALE_API_SECRET first
```

`-tags=api` needs `EXOSCALE_API_KEY`, `EXOSCALE_API_SECRET`. `EXOSCALE_ZONE` is optional, defaults to `ch-gva-2`.

## Layout

- `scenarios/without-api/` runs by default, no credentials needed.
- `scenarios/with-api/` runs with `-tags=api`, touches the real org.
- Custom commands are defined in `testscript_api_test.go`.

## Resource naming

API tests share an Exoscale org with other repos (terraform-provider, csi-driver, ...). The runner injects a unique `TEST_RUN_ID` prefixed with `cli-e2e-` into every scenario. Use it for every resource you create.

- good: `${TEST_RUN_ID}-nlb-a`, `cli-e2e-pg-$TEST_RUN_ID`
- bad: hardcoded names, or `e2e-...` without the `cli-` part

Placeholders that intentionally don't exist (for not-found error tests) are fine as-is, e.g. `nonexistent-e2e-instance`.

## Writing a scenario

- File goes under `scenarios/with-api/<area>/`.
- `exec exo ...` runs the binary against the isolated workdir and env the runner sets up.
- `json-setenv VAR FIELD FILE` reads `FILE` (relative to workdir) and sets `VAR` to a top-level JSON field.
- `wait-instance-state ZONE ID TARGET [SECS]` polls instance show until state matches, default 300s.
- `wait-dbaas-state ZONE NAME TARGET [SECS]` same for dbaas, default 600s.
- `execpty --stdin=<file> <bin> [args...]` runs `<bin>` inside a PTY and feeds tokens from `<file>`. Tokens can be literals, arrow key names (`@down`, `@up`, ...), or `@wait:<pattern>` to delay the next input until `<pattern>` appears in PTY output.
- Each scenario creates and deletes its own resources. The runner does not clean up after you.
