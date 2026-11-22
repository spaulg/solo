---
name: spec-tasks
description: Break the active spec's plan into an ordered tasks.md checklist
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

`spec.md` must be `Approved` and `plan.md` must exist. Otherwise stop and tell me.

## Write the tasks

Write `tasks.md` in the spec folder from `../../../docs/specs`:

- Order tasks so each builds on the previous one and the code compiles after each task.
- Writing of tests must be declared as their own tasks
- Test-first: where practical, a task writing a failing test precedes the task that makes it pass.
- Each task references the acceptance criterion (or criteria) it satisfies, e.g. `T3 (AC2)`.
- Keep tasks small enough to complete and review independently.
- Every acceptance criterion must be covered by at least one task.
- Finish with a task to run `make lint` and `make test`.

Stop for review. Do not write code. Do not commit.
