# contributing to lime

Thank you for your interest in contributing! lime follows standard Go project conventions.

## Prerequisites

- Go 1.24+
- `golangci-lint` — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `make`

## Development workflow

```bash
# Clone
git clone https://github.com/freeoss-space/lime
cd lime

# Run tests
make test

# Run tests with race detector
make test-race

# Run linter
make lint

# Build binary
make build

# Generate coverage report
make coverage-html
```

## TDD requirement

All new functionality must follow red-green-refactor:

1. Write a failing test that describes the desired behaviour.
2. Write the minimal implementation to make the test pass.
3. Refactor for clarity and maintainability.
4. Commit each logical step separately.

## Code style

- Run `gofmt -s` before committing.
- Follow `golangci-lint` recommendations.
- No magic numbers — use named constants.
- Write table-driven tests where applicable.
- Keep functions small and single-purpose.

## Adding a package manager

1. Create `internal/managers/<name>.go` implementing `PackageManager`.
2. Add it to `managers.DefaultRegistry()` in `registry.go`.
3. Add the repo mapping(s) to `repology.repoFamilyMap` in `types.go`.
4. Write tests in `internal/managers/managers_test.go`.

## Submitting changes

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature/my-feature`.
3. Make changes following TDD.
4. Ensure `make test lint` passes.
5. Open a pull request describing the change and motivation.

## Repology API guidelines

- Do not increase `rate_limit_per_second` beyond 1 in production code.
- Cache responses where reasonable.
- Include a descriptive `User-Agent` header (already configured).
- See https://repology.org/api for full guidelines.

## License

By contributing, you agree that your contributions will be licensed under the AGPL-3.0 License.
