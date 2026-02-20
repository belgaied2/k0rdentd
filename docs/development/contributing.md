# Contributing

Thank you for your interest in contributing to K0rdentd!

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make
- Git
- curl (for downloading assets)
- sudo access (for testing)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/k0rdentd.git
cd k0rdentd

# Add upstream remote
git remote add upstream https://github.com/belgaied2/k0rdentd.git
```

### Build

```bash
# Build online flavor
make build

# Build airgap flavor
make build-airgap
```

### Test

```bash
# Run tests
make test

# Run specific test
go test ./pkg/config/... -v
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b bugfix/my-fix
```

### 2. Make Changes

- Write code
- Add tests
- Update documentation

### 3. Test Your Changes

```bash
# Run tests
make test

# Build
make build

# Test locally
sudo ./bin/k0rdentd install
```

### 4. Commit

We follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Tests
- `refactor:` Code refactoring
- `chore:` Maintenance

```bash
git commit -m "feat: add new feature"
```

### 5. Push and Create PR

```bash
git push origin feature/my-feature
```

Then create a Pull Request on GitHub.

## Code Style

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `goimports` for imports

### Logging

Use logrus with appropriate levels:

```go
// DEBUG: Intermediate steps
utils.GetLogger().Debug("Checking k0s status")

// INFO: User-relevant events
utils.GetLogger().Info("K0s installed successfully")

// WARN: Unexpected behavior
utils.GetLogger().Warn("Credential creation failed, continuing")

// ERROR: Critical errors
utils.GetLogger().Error("Failed to start k0s")
```

### Error Handling

```go
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

## Testing

### Unit Tests

Use ginkgo/gomega:

```go
Describe("CheckK0s", func() {
    Context("when k0s is installed", func() {
        It("should return version", func() {
            result, err := CheckK0s()
            Expect(err).NotTo(HaveOccurred())
            Expect(result.Installed).To(BeTrue())
        })
    })
})
```

### Test Coverage

Maintain >80% code coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Documentation

### Code Comments

```go
// CheckK0s checks if k0s is installed and returns its version.
// It returns CheckResult with Installed=true if found, false otherwise.
func CheckK0s() (CheckResult, error) {
    // ...
}
```

### README and Docs

Update documentation for:

- New features
- Configuration changes
- CLI changes
- Bug fixes (if user-facing)

## Pull Request Guidelines

### PR Title

Use conventional commit format:

- `feat: add AWS credential support`
- `fix: correct image tag parsing`
- `docs: update installation guide`

### PR Description

Include:

1. **What**: Description of changes
2. **Why**: Motivation for changes
3. **How**: Implementation details
4. **Testing**: How to test

### Checklist

- [ ] Tests pass
- [ ] Code is formatted
- [ ] Documentation updated
- [ ] Commit messages are clear

## Release Process

Releases are automated via GitHub Actions:

1. Create and push a tag: `git tag v0.3.0 && git push origin v0.3.0`
2. GitHub Actions builds and releases
3. Update documentation

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/belgaied2/k0rdentd/issues)
- **Discussions**: [GitHub Discussions](https://github.com/belgaied2/k0rdentd/discussions)

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
