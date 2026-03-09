# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x     | Yes       |

## Reporting a Vulnerability

We take the security of StatusApp Agent seriously. If you discover a security vulnerability, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email us at **security@statusapp.io** with:

- A description of the vulnerability
- Steps to reproduce the issue
- The potential impact
- Any suggested fixes (optional)

## Response Timeline

- **Acknowledgment**: Within 48 hours of your report
- **Initial assessment**: Within 5 business days
- **Resolution target**: Within 30 days for critical issues

## Disclosure Policy

- We will work with you to understand and resolve the issue before any public disclosure.
- We will credit reporters in release notes (unless you prefer to remain anonymous).
- We ask that you give us reasonable time to address the issue before disclosing it publicly.

## Scope

The following are in scope:

- The StatusApp Agent codebase (this repository)
- Agent-to-API communication security
- Authentication and authorization bypass
- Denial of service vulnerabilities

The following are **out of scope**:

- The StatusApp web application or API (report those separately at security@statusapp.io)
- Social engineering attacks
- Issues in dependencies we don't control

## Best Practices for Deployment

- Always use TLS (`https://`) for your `STATUSAPP_API_URL`
- Keep your `STATUSAPP_AGENT_KEY` secret and rotate it periodically
- Run the agent with minimal OS privileges
- Use the official Docker image or verified builds from GitHub Releases
- Keep the agent updated to the latest version
