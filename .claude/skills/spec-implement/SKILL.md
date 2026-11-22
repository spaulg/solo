---
name: spec-implement
description: Implement tasks from the active spec's tasks.md
argument-hint: [spec id] [task id]
disable-model-invocation: true
allowed-tools: Bash(git branch:*), Bash(ls:*), Bash(make lint:*), Bash(make test:*), Bash(make build:*), Read, Write, Edit, Grep, Glob
---
Current branch: !`git branch --show-current`
Existing specs: !`ls -1 docs/specs`
Arguments: $ARGUMENTS

## Resolve the active spec

1. If the arguments start with a spec id (e.g. `001`), use `../../../docs/specs`.
2. Otherwise, if the current branch matches `spec/NNN-name`, use `../../../docs/specs`.
3. Otherwise, list specs whose Status is not `Implemented` and ask me which one. Do not guess.

A task id (e.g. `T3`) may follow the spec id or appear alone. State the resolved spec and task scope before continuing.

## Preconditions

`spec.md` must be `Approved`, and `plan.md` and `tasks.md` must exist. Otherwise stop and tell me.

## Implement

Read `spec.md`, `plan.md`, and `tasks.md` in full first.

- Work through unchecked tasks in order, or only the given task.
- Tick each task (`- [x]`) in `tasks.md` once it is done and its tests pass.
- Only write tests when explicitly instructed to do so for the task
- Run `make lint` and `make test` before reporting a task complete. Report failures faithfully.
- If the implementation needs to diverge from the spec or plan, stop and propose an update to those documents. Do not drift silently.
- Do not implement behaviour that is not in the spec.

## Completion

When all tasks are ticked:

1. Walk through each acceptance criterion and name the test(s) that verify it.
2. If every AC is verified, set `Status: Implemented` in `spec.md`.
3. Summarise what changed.

Do not commit.
