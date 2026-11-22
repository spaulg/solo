# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Solo wraps Docker Compose to build containerised dev environments, running user-defined workflow steps inside containers on lifecycle events (start, stop, rebuild) and exposing "tools" (commands run in a service).

## Build and test

- `make build` — generate protobuf code (`protoc`) and build both binaries into `.build/output/`
- `make test` — run tests with `gotestsum` (coverage in `.build/tests/coverage.txt`); `make citest` is the CI variant
- `make lint` — run `golangci-lint`
- Single package/test: `go test ./internal/pkg/host/app/wms/ -run TestWorkflowRunnerTestSuite/TestName` (tests use testify suites, so `-run` takes `Suite/Method`)
- Regenerate bubbletea golden files (`testdata/*.golden`, via teatest): add `-update` to the `go test` command

`*.pb.go` files are generated and gitignored — run `make build` (or `make protos`) after a fresh clone or a change to `internal/pkg/common/infra/grpc/services/workflow.proto`, otherwise the code won't compile.

Imports are formatted with `goimports -local github.com/spaulg/solo` (enforced by a lefthook pre-commit hook and the linter). The linter forbids globals and `init()` functions (`gochecknoglobals`, `gochecknoinits`).

## Architecture

There are two binaries:

- `cmd/solo` — host CLI (cobra). It loads `solo.yml` (a Compose file with `x-tools` and per-service `x-workflows` extensions) and `solo-config.yml`, generates a plain compose file in the project state directory (`.solo/`), runs `docker compose`, and serves a gRPC server over mutual TLS using a self-signed CA it generates.
- `cmd/solo-entrypoint` — built for Linux, copied into the state directory and used as the container entrypoint. On a lifecycle event it connects back to the host over gRPC (`Workflow.RunWorkflowStream`, a bidirectional stream): it sends the workflow name, the host replies with commands to run, and the entrypoint streams back stdout/stderr/exit codes until the host sends `COMPLETE_ACTION`.

On `start`, `host/app/ProjectControl` builds the orchestrator, launches the gRPC server, starts compose, then blocks on a workflow guard until every container finishes the `first_pre_start_*`, `pre_start_*`, `post_start_*` and `first_post_start_*` workflows (`_container` runs per replica, `_service` once per service). Workflow execution, step logs and execution events are recorded as JSON files in the state directory by the auditor, which subscribes to the event manager. The `wflog` command reads these back in a bubbletea TUI (`host/app/bubbletea`).

The `solo.yml` in the repo root is a dev fixture: it runs the Symfony example in `examples/symfony-webapp`, mounting the source in so containers run the entrypoint with `go run` (`solo-entrypoint-wrapper.sh`).

### Layering (enforced by tests)

`internal/pkg/architecture_test.go` fails the test suite on violations:

- Code is split into `host`, `entrypoint` and `common`, each with `app` / `domain` / `infra` layers. `domain` depends only on domain; `infra` on infra and domain; `app` may use everything in its own area plus `common`.
- `cmd/solo` may only import `internal/pkg/host`; `cmd/solo-entrypoint` only `internal/pkg/entrypoint` and `common/domain`.
- `host/app` has an allow-list of external dependencies (cobra, bubbletea/bubbles/lipgloss, flock, ulid). Adding a new third-party import to `host/app` means updating that list.
- `New*` constructors must return concrete types, never an interface declared in this module.

Mocks are handwritten with `testify/mock` in `test/mocks/`, mirroring the package path they mock. Shared test helpers and fixtures are in `test/helpers.go` and `test/data/`.

## Spec-driven development

See `docs/specs` for the workflow. Skills: `/spec`, `/spec-plan`, `/spec-tasks`, `/spec-implement`.

- Non-trivial features and behaviour changes start with a spec in `docs/specs` (use `/spec`).
- Bug fixes, dependency bumps and pure refactors do not need a spec.
- `docs/specs` is the source of truth for behaviour. If code and spec disagree, raise it rather than silently picking one.
- Do not implement from a spec that is in `Draft` status or contains `[NEEDS CLARIFICATION]` markers.
- **Active spec:** if the current git branch is `spec/NNN-name`, the active spec is `docs/specs`. Before changing code on such a branch, read its `spec.md`, `plan.md` and `tasks.md`, and keep `tasks.md` up to date as work is completed.
- Reference the spec id in commit messages, e.g. `Add history command (spec 001)`.
