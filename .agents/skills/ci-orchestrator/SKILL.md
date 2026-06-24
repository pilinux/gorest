---
name: ci-orchestrator
description: Run a CI-like pipeline locally (format, lint, vet, static-analysis, tests) and summarize per-step results with remediation guidance.
license: MIT
metadata:
  mode: verify
  purpose: ci
---

# CI Orchestrator

## When to Use

- The user wants a single command-style check that reproduces CI checks locally or validates a PR prior to submission.

## Responsibilities

- Execute repository-standard steps in order and summarize failures with actionable fixes.
- Provide per-step outputs (pass/fail), first error lines, and suggested remediations.

## Pipeline Steps

1. **Format:** `go fmt ./...`
2. **Lint:** `golangci-lint run --timeout 5m --verbose ./...` (CI uses v2.12.2)
3. **Lint (revive):** `revive ./...`
4. **Vet:** `go vet -v ./...`
5. **Vet (cross-platform):** `GOOS=linux GOARCH=amd64 go vet -v ./...` and `GOOS=darwin GOARCH=arm64 go vet -v ./...` (CI runs all six OS/arch combos)
6. **Security:** `gosec ./...`
7. **Vulnerability:** `govulncheck ./...`
8. **Tests:** `source setTestEnv.sh && go test -v -cover ./...`

## Rules

- Do not change code automatically; return clear next steps and small fix suggestions.
- Respect the repo CI order as configured in `.github/workflows/`.
- Report per-step status before moving to the next step.

## Output

- Step-by-step status table (step, pass/fail, key error lines).
- Short remediation for each failing step.
- Recommended next commands to re-run after fixes.

## Related Skills

- `linter-runner`, `static-analysis`, `test-runner`, `code-formatter`
