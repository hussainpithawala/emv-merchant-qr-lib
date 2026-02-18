# Contributing to emv-merchant-qr-lib

Thank you for considering contributing! Here's what you need to know.

## Getting Started

1. **Fork** the repository and clone your fork.
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes, add tests, and ensure everything passes.
4. Push and open a pull request against `main`.

## Development Setup

```bash
git clone https://github.com/hussainpithawala/emv-merchant-qr-lib.git
cd emv-merchant-emvqr
go mod download
go test ./...
```

## Code Style

- Run `gofmt -w .` before committing.
- Run `go vet ./...` and fix any issues.
- Follow standard Go conventions; document all exported symbols.
- Keep the zero-dependency policy — the library must import only the Go standard library.

## Testing

All changes must include tests. Run the full suite with race detection:

```bash
go test -race -count=1 ./...
```

For coverage:

```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Spec Compliance

This library implements EMV QRCPS Merchant-Presented Mode v1.0. Any new feature
or bug fix must reference the relevant section of the spec in the PR description.
Behaviour that deviates from the spec requires explicit justification.

## Pull Request Guidelines

- Keep PRs focused — one concern per PR.
- Reference any relevant issue numbers in the PR description.
- Ensure CI passes before requesting review.
- Update `CHANGELOG.md` under the `[Unreleased]` section.

## Reporting Issues

Please include:
- The raw QR string that caused the problem (redact sensitive data).
- The Go version (`go version`).
- The expected vs. actual behaviour.

## License

By contributing, you agree that your contributions will be licensed under the
project's [MIT License](LICENSE).
