# K0s Binary Existence Check Feature

## Feature Description

`K0rdentd` should check if the binary `k0s` exists. This check will be used in `k0rdentd install` and `k0rdentd uninstall`, and if needed, other commands. If `k0s` is not present, the command `k0rdentd install` or `k0rdentd uninstall` should fail.

## Verification

Verification should be as simple as: `which k0s` gives `/usr/local/bin/k0s`.

- If it does, then `k0s` is available and its version is logged to the terminal, as follows: `Found K0s in version <VERSION>` where `<VERSION>` is extracted from the command `k0s version`.
- If it does not, `k0s` should be installed using the official k0s install script at `https://get.k0s.sh`.

## Implementation

### Files Created

1. **[`pkg/k0s/checker.go`](pkg/k0s/checker.go)** - Main k0s binary checker utility
   - [`CheckK0s()`](pkg/k0s/checker.go:18) - Checks if k0s binary exists and returns its version
   - [`getK0sVersion()`](pkg/k0s/checker.go:45) - Executes `k0s version` and parses the output
   - [`InstallK0s()`](pkg/k0s/checker.go:65) - Installs k0s binary using the official install script

2. **[`pkg/k0s/checker_test.go`](pkg/k0s/checker_test.go)** - Unit tests for the k0s checker
   - Tests for [`CheckK0s()`](pkg/k0s/checker_test.go:11) function
   - Tests for [`getK0sVersion()`](pkg/k0s/checker_test.go:33) function
   - Tests for [`InstallK0s()`](pkg/k0s/checker_test.go:48) function
   - Tests for [`CheckResult`](pkg/k0s/checker_test.go:63) struct

### Files Modified

1. **[`pkg/cli/install.go`](pkg/cli/install.go)** - Integrated k0s check into install command
   - Added k0s binary check before installation
   - Automatically installs k0s if not found

2. **[`pkg/cli/uninstall.go`](pkg/cli/uninstall.go)** - Integrated k0s check into uninstall command
   - Added k0s binary check before uninstallation
   - Fails with error if k0s is not installed

## Usage

### Install Command

When running `k0rdentd install`, the tool will:
1. Check if k0s binary exists
2. If not found, download and install it using the official install script
3. Log the k0s version if found
4. Proceed with K0s and K0rdent installation

### Uninstall Command

When running `k0rdentd uninstall`, the tool will:
1. Check if k0s binary exists
2. Fail with error if k0s is not installed
3. Proceed with uninstallation if k0s is found

## Design Decisions

- **Binary Detection**: Uses `exec.LookPath()` to check if k0s is in PATH
- **Version Parsing**: Parses output from `k0s version` command
- **Installation**: Uses the official k0s install script (`curl -sSf https://get.k0s.sh | sudo sh`) which installs k0s to `/usr/local/bin/k0s`
- **Logging**: Uses logrus with appropriate log levels (DEBUG for intermediate steps, INFO for user-relevant events)
