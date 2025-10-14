# Contributing to kstack

Thanks for your interest in contributing! This project is a Go CLI for spinning up local Kubernetes stacks with Helm-based addons.

- Code of Conduct: see CODE_OF_CONDUCT.md
- License: Apache-2.0 (contributions are licensed under the same terms)

## Development quickstart

- Requirements: Go 1.25, Docker, Helm, and a provider CLI (kind or k3d)
- Build: `go build -o kstack ./cmd/kstack`
- Test: `go test ./...`

## Pull requests

1. Fork the repo and create a feature branch.
2. Keep PRs focused and small when possible.
3. Include tests for new behavior; update docs when flags/UX change.
4. Ensure `go test ./...` passes locally.

## Security issues

Please do not open public issues for sensitive reports. See SECURITY.md for instructions.

Thank you for helping improve kstack.
