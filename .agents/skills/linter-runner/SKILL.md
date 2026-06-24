---
name: linter-runner
description: Execute repository linters and surface highest-priority issues with minimal, targeted fixes.
license: MIT
metadata:
  mode: verify
  purpose: lint
---

# Linter Runner

## When to Use

- After formatting and before/after tests or PR creation to ensure linter parity with CI.

## Rules

- Use the repository's configured linter commands (see `AGENTS.md`).
- Prioritize actionable issues: security/correctness > concurrency > API misuse > style.
- Provide minimal code-change suggestions; avoid large refactors unless explicitly requested.

## Commands

- **Primary:** `golangci-lint run --timeout 5m --verbose ./...` (CI uses golangci-lint v2.12.2)
- **Secondary:** `revive ./...`

## Output

- Top findings (5-15) grouped by file and rule.
- One-line fix suggestion per finding.
- Verification steps (re-run lint, run focused tests).

## Related Skills

- `code-formatter`, `static-analysis`, `test-runner`
