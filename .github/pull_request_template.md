---
name: Pull request
about: Send a pull request to improve the project
title: ''
labels: ''
assignees: ''

---

## Summary

Describe the change and why it is needed.

Fixes # (issue)

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Refactor / internal cleanup
- [ ] Documentation update
- [ ] Test-only change

## What changed

- [ ] Code
- [ ] Configuration
- [ ] Documentation
- [ ] Tests

## How to test

List the exact commands used.

```bash
# Example
go test -v -cover ./...
```

## Verification

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./...`
- [ ] `revive ./...`
- [ ] `golangci-lint run ./...`
- [ ] docs updated if needed

## Breaking changes

- [ ] No
- [ ] Yes

If yes, describe the API, config, or behavior change.

## Checklist

- [ ] My PR targets the `main` branch
- [ ] My changes are limited to a single logical task
- [ ] I tested the change locally
- [ ] I added or updated tests when appropriate
- [ ] I updated `README.md` or other docs when behavior or public usage changed
- [ ] I did not include secrets or generated artifacts unintentionally
