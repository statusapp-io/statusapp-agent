# Contributing to StatusApp Agent

Thank you for your interest in contributing to StatusApp Agent. This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- Go 1.26 or later
- Docker (optional, for container builds)
- Git

### Development Setup

```bash
# Clone the repository
git clone https://github.com/statusapp-io/statusapp-agent.git
cd statusapp-agent

# Build
go build -o statusapp-agent .

# Run tests
go test ./...

# Run linter
go vet ./...
```

## How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates. When filing a bug, use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml) and include:

- Your Go version (`go version`)
- Your OS and architecture
- Steps to reproduce the issue
- Expected vs actual behavior
- Relevant log output

### Suggesting Features

Feature requests are welcome. Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.yml) and describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Submitting Changes

1. Fork the repository
2. Create a feature branch from `main` (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass (`go test ./...`)
6. Ensure code passes vet (`go vet ./...`)
7. Commit your changes with a clear commit message
8. Push to your fork and open a Pull Request

### Adding a New Check Type

To add a new monitor check type:

1. Create a new file in `checks/` (e.g., `checks/mycheck.go`)
2. Implement a function with the signature `func runMyCheck(m Monitor) Result`
3. Register the check type in the `Run` dispatcher in `checks/types.go`
4. Add tests in `checks/mycheck_test.go`

## Coding Standards

- Follow standard Go conventions and [Effective Go](https://go.dev/doc/effective-go)
- Use `go vet` and `go fmt` before committing
- Keep the zero-dependency philosophy — avoid adding external dependencies unless absolutely necessary
- Write clear, descriptive commit messages
- Keep functions focused and small

## Commit Messages

Use clear, concise commit messages:

```
Add DNS CNAME record validation

Check CNAME records during DNS monitor checks and validate
against the expected value if configured.
```

- Use the imperative mood ("Add feature" not "Added feature")
- First line should be under 72 characters
- Add a blank line before any detailed description

## Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation if your change affects user-facing behavior
- All CI checks must pass before merging
- PRs require at least one review approval

## License

By contributing to StatusApp Agent, you agree that your contributions will be licensed under the same [Business Source License 1.1](LICENSE) that covers the project.
