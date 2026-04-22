# Claude Code Instructions for Terragrunt (Cycloid fork)

## Project Overview

Terragrunt is an orchestration tool on top of [OpenTofu](https://opentofu.org) / [Terraform](https://www.terraform.io) that adds features for managing large, multi-module IaC repos: remote-state config, DAG-aware runs across units, dependency wiring, code generation, catalogs, stacks, filters, and so on.

This repository is **Cycloid's fork** of the upstream [`gruntwork-io/terragrunt`](https://github.com/gruntwork-io/terragrunt). The module path stays `github.com/gruntwork-io/terragrunt` (do not change it), but:

- Our primary branch is `master` (upstream uses `main`).
- Our origin is `git@github.com-cycloid:cycloidio/terragrunt.git`.
- Upstream is periodically merged in via `update-upstream` PRs (see commit `8efe8675` style). Keep divergence from upstream small and well-scoped so those merges stay tractable.

Terragrunt is a **binary**, not a library — `main.go` wires `cli.NewApp(...)` and produces the `terragrunt` CLI. Downstream users run the binary; they don't import this module.

Key entry points:

- `main.go` — CLI bootstrap, log/exit-code setup.
- `cli/app.go` — command tree registration.
- `cli/commands/**` — one directory per top-level command (`run`, `stack`, `catalog`, `hcl`, etc.).
- `config/` — HCL parsing and `terragrunt.hcl` semantics (includes, locals, dependencies, feature flags, stacks).
- `internal/runner/`, `internal/queue/`, `internal/worker/` — unit execution engine and DAG scheduling.

Module path: `github.com/gruntwork-io/terragrunt`.

## Core Rules

- **Match existing conventions.** This is a large, long-lived codebase with a consistent house style. Follow what the surrounding file does for logging, errors, flags, and test layout rather than importing patterns from other projects.
- **Stay close to upstream.** When adding or fixing something that is obviously upstream-relevant (not Cycloid-specific), write it in a way that could be PR'd upstream and doesn't fight `update-upstream` merges. Avoid gratuitous refactors of files we don't need to touch.
- **English only**, no emojis in code or comments, no trailing periods on comments, state the *why* not the *what*.
- **Godoc style**: exported identifiers get a comment starting with the identifier name (`// Runner executes ...`).

## Repository Layout

```
terragrunt/
├── main.go                 # CLI entrypoint
├── cli/                    # CLI app, commands, flags (urfave/cli-based)
│   ├── app.go, help.go
│   ├── commands/           # one subdir per command
│   └── flags/              # shared / global flag definitions
├── config/                 # terragrunt.hcl parsing: includes, locals, deps, stacks, features
│   └── hclparse/           # HCL parser wrappers and diagnostics
├── codegen/                # generate blocks / backend + provider file generation
├── engine/                 # experimental plugin engine integration
├── options/                # TerragruntOptions (the big shared config struct)
├── pkg/log/                # public logging package (thin wrapper over logrus)
├── internal/               # private packages — import only from inside this repo
│   ├── cli/                # lower-level CLI primitives
│   ├── errors/             # error wrapping, stacks, Recover
│   ├── runner/, queue/, worker/     # unit execution / DAG scheduler
│   ├── discovery/, filter/          # unit discovery + filter-graph selection
│   ├── providercache/, remotestate/ # provider cache server + remote state backends
│   ├── awshelper/, git/, github/    # integrations
│   ├── cas/, cloner/, worktrees/    # content-addressable store, git cloning, worktrees
│   ├── component/, stacks/          # stacks feature
│   ├── locks/, cache/, strict/      # locking, caching, strict-mode checks
│   ├── experiment/, report/, view/  # experiments gating, run reports, TUI/output
│   ├── hclhelper/, ctyhelper/, os/  # small helpers
│   └── services/                    # cross-cutting services
├── shell/                  # running tofu/terraform subprocesses, error explanations
├── telemetry/              # OpenTelemetry setup
├── tf/                     # tofu/terraform detection, provider registry, getproviders, lock files
│   └── getproviders/       # has gomock'd interfaces (mocks/ subdir)
├── tflint/                 # tflint integration
├── util/                   # many small shared helpers
├── test/                   # integration tests (one file per feature area)
│   ├── fixtures/           # terragrunt.hcl + tofu fixtures used by integration tests
│   └── helpers/
├── docs-starlight/         # Astro/Starlight docs site (the live docs)
├── scripts/                # repo scripts (pre-commit, gofmtcheck, etc.)
├── _ci/                    # CI-only scripts (release verification)
├── mise.toml               # pinned toolchain (go, opentofu, golangci-lint, mockgen, licensei)
└── Makefile
```

## Build, Test & Code Quality

This repo uses a Makefile, but it's lean — most targets just wrap `go` / `golangci-lint`. Plain `go` commands work fine for everyday iteration.

```bash
make build              # go build -> ./terragrunt
make fmt                # gofmt -w on all non-vendored .go
make fmtcheck           # scripts/gofmtcheck.sh
make run-lint           # golangci-lint run -v --timeout=10m ./...
make run-strict-lint    # stricter config, only new issues vs origin/main
make generate-mocks     # go generate ./...   (regenerates mockgen outputs)
make license-check      # licensei cache + check + header
```

Testing:

- `go test ./...` works out of the box for unit tests — **there is no MySQL/docker dependency** (that's the Terracost repo, not this one).
- `test/integration_*_test.go` are the heavy integration suites. Many hit real clouds (AWS OIDC, GCP, Azure) or spawn real `tofu`/`terraform` processes; they expect credentials, network, and a working `tofu` binary on `$PATH`. Run individual integration tests with `go test -run TestX ./test/...` rather than the whole suite locally.
- Unit tests for a single package: `go test ./config/...`, `go test ./internal/runner/...`, etc.

Toolchain pinning lives in `mise.toml` (`go 1.25.0`, `opentofu 1.11.1`, `golangci-lint 2.4.0`, `mockgen v0.6.0`, `licensei v0.9.0`). Prefer `mise install` / `mise exec` to match CI exactly.

Generated files (`**/mocks/mock_*.go`, anything produced by a `//go:generate` directive) must never be hand-edited — change the source interface and re-run `make generate-mocks`. Mocks use `go.uber.org/mock` (uber's fork), not `golang/mock`.

## Code Conventions

- **Errors**: wrap via `github.com/gruntwork-io/terragrunt/internal/errors` (`errors.New`, `errors.Errorf`, `errors.As`, `errors.Recover`, `errors.ErrorStack`). Don't reach for the stdlib `errors` / `fmt.Errorf` in new code — the internal package preserves stack traces that the CLI surfaces on failure. Reuse existing sentinel/structured errors before inventing new ones.
- **Logging**: use `github.com/gruntwork-io/terragrunt/pkg/log`, typically passed around as `log.Logger` or pulled from context via `log.LoggerFromContext(ctx)`. Don't import `logrus` directly (it's an implementation detail of `pkg/log`) and don't introduce `slog` or `hclog`.
- **Options**: most functions that need user configuration take `*options.TerragruntOptions`. This struct is large — don't add fields casually; prefer threading a narrower type or reusing an existing field if one already fits.
- **Context**: public functions that do I/O (subprocesses, HTTP, cloud APIs, file walks) take `context.Context` as the first arg. Pure helpers don't.
- **Flags**: CLI flags live under `cli/flags/` and per-command `flags.go` files. Match the existing patterns (deprecated-name aliasing, env var wiring, strict-mode controls) when adding a flag.
- **Tests**: table-driven with `t.Run(name, func(t *testing.T) { ... })`. Use `testify/assert` and `testify/require` (already heavily used). Integration tests under `test/` follow a per-feature-area file naming convention (`integration_<feature>_test.go`) with fixtures under `test/fixtures/<feature>/`. Use `t.Parallel()` where safe — lint enforces `paralleltest`.
- **Mocks**: add `//go:generate mockgen -source=$GOFILE -destination=mocks/mock_$GOFILE -package=mocks` to the interface's file (see `tf/getproviders/` for the pattern), then `make generate-mocks`.

## Working with Upstream

- Commit messages use conventional-commit prefixes: `feat:`, `fix:`, `bug:`, `chore:`, `docs:`, `build(deps):`, etc. — match what's already in `git log`.
- PRs target `cycloidio/terragrunt` `master`. Do not open PRs directly against `gruntwork-io/terragrunt` from this checkout unless explicitly asked — those are separate workflows.
- When a bug is clearly upstream's (reproduces on `gruntwork-io/terragrunt` `main`), consider whether it's better to fix it upstream and pull via `update-upstream` vs. patching locally. Local patches are fine but should be minimal and clearly documented so they survive the next merge.

## Things NOT to Do

- Don't change the module path from `github.com/gruntwork-io/terragrunt`.
- Don't hand-edit generated mocks or any `_enumer.go` / generated file — re-run `make generate-mocks`.
- Don't import `logrus`, `slog`, or `hclog` directly; go through `pkg/log`.
- Don't replace `internal/errors` usage with stdlib `errors` / `fmt.Errorf` in existing code.
- Don't confuse this repo with Terracost — there is no MySQL, no `shopspring/decimal` pricing math, no `aws/`, `google/`, `azurerm/` provider packages in this repo. Pricing and the provider-ingester code live in `cycloidio/terracost`.
- Don't assume integration tests under `test/` are runnable offline — many need cloud credentials or a real `tofu`/`terraform` binary.
- Don't add dependencies casually; this is a CLI that ships to every Terragrunt user and every new dep widens the supply chain.
- Don't churn upstream files unrelated to the change at hand — it makes `update-upstream` merges painful.
