# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 2.x     | :white_check_mark: |
| 1.x     | :x:                |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Preferred method:** Use GitHub's private vulnerability reporting:

[https://github.com/lfaoro/ssm/security/advisories/new](https://github.com/lfaoro/ssm/security/advisories/new)

You can also email **ssm@leonardofaoro.com** if you prefer not to use the GitHub form.

- Please **do not** open a public issue for security vulnerabilities.
- Include steps to reproduce, affected versions, and any relevant details.
- We aim to acknowledge reports within 48 hours.

## Security Model & Hardening

SSM is built for people who operate production infrastructure through SSH. Its design follows these core principles:

- No software or agents are ever installed on remote servers
- All connections use hardened SSH flags (`BatchMode=yes`, `RequestTTY=no`)
- The `--` delimiter is used before hostnames in all SSH, mosh, and `syscall.Exec` invocations to prevent command injection
- Sensitive keys (`IdentityFile`, `ProxyCommand`, etc.) are filtered from the configuration inspector
- SSH config file permissions are checked on load with warnings for insecure modes
- Remote command output is sanitized (ANSI stripped, stderr truncated)

For the complete list of implemented security measures, see the [Security section in AGENTS.md](AGENTS.md#security).

## Scope

We consider the following in scope for security reports:

- Remote code execution via the TUI or SSH invocation paths
- Information disclosure of sensitive keys or configuration data
- Authentication or authorization bypasses

## Coordinated Disclosure

We follow responsible disclosure practices. We will work with reporters on a reasonable timeline before any public disclosure and will credit security researchers unless they prefer to remain anonymous.
