---
name: spec-plan
description: Write a technical plan (plan.md) for the active, approved spec
argument-hint: [spec id]
disable-model-invocation: true
allowed-tools: Bash(git branch:*), Bash(ls:*), Read, Write, Edit, Grep, Glob
---
Current branch: !`git branch --show-current`
Existing specs: !`ls -1 docs/specs`
Arguments: $ARGUMENTS

## Resolve the active spec

1. If the arguments start with a spec id (e.g. `001`), use `../../../docs/specs`.
2. Otherwise, if the current branch matches `spec/NNN-name`, use `../../../docs/specs`.
3. Otherwise, list specs whose Status is not `Implemented` and ask me which one. Do not guess.

State which spec you resolved before continuing.

## Preconditions

- `spec.md` must have `Status: Approved`.
- `spec.md` must contain no `[NEEDS CLARIFICATION]` markers.

If either fails, stop and tell me what needs resolving.

## Write the plan

1. Explore the relevant code under `internal/pkg/` and `cmd/` and follow existing patterns, naming, and package structure.
2. Write `plan.md` in the spec folder from `../../../docs/specs`.
3. Every requirement must appear in the requirement mapping; every acceptance criterion must appear in the test strategy.
4. Note any alternatives considered and why they were rejected.
5. If planning reveals a problem with the spec itself, stop and propose a spec change instead of working around it.

Stop for review. Do not write tasks or code. Do not commit.
