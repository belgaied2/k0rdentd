# K0rdentd

K0rdentd is a CLI tool which deploys K0s and K0rdent on the VM it runs on. Its primary goal is to simplify the multi-step installation process `K0rdentd`. The idea is to offer limited options and sane defaults to the user, in order to not overwhelm him. At best the user should be able to install K0rdent with minimal Kubernetes, Helm or even command line knowledge.

## Before Implementing anything

- Always check the [ARCHITECTURE.md](./ARCHITECTURE.md) file for references about the architecture.
- If you ever want to check the codebase, begin by reading [the codebase reference](./AGENT_DOCS/CODEBASE_REFERENCE.md).
- Make sure you follow what is written in the `## Documentation` section below.

## Testing

- Write unit tests for all business logic
- Maintain >80% code coverage
- Use `ginkgo` and `gomega` for Unit tests
- For verifying implementation manually, you can run `go build -o k0rdentd cmd/k0rdentd/main.go` and then `sudo ./k0rdentd -c k0rdentd-config.yaml install` and check the output of the command. After that, you can also use `sudo k0s kubectl` to access the `k0s` cluster if it was created successfully.

## Security

- Never commit API keys or secrets
- Validate all user inputs
- Use parameterized queries for database access

## Terminal Command usage

- Go is installed in `/home/linuxbrew/.linuxbrew/bin/`, this is present in the `$PATH` that is defined in `~/.bashrc`, do not forget to source `~/.bashrc` before running commands.
- Use `/bin/bash` as a shell interpreter.

## Logging

- The library `sirupsen/logrus` shall be used, with the following Log levels
  - for each intermediate step during any kind of processing, a `DEBUG` log will be triggered.
  - for each finished step that is relevant to the final user, such as a status change or a completed task, an `INFO` log should be triggered.
  - for each unexpected behavior happening during a processing, such as an error that is ignored, a `WARN` log should be triggered.

## Documentation

- The codebase (all functions and their description) is documented in @AGENT_DOCS/CODEBASE_REFERENCE.md, use that in conjunction with @ARCHITECTURE.md to understand the codebase. Only if it is still unclear, check the actual code (.go files) in a targetted manner.
- Always update the @AGENT_DOCS/CODEBASE_REFERENCE.md whenever you make changes to the code. This file MUST be always up-to-date.
- Prefer using `context7` when you need library/API documentation, setup steps, architecture decisions or code generation.
- More specifically if you need to do ANYTHING related to :
  - `k0s`: check `k0sproject/k0s` on context7 before generating code or suggesting architecture.
  - `k0rdent`: check all the following `docs.k0rdent.io/latest`, `k0rdent/kcm` and `k0rdent/docs`
- For all other libraries, before suggesting code, use the `resolve-library-id` tool to find the correct documentation context.

### Documentation Directories

- **`./docs/`** - User-facing MkDocs documentation
  - Architecture overview, getting started guides, user guides
  - For end users of k0rdentd

- **`./AGENT_DOCS/`** - Agent-facing documentation
  - Implementation plans, bug reports, feature specifications
  - Read whatever file might be relevant to your task based on the catalog below
  - **IMPORTANT**: When creating new AGENT_DOCS files or making significant changes to existing ones, update this catalog to keep it accurate

  **File Catalog:**

  | File | Type | Description |
  |------|------|-------------|
  | `IMPLEMENTATION_PLAN.md` | Plan | Go module structure and initial directory setup (foundational) |
  | `IMPLEMENTATION_PLAN_CREDENTIALS.md` | Plan | Cloud credentials support implementation (AWS, Azure, OpenStack) |
  | `IMPLEMENTATION_PLAN_AIRGAP.md` | Plan | Airgap installation with OCI registry daemon (multi-phase) |
  | `IMPLEMENTATION_PLAN_SKOPEO_EMBEDDING.md` | Plan | Skopeo binary embedding for airgap builds (COMPLETED) |
  | `IMPLEMENTATION_PLAN_E2E_TESTING.md` | Plan | E2E testing infrastructure using AWS EC2 and Terratest |
  | `FEATURE_check_k0s.md` | Feature | K0s binary existence check and installation logic |
  | `FEATURE_expose_k0rdent-ui.md` | Feature | Exposing k0rdent UI via ingress with IP detection |
  | `FEATURE_Use_clientgo.md` | Feature | Switch from kubectl exec to client-go for K8s operations |
  | `FEATURE_credentials_support.md` | Feature | Cloud provider credentials auto-creation (design spec) |
  | `FEATURE_airgap.md` | Feature | Air-gapped installation with external bundle + OCI registry |
  | `FEATURE_CONTAINERD_MIRROR.md` | Feature | Containerd registry mirror configuration for airgap |
  | `FEATURE_GitHub_Action_Release.md` | Feature | GitHub Actions workflow for releases (online + airgap builds) |
  | `FEATURE_K0RDENT_BUNDLE_CATALOG.md` | Reference | Catalog of k0rdent enterprise airgap bundle contents (images, charts) |
  | `FEATURE_k0s_version_management.md` | Feature | K0s version conflict handling between bundled/config/installed |
  | `FEATURE_multi_node.md` | Feature | Multi-node cluster support (controllers, workers, join tokens) |
  | `BUG_K0rdent_install_check_wrong.md` | Bug | Fixed: K0rdent readiness check using deployments instead of pods |
  | `BUG_Credentials_timing_issue_CAPI_providers_not_ready.md` | Bug | Fixed: Wait for CAPI providers before creating credentials |
  | `BUG_OCI_REGISTRY_PUSH_TAG.md` | Bug | Fixed: OCI image reference tag formatting and path structure |
  | `CODEBASE_REFERENCE.md` | Reference | Complete function/struct documentation by package (always up-to-date) |

- **`./ARCHITECTURE.md`** - High-level architecture reference
  - Always check this for overall architecture context

- **`./TODO.md`** - General TODO tracking

When implementing a new feature or fixing a bug:
1. Read relevant `AGENT_DOCS/` files for context
2. Read `./docs/` for user-facing documentation patterns
3. Update or create `AGENT_DOCS/FEATURE_*.md` or `AGENT_DOCS/BUG_*.md` with design/specifications
4. Update `./docs/` with user-facing changes after implementation
5. Update implementation plans in `AGENT_DOCS/IMPLEMENTATION_PLAN_*.md` as needed