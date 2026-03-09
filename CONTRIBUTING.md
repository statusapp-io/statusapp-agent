# Contributing to StatusApp Agent

Thank you for your interest in contributing! This guide will help you get started.

## Getting Started

### Prerequisites

- Go 1.26 or later
- Docker (optional, for container testing)

### Local Development

```bash
git clone https://github.com/statusapp-io/statusapp-agent.git
cd statusapp-agent
go build -o statusapp-agent .
go test ./...
```

### Running Locally

```bash
export STATUSAPP_API_URL=https://api.statusapp.io
export STATUSAPP_AGENT_KEY=your-agent-key
./statusapp-agent
```

## Making Changes

1. Fork the repository
2. Create a feature branch from `main` (`git checkout -b feature/your-change`)
3. Make your changes
4. Run tests: `go test ./...`
5. Run the linter: `go vet ./...`
6. Commit with a clear message
7. Push to your fork and open a pull request

## Code Guidelines

- **No external dependencies.** This project uses only the Go standard library. Do not add third-party modules.
- **Keep it simple.** The agent is intentionally minimal and lightweight.
- **Write tests** for new check types or changes to existing checks.
- **Follow Go conventions** — use `gofmt` and `go vet` before committing.

## Adding a New Check Type

1. Create a new file in `checks/` (e.g., `checks/yourcheck.go`)
2. Implement a function matching the signature: `func runYourCheck(m Monitor) Result`
3. Register it in the `Run()` dispatcher in `checks/types.go`
4. Add tests in `checks/yourcheck_test.go`

## Pull Requests

- Keep PRs focused — one feature or fix per PR
- Reference any related issues
- Ensure CI passes before requesting review
- Fill out the PR template

## Reporting Bugs

Use the [bug report template](https://github.com/statusapp-io/statusapp-agent/issues/new?template=bug_report.yml) on GitHub Issues.

## Security Issues

Do **not** open a public issue for security vulnerabilities. See [SECURITY.md](SECURITY.md) for responsible disclosure instructions.

## License

By contributing, you agree that your contributions will be licensed under the same [Business Source License 1.1](LICENSE) that covers this project.
