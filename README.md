# StatusApp Agent

[![Build and Test](https://github.com/statusapp-io/statusapp-agent/actions/workflows/build.yml/badge.svg)](https://github.com/statusapp-io/statusapp-agent/actions/workflows/build.yml)
[![Release](https://github.com/statusapp-io/statusapp-agent/actions/workflows/release.yml/badge.svg)](https://github.com/statusapp-io/statusapp-agent/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/statusapp-io/statusapp-agent)](https://go.dev/)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL%201.1-blue.svg)](LICENSE)

A lightweight, zero-dependency distributed monitoring agent written in Go. It polls the StatusApp API for monitors to check, executes health checks in parallel, and reports results back to the platform.

## Monitor Types

| Type | Description |
|------|-------------|
| **HTTP / HTTPS / WEBSITE** | HTTP status and response checks |
| **API** | API endpoint checks with custom methods, headers, and body |
| **GraphQL** | GraphQL query execution and validation |
| **Keyword** | HTTP response body keyword matching (contains / not contains) |
| **Port** | TCP port connectivity |
| **Ping** | Host reachability (TCP-based ICMP fallback) |
| **DNS** | DNS record resolution (A, AAAA, MX, TXT, NS, CNAME) |
| **SSL Certificate** | Certificate validity and expiry monitoring |

## Quick Start

### Prerequisites

- **Go 1.26+** (building from source)
- **Docker** (containerized deployment)
- A StatusApp account with an agent key

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STATUSAPP_API_URL` | Yes | — | Base URL of the StatusApp API |
| `STATUSAPP_AGENT_KEY` | Yes | — | Agent authentication key |
| `STATUSAPP_INSTANCE_ID` | No | `hostname-PID` | Unique identifier for this agent instance |
| `STATUSAPP_POLL_INTERVAL` | No | `30s` | How often to poll for monitors (minimum `10s`) |
| `STATUSAPP_CONCURRENCY` | No | `5` | Number of parallel checks (1–50) |

> `API_URL` and `AGENT_KEY` are accepted as fallback variable names.

### Run with Docker Compose (Recommended)

```bash
export STATUSAPP_API_URL=https://api.statusapp.io
export STATUSAPP_AGENT_KEY=your-agent-key
docker-compose up -d
```

### Run with Docker

```bash
docker build -t statusapp-agent .

docker run -d --restart unless-stopped \
  -e STATUSAPP_API_URL=https://api.statusapp.io \
  -e STATUSAPP_AGENT_KEY=your-agent-key \
  statusapp-agent
```

### Build and Run from Source

```bash
go build -ldflags="-s -w" -o statusapp-agent .
```

```bash
export STATUSAPP_API_URL=https://api.statusapp.io
export STATUSAPP_AGENT_KEY=your-agent-key
./statusapp-agent
```

## How It Works

1. **Startup** — Loads configuration from environment variables and generates an instance ID if not provided.
2. **Heartbeat** — Sends a heartbeat to the API every 30 seconds so the platform knows the agent is alive.
3. **Poll** — Fetches assigned monitors from the API at the configured interval.
4. **Execute** — Runs checks in parallel (bounded by the concurrency setting) and measures response times.
5. **Report** — Submits results (status, response time, status code, errors) back to the API.
6. **Shutdown** — Handles `SIGINT` and `SIGTERM` for graceful shutdown.

## API Communication

The agent authenticates with the StatusApp API using two headers:

- `X-Agent-Key` — Your agent key
- `X-Instance-Id` — The agent's instance identifier

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/agent/heartbeat` | POST | Report agent liveness |
| `/api/agent/poll` | GET | Fetch monitors to check |
| `/api/agent/results` | POST | Submit check results |

## Project Structure

```
statusapp-agent/
├── main.go              # Agent orchestration, polling, and API communication
├── checks/
│   ├── types.go         # Monitor/Result types and check dispatcher
│   ├── http.go          # HTTP, HTTPS, API, GraphQL, and keyword checks
│   ├── tcp.go           # TCP port connectivity checks
│   ├── ping.go          # Ping/ICMP checks (TCP fallback)
│   ├── dns.go           # DNS record resolution checks
│   └── ssl.go           # SSL certificate validation and expiry checks
├── Dockerfile           # Multi-stage Docker build (Alpine)
├── docker-compose.yml   # Compose deployment configuration
├── go.mod               # Go module (zero external dependencies)
└── LICENSE              # Business Source License 1.1
```

## CI/CD

GitHub Actions workflows are included:

- **build.yml** — Runs `go vet`, `go build`, and `go test` on pushes and PRs to `main`.
- **release.yml** — Builds multi-architecture Docker images (`amd64`/`arm64`) and standalone binaries on version tags (`v*`).

## License

Business Source License 1.1 — free to use with a valid StatusApp subscription. Converts to Apache 2.0 on March 9, 2030. See [LICENSE](LICENSE) for full terms. Contact licensing@statusapp.io for alternative arrangements.

## Contributing

Contributions are welcome! Please read the [Contributing Guide](CONTRIBUTING.md) before opening a pull request.

## Security

To report a vulnerability, see [SECURITY.md](SECURITY.md). **Do not open a public issue.**

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). Please be respectful in all interactions.
