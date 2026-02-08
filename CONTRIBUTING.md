# Contributing to prjct

Thank you for considering contributing to prjct. This document outlines how to get started.

## Development Setup

1. Fork and clone the repository
2. Ensure Go 1.25+ is installed
3. Run `go mod tidy` to install dependencies
4. Run `go test ./...` to verify everything works

## Making Changes

### Branch Naming

- `feature/description` for new features
- `fix/description` for bug fixes
- `docs/description` for documentation changes

### Code Guidelines

- Follow standard Go conventions (`gofmt`, `go vet`)
- Write tests for new functionality
- Keep changes focused and atomic
- Update documentation if behavior changes

### Testing

All changes must pass existing tests:

```bash
go test ./...
```

New features should include tests. Test files follow Go conventions (`*_test.go` alongside source files).

### Commit Messages

Use clear, descriptive commit messages:

```
Add template validation for duplicate IDs

The doctor command now checks for duplicate template IDs
and reports them as validation errors.
```

## Submitting a Pull Request

1. Create a feature branch from `main`
2. Make your changes with tests
3. Ensure `go test ./...` passes
4. Ensure `go vet ./...` reports no issues
5. Push your branch and open a pull request
6. Fill in the PR template

## Reporting Issues

Use the [issue templates](.github/ISSUE_TEMPLATE/) to report bugs or request features. Include:

- OS and Go version for bugs
- Steps to reproduce
- Expected vs actual behavior
- Config file (sanitized) if relevant

## Code of Conduct

Be respectful and constructive. We're all here to build useful software.
