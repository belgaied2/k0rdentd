# K0rdentd

K0rdentd is a CLI tool which deploys K0s and K0rdent on the VM it runs on. It follows a similar pattern to [RancherD](https://github.com/harvester/rancherd) which installs RKE2 and Rancher in one go. 

## Architecture

- Use urfave/cli to handle CLI features
- Accept a config file in the same manner as rancherd, the config file should be in YAML format and offer configuration options for k0s and for k0rdent. Its default location should be `/etc/k0rdentd/k0rdentd.yaml` but it should be configurable using CLI flags or environment variables.
- RancherD relies on a `Plan` mechanism that relies on an external component called `Upgrade Controller`, we don't want that here, we want to simply call `k0s` binary and configure it with `/etc/k0s/k0s.yaml` based on the content of the `/etc/k0rdentd/k0rdentd.yaml` file.
- `k0rdent` should be installed using the k0s addon mechanism (under `.spec.extensions.helm` of the `k0s.yaml`, more information avaible [here](https://docs.k0sproject.io/stable/helm-charts/)).
- Always check the [ARCHITECTURE.md](./ARCHITECTURE.md) file for references about the architecture.
- Make sure that you read all docuement under ./docs before design and implementing anything new.
- under [spinner.go](./pkg/utils/spinner.go), there is a spinner that is implemented for re-use whenever a waiting loop is used for checking something repeatedly. Make sure to re-use it whenever you think of implementing a waiting mechanism.

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

- Always use the Context7 MCP server when you need use library/API documentation, setup steps, architecture decisions or code generation. 
- More specifically if you need to do ANYTHING related to :
  - `k0s`: check `k0sproject/k0s` on context7 before generating code or suggesting architecture.
  - `k0rdent`: check all the following `docs.k0rdent.io/latest`, `k0rdent/kcm` and `k0rdent/docs`
- For all other libraries, before suggesting code, use the `resolve-library-id` tool to find the correct documentation context.
- Whenever you implement a new feature, document its design choices in the `./docs` directory. Also, always read all files in the `./docs` directory to understand previous design choices.