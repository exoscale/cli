# AGENTS.md

## Command pattern - always use RegisterCLICommand

All new commands must use the struct-based pattern.
**Do not copy the old flat `var cmd = &cobra.Command{...}` commands** - they
exist in `cmd/storage/`, `cmd/dns/`, `cmd/dbaas/` and are legacy.

Read this file before writing any new command:
`cmd/aiservices/deployment/deployment_create.go`

For subcommand placement in an existing group, also read `## Folder layout` below.

The command framework (`RegisterCLICommand`, the flag helpers, `BinaryValue`,
the aliases and the default constants) lives in `pkg/clicmd`, imported as
`"github.com/exoscale/cli/pkg/clicmd"`. The root `cmd` package (`exocmd`) only
holds `RootCmd`, the config loading and the global state such as `GContext` and
`SwitchClientZoneV3`. Don't add generic helpers to it.

Skeleton:

```go
type myResourceCreateCmd struct {
    clicmd.CliCommandSettings `cli-cmd:"-"`

    _ bool `cli-cmd:"create"`

    Name string `cli-arg:"#"  cli-usage:"NAME"`        // required positional
    Tag  string `cli-arg:"?"  cli-usage:"TAG"`          // optional positional
    Zone string `cli-short:"z" cli-usage:"zone"`        // becomes --zone / -z flag
}

func (c *myResourceCreateCmd) CmdAliases() []string { return clicmd.GCreateAlias }
func (c *myResourceCreateCmd) CmdShort() string     { return "Create a resource" }
func (c *myResourceCreateCmd) CmdLong() string      { return "..." }

func (c *myResourceCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
    clicmd.CmdSetZoneFlagFromDefault(cmd)
    return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *myResourceCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
    // c.Name, c.Zone are already populated here
    return nil
}

func init() {
    cobra.CheckErr(clicmd.RegisterCLICommand(parentCmd, &myResourceCreateCmd{
        CliCommandSettings: clicmd.DefaultCLICmdSettings(),
    }))
}
```

Supported struct tags:

| Tag | Purpose |
|---|---|
| `cli-cmd:"-"` | Exclude field from CLI processing |
| `cli-cmd:"name"` | Sets the cobra command name (use on a `_ bool` field) |
| `cli-arg:"#"` | Required positional argument |
| `cli-arg:"?"` | Optional positional argument; `[]string` becomes variadic |
| `cli-flag:"flag-name"` | Override the auto-derived kebab-case flag name |
| `cli-short:"z"` | Single-character shorthand |
| `cli-usage:"text"` | Flag usage string; also used as positional arg label in help |
| `cli-hidden:""` | Mark flag hidden |
| `cli-deprecated:"msg"` | Mark flag deprecated |

Supported field types: `string`, `int64`, `bool`, `[]string`, `map[string]string`,
and any type whose pointer implements `pflag.Value` (flag or single positional arg,
not variadic). Any other type panics at startup.

For binary input (keys, ciphertexts, signatures...) use `clicmd.BinaryValue`
(`pkg/clicmd/binary_value.go`) instead of a base64 `string`. The field holds the
decoded bytes, and users can pass:

- `<base64>` - the value itself, base64-encoded
- `file://<path>` - a file containing the base64-encoded value
- `fileb://<path>` - a file containing the raw bytes

```go
Payload clicmd.BinaryValue `cli-arg:"#" cli-usage:"PAYLOAD"`
Context clicmd.BinaryValue `cli-usage:"encryption context"` // --context flag
```

Global alias vars to use for `CmdAliases()`:
`clicmd.GListAlias`, `GShowAlias`, `GCreateAlias`, `GDeleteAlias`, `GRemoveAlias`.

### Deprecating a command

Don't edit `Short`/`Long` by hand and don't use cobra's `Deprecated` field (it
hides the command from its parent's help). Use `clicmd.Deprecate` (`pkg/clicmd/deprecation.go`):

- a group: wrap its declaration, its subcommands inherit the deprecation:
  `var fooCmd = clicmd.Deprecate(&cobra.Command{...}, "Foo is deprecated, use bar instead")`
- a single struct command: implement `CmdDeprecated() string` and return the message.

The command stays listed in the help with `(deprecated)` after its short
description, the message is put on top of its long description, and it is
printed as a warning on stderr when the command runs.

## Folder layout

**Anti-pattern: don't extend flat layouts.** `cmd/storage/`, `cmd/dns/`,
`cmd/dbaas/` and similar are legacy flat groups. Adding or refactoring a
subcommand under one of them MUST promote the cluster to its own subpackage
following the layout already used elsewhere in the repo —
`cmd/aiservices/deployment/` is the canonical example (`deployment.go`
declares `var Cmd`; siblings live as `deployment_<verb>.go` in the same dir).
Concrete case in flight: `cmd/storage/storage_bucket_lifecycle*.go` ->
`cmd/storage/bucket/lifecycle/{lifecycle,lifecycle_show,lifecycle_set,lifecycle_delete}.go`.

