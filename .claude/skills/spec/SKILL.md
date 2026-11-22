---
name: spec
description: Create a new feature spec in specs/ and a matching spec/ branch
argument-hint: <feature description>
disable-model-invocation: true
allowed-tools: Bash(git branch:*), Bash(git status:*), Bash(git switch:*), Bash(ls:*), Read, Write, Edit
---
Create a feature spec for: $ARGUMENTS

Current branch: !`git branch --show-current`
Working tree: !`git status --short`
Existing specs: !`ls -1 docs/specs`

Steps:

1. Choose the next spec number: the highest `NNN` among existing spec folders plus one, zero-padded to 3 digits (start at 001). Choose a short kebab-case name (2–4 words) from the description.
2. Branch:
   - If the working tree has changes, stop and notify me.
   - If the current branch is not `main`, stop, notify me, and ask me whether I still want to proceed.
   - Otherwise, run `git switch -c spec/NNN-name`.
3. Create `../../../docs/specs` from `../../../docs/specs` and fill it in.
   - Describe WHAT and WHY only. No package names, types, or implementation details.
   - Number every requirement (R1, R2, ...) and every acceptance criterion (AC1, ...), and reference the requirement each AC verifies.
   - Every AC must be objectively testable.
   - Read the relevant parts of the codebase and README only as far as needed to understand current behaviour.
4. Where anything is ambiguous, add a `[NEEDS CLARIFICATION]` item under Open questions rather than guessing.
5. Leave `Status: Draft`.
6. Summarise the spec and ask me the clarification questions. When I answer, update the spec and remove the resolved markers.

Do not write a plan, tasks, or any code. Do not commit.
