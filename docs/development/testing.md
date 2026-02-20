# Testing

This document describes the testing strategy and guidelines for K0rdentd.

## Testing Framework

We use:

- **Ginkgo**: BDD testing framework
- **Gomega**: Matcher library
- **Fake Clientset**: For mocking Kubernetes API

## Running Tests

### All Tests

```bash
make test
```

### Specific Package

```bash
go test ./pkg/config/... -v
```

### Specific Test

```bash
go test ./pkg/k0s -v -run TestCheckK0s
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Structure

### Unit Tests

Place tests alongside the code:

```
pkg/
├── k0s/
│   ├── checker.go
│   └── checker_test.go
├── config/
│   ├── k0rdentd.go
│   └── k0rdentd_test.go
└── credentials/
    ├── credentials.go
    └── credentials_test.go
```

### Test File Naming

- Unit tests: `<name>_test.go`
- Integration tests: `<name>_integration_test.go`

## Writing Tests

### Basic Test Structure

```go
package k0s_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "github.com/belgaied2/k0rdentd/pkg/k0s"
)

var _ = Describe("CheckK0s", func() {
    Context("when k0s is not installed", func() {
        It("should return Installed=false", func() {
            result, err := k0s.CheckK0s()
            Expect(err).NotTo(HaveOccurred())
            Expect(result.Installed).To(BeFalse())
        })
    })
    
    Context("when k0s is installed", func() {
        BeforeEach(func() {
            // Setup: install mock k0s
        })
        
        AfterEach(func() {
            // Cleanup
        })
        
        It("should return Installed=true and version", func() {
            result, err := k0s.CheckK0s()
            Expect(err).NotTo(HaveOccurred())
            Expect(result.Installed).To(BeTrue())
            Expect(result.Version).NotTo(BeEmpty())
        })
    })
})
```

### Testing Kubernetes Client

Use fake clientset:

```go
import (
    "k8s.io/client-go/kubernetes/fake"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NamespaceExists", func() {
    var (
        client *k8sclient.Client
        fakeClient *fake.Clientset
    )
    
    BeforeEach(func() {
        fakeClient = fake.NewSimpleClientset()
        client = &k8sclient.Client{
            Clientset: fakeClient,
        }
    })
    
    Context("when namespace exists", func() {
        BeforeEach(func() {
            ns := &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{Name: "test"},
            }
            fakeClient.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
        })
        
        It("should return true", func() {
            exists, err := client.NamespaceExists("test")
            Expect(err).NotTo(HaveOccurred())
            Expect(exists).To(BeTrue())
        })
    })
    
    Context("when namespace does not exist", func() {
        It("should return false", func() {
            exists, err := client.NamespaceExists("nonexistent")
            Expect(err).NotTo(HaveOccurred())
            Expect(exists).To(BeFalse())
        })
    })
})
```

### Testing Exec Commands

Mock exec.Command:

```go
// Use a helper function that can be mocked
type CommandExecutor interface {
    Execute(name string, args ...string) ([]byte, error)
}

type RealExecutor struct{}

func (e *RealExecutor) Execute(name string, args ...string) ([]byte, error) {
    cmd := exec.Command(name, args...)
    return cmd.Output()
}

type MockExecutor struct {
    Output []byte
    Error  error
}

func (e *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
    return e.Output, e.Error
}
```

## Test Categories

### Unit Tests

Test individual functions and methods:

- Configuration parsing
- K0s config generation
- Credential creation logic
- Helper functions

### Integration Tests

Test component interactions:

- K0s installation
- K0rdent deployment
- Credential creation
- UI exposure

### End-to-End Tests

Test complete workflows:

- Full installation
- Full uninstallation
- Airgap installation

## Test Coverage

### Requirements

- Maintain >80% code coverage
- Cover all business logic
- Cover error paths

### Checking Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# View summary
go tool cover -func=coverage.out
```

## Mocking

### Kubernetes API

```go
import (
    "k8s.io/client-go/kubernetes/fake"
)

fakeClient := fake.NewSimpleClientset()
```

### File System

```go
import (
    "testing/fstest"
)

memFS := &fstest.MapFS{
    "test.yaml": &fstest.MapFile{
        Data: []byte("content"),
    },
}
```

### Environment Variables

```go
BeforeEach(func() {
    os.Setenv("TEST_VAR", "value")
})

AfterEach(func() {
    os.Unsetenv("TEST_VAR")
})
```

## Best Practices

### 1. Isolate Tests

Each test should be independent:

```go
BeforeEach(func() {
    // Setup
})

AfterEach(func() {
    // Cleanup
})
```

### 2. Use Descriptive Names

```go
It("should return error when config file does not exist", func() {
    // ...
})
```

### 3. Test Edge Cases

```go
Context("when input is empty", func() {
    It("should return error", func() {
        // ...
    })
})

Context("when input is invalid", func() {
    It("should return error", func() {
        // ...
    })
})
```

### 4. Don't Test External Dependencies

Mock external calls:

- Network requests
- File system (when possible)
- External commands

## Continuous Integration

Tests run automatically on:

- Pull requests
- Pushes to main branch

### CI Configuration

Tests run in GitHub Actions:

```yaml
- name: Run tests
  run: make test
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./...
```

### Focus Specific Test

```go
It("should work", func() {
    // ...
})
```

Run with:

```bash
go test -v -focus "should work"
```

### Debug Logging in Tests

```go
import (
    "github.com/sirupsen/logrus"
)

BeforeEach(func() {
    logrus.SetLevel(logrus.DebugLevel)
})
```

## References

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Documentation](https://onsi.github.io/gomega/)
- [Go Testing](https://golang.org/pkg/testing/)