New command group:

```
cmd/<domain>/<subdomain>/
  <subdomain>.go          <- var Cmd; init -> parent.AddCommand(Cmd)
  <subdomain>_list.go
  <subdomain>_show.go
  <subdomain>_create.go
  <subdomain>_delete.go
  <subdomain>_list_test.go
  ...
```

Then add a blank import to `cmd/subcommands/init.go` - nothing else needs wiring.

## Tests

The `.txtar` scenarios under `tests/e2e/` are the source of truth for the CLI:
they run the real binary, and the `with-api/` ones run it against the real API.

**Don't mock the Exoscale API.** In new tests, don't add mocks, fake clients or
`httptest` servers that imitate API responses, and don't extend the existing
ones (e.g. those built on `pkg/testutils`). The API is not versioned: a mock
freezes a copy of a contract that changes under us, goes out of sync silently,
and has to be updated by hand every time the API moves. Test API interactions
with a `with-api/` scenario instead (see below).

Go unit tests are still welcome for code that doesn't talk to the API: utils,
parsing, flag and argument validation, output formatting, pure logic. Tests live
in the same package (white-box). When the logic you want to test sits behind an
API call (merging results from several zones, building a request, deciding
what to print), move it into a plain function that takes egoscale types as
input, and unit test that function. Leave the API call itself to the txtar
scenario.

Behaviour that needs no API access (argument errors, `--help`, deprecation
warnings, errors raised before any call) goes in a scenario under
`tests/e2e/scenarios/without-api/`. Those run by default, with no credentials.

`httptest` is only acceptable for generic HTTP helpers (timeouts, retries,
endpoint selection) whose server returns nothing Exoscale-specific, as in
`utils/utils_test.go`.

When you touch an existing test that mocks the API, replacing it with a txtar
scenario is welcome. It is not required.

### E2E tests for new features

Every new feature must ship with at least one `.txtar` scenario under
`tests/e2e/scenarios/with-api/<domain>/`. A single scenario covering only
the happy path is not enough - cover the primary CLI usages:

- The main command (create/list/show/delete or equivalent)
- At least one flag or option variant that changes the output or behaviour
- Any path-syntax shorthand the user would reach for (e.g. `sos://bucket/`)

Keep scenarios focused: one file per logical feature or sub-command cluster.
Name files `<feature>_<scenario>.txtar` (e.g. `bucket_lifecycle_crud.txtar`).

Register the domain under `TestScriptsAPI<Domain>` in
`tests/e2e/testscript_api_test.go` and `TestAPI<Domain>Local` in
`tests/e2e/local_test.go`.

Run locally with:

```
cd tests/e2e
go test -v -tags=local_integration -timeout 30m -run TestAPI<Domain>Local -account=<name>
```

## File header on every AI-touched file

Add a comment at the top of every file you create or modify. Use the
appropriate syntax for the language, and pick the right verb. Substitute
`<model>` with your own identifier (e.g. `claude-sonnet-4.6`, `gpt-5`, etc.):

```go
// AI-generated by <model> - not reviewed yet   // new file
// AI-modified by <model> - not reviewed yet    // existing file
```
```yaml
# AI-generated by <model> - not reviewed yet
# AI-modified by <model> - not reviewed yet
```
```html
<!-- AI-generated by <model> - not reviewed yet -->
<!-- AI-modified by <model> - not reviewed yet -->
```

The reviewer removes the line once they have read and approved the file.
It must be the very first line (before `package`, shebangs, etc.) except
where the language requires otherwise (e.g. after `<?xml` declarations).

## Marking AI-assisted commits and PRs

Add these trailers to every commit that contains agent-generated code:

    AI-assisted: true
    Co-authored-by: copilot-swe-agent <203248971+copilot-swe-agent@users.noreply.github.com>

This makes AI-touched commits greppable across history:

    git log --grep="AI-assisted"

and shows Copilot as a co-author in GitHub's commit view.

For PRs, append this footer to the description (after a `---`), listing only categories that apply:

```
> [!NOTE]
> AI assistance: PR description, test scaffolding, boilerplate code.
```

Omit categories that don't apply. Omit the footer entirely if AI contributed nothing.

## Before opening a PR or issue

Use the existing templates - don't replace them with free-form text:
- PRs: `.github/pull_request_template.md`
- Issues: `.github/ISSUE_TEMPLATE/` (bug-report, feature_request, other_request)

Keep every heading, fill each section in.

## Before opening a PR

```
make build
make test-verbose
golangci-lint run --timeout 4m
```

Update `CHANGELOG.md` under `## Unreleased` with the PR number.
