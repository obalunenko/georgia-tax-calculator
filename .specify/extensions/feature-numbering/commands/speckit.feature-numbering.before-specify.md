---
description: "Bind specify to the next exact NNN-feature-name directory"
---

# Resolve Feature Directory

This command is the mandatory `before_specify` hook for both stock specify and
Companion specify. It resolves the target directory only; the caller remains
responsible for creating the directory and specification files.

## Input

Use the feature description from the triggering specify request. Hook dispatch
may leave `$ARGUMENTS` empty, so recover the description from the current
conversation instead of asking the user to repeat it.

If the request explicitly supplies `SPECIFY_FEATURE_DIRECTORY` and that target
already contains `spec.md`, treat the operation as an update: return that path
unchanged and do not allocate a new number.

## New-feature naming

For a new feature:

1. Derive a concise 2–4 word feature name from the description.
2. Normalize it to lowercase ASCII kebab-case containing only `a-z`, `0-9`, and
   single hyphens. Preserve meaningful technical terms and acronyms after
   lowercasing.
3. Run exactly once from the repository root:

   ```bash
   python3 .specify/extensions/feature-numbering/scripts/next-feature.py --slug "<feature-name>"
   ```

The script inspects only immediate `specs/` child directories whose complete
basename matches `NNN-<feature-name>`, takes the highest base-10 `NNN`, adds
exactly 1, and returns JSON. Non-matching directories do not affect numbering.
The first number is `001`; allocation stops with an error after `999`.

## Binding handoff

Return the script's JSON unchanged. `SPECIFY_FEATURE_DIRECTORY` in that result
is binding for the current specify invocation: the caller MUST treat it as the
explicit target path, MUST NOT recompute or rename it, and MUST create the spec
there. This handoff applies equally to `speckit-specify` and
`speckit-companion-specify` because both dispatch `before_specify`.

Do not create the directory, write `.specify/feature.json`, create `spec.md`, or
switch branches here. Those remain the caller's responsibilities.
