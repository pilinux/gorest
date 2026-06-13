# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Authoritative references

This repo already ships detailed agent docs. Read these before making changes — they
are the source of truth and far more complete than this file:

- `AGENTS.md` — build/lint/test commands, code style, import aliases, layer
  responsibilities, controller↔handler 1:1 map, config/db/middleware/service APIs.
- `llms.txt` — machine-readable interface describing repo purpose, capabilities, and constraints.
- `.agents/skills/<name>/SKILL.md` — task-specific workflow guides (test-runner,
  ci-orchestrator, migration-helper, config-loader-helper, etc.).

## Files to NEVER read or modify

Per `.gitignore` (and `AGENTS.md`), these contain secrets, keys, or generated noise.
Do not open, read, or edit them:

- `.env` and any `*.log` (includes `log.txt`), `*.bak`
- `keys/`, `tmp/`, `.build/`, `.vscode/`, `vendor/`
- `coverage.txt`, `coverage_0.txt`
- `pri*key*.pem`, `pub*key*.pem`
- `crosscompile*.sh`

## Common commands

```bash
source setTestEnv.sh                           # required before running tests locally
go test -v -cover ./...                        # all tests
go test -v -run TestHashPass ./lib/...         # single test function
golangci-lint run ./...                        # lint (CI uses v2.8.0, --timeout 5m)
go vet -v ./... && go build -v ./...           # vet + build
```

CI additionally runs `gosec`, `govulncheck`, and cross-platform `go vet`/build across
linux/darwin/windows × amd64/arm64. Tests need env vars (CI uses secrets;
locally `setTestEnv.sh`).

## Architecture (the big picture)

gorest is a reusable Go module (`github.com/pilinux/gorest`), not just an app. It is
both a library of auth/crypto/middleware/db primitives **and** two example apps that
wire them together. Requires Go 1.24.1+ (1.25.0+ for 1.12.x).

Request flow follows a strict layered pipeline:

```txt
controller/  → thin: bind request, call handler, renderer.Render(c, data, status)
handler/     → business logic + validation; returns (model.HTTPResponse, int)
service/     → shared utilities: auth, JWT blacklist, email (Postmark), crypto, 2FA
database/    → connection management for RDBMS (GORM), Redis (radix), MongoDB
config/      → loads .env → Configuration struct; IsX() feature toggles gate everything
```

Two things make this codebase navigable:

1. **Controller↔Handler 1:1 mapping.** Every exported `controller` function is a thin
   wrapper over the identically named `handler` function (19 pairs: registration, login,
   JWT refresh/logout, password recovery, 2FA lifecycle, email verification/change). See
   the full table in `AGENTS.md`. Handlers never touch `gin.Context`; controllers never
   contain business logic.

2. **Feature toggles drive behavior.** `config.IsRDBMS()`, `IsRedis()`, `IsMongo()`,
   `IsJWT()`, `Is2FA()`, `IsCipher()`, `IsEmailVerificationService()`, etc. are checked
   throughout to conditionally enable subsystems. A feature being "off" in config means
   its code paths are skipped, so reproducing behavior requires matching env config.

`example/` is the legacy app; `example2/` is the recommended interface-driven app
(adds a `repo/` repository layer with DI on top of the same library).

### Conventions that matter

- Import gorest packages with `g`-prefixed aliases (`gconfig`, `ghandler`,
  `gmiddleware`, `glib`, `gservice`, ...); logrus as `log`. Full list in `AGENTS.md`.
- Handlers use named returns with bare `return` for early exits.
- Errors: `log.WithError(err).Error("error code: XXXX.X")` with numbered codes; never
  expose internal errors to API consumers — return user-facing text via `httpResponse.Message`.
- Tests: external test packages (`package lib_test`), table-driven with named structs,
  `t.Run` subtests, stdlib `t.Errorf` (no assertion library).
- Sensitive model fields are hidden with `json:"-"`; non-DB fields with `gorm:"-"`.
  Email is stored encrypted (cipher/nonce/hash columns), not plaintext.

## Contributing

PRs target `main`, one commit per task, test before committing, document new features in `README.md`.
