# Specs

Feature specifications for spec-driven development with Claude Code.

## Layout

Each feature lives in ``:

- `spec.md` — what and why (requirements, acceptance criteria). Source of truth for behaviour.
- `plan.md` — how (design, affected packages, test strategy).
- `tasks.md` — ordered, test-first checklist.

Templates are in `_templates`.

## Branches

Each spec is worked on in a branch named `spec/NNN-kebab-name`, matching its folder.
The Claude skills detect the active spec from the current branch; an explicit spec id argument overrides it.

## Workflow

| Step | Command | Output | Stop for review |
|------|---------|--------|-----------------|
| 1 | `/spec <description>` | Creates folder + branch, writes `spec.md` (Draft) | Answer clarifications, set Status to `Approved` |
| 2 | `/spec-plan [id]` | `plan.md` | Review design |
| 3 | `/spec-tasks [id]` | `tasks.md` | Review task breakdown |
| 4 | `/spec-implement [id] [task]` | Code + tests, ticks tasks | Review PR |

Bug fixes, dependency bumps and pure refactors do not need a spec.
