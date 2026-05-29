---
name: Bug report
about: Report a reproducible problem in gorest
title: '[bug] '
labels: ''
assignees: ''

---

## Summary

Describe the bug clearly and concisely.

## Affected area

- [ ] config
- [ ] database / migrate
- [ ] controller / handler
- [ ] middleware
- [ ] service
- [ ] example / example2
- [ ] documentation
- [ ] other

## Reproduction

Provide a minimal reproduction.

1. Configuration or environment used
2. Exact commands run
3. Request, input, or code sample
4. Actual result

## Minimal code or request sample

```go
// Paste a minimal reproduction here, or replace with HTTP request / curl output.
```

## Expected behavior

Describe what should have happened instead.

## Logs or error output

```text
Paste panic output, logs, stack trace, or failing test output here.
```

## Environment

- gorest version or commit SHA:
- Go version:
- OS and architecture:
- Database(s) involved: RDBMS / Redis / MongoDB / none
- Relevant env flags: `ACTIVATE_RDBMS`, `ACTIVATE_REDIS`, `ACTIVATE_MONGO`, `ACTIVATE_JWT`, etc.

## Verification already attempted

- [ ] I searched existing issues first
- [ ] I tested against the latest `main` branch or latest release
- [ ] I included exact reproduction steps
- [ ] I included relevant config values with secrets removed

## Additional context

Add anything else that may help reproduce or diagnose the issue.
