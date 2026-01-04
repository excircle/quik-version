# Quik Version

A Golang application to version and build containers. 

Quik Version is built using the following tools:

- [Cobra](https://github.com/spf13/cobra)
- [Viper](https://github.com/spf13/viper)
- [go-github](https://github.com/google/go-github)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)
- [go-yaml](https://github.com/go-yaml/yaml)
- [buildah](https://pkg.go.dev/github.com/containers/buildah)

Quik Version followings the versioning guidelines specified in the locally available `semver.md`

# How Does Quik Version Work

Quik Version uses binary called `qv`. Below is a basic workflow for using Quik Version

- `qv init` initializes the application by doing the following
    - Checks for existence of `quik.conf`
    - Checks for the existence of `qv.db`
    - Prompts user to create these files (if not exists)
- `qv vet` checks `quik.conf` for `git_url`
    - Checks if a tag and version have been applied to latest 'main' version
    - Checks if `qv.db` reflects current information, and offers options to reconcile if mismatching
- `qv status` reports the latest versioning data from `qv.db`
- `qv plan` reads last commit history and takes the following programmatic logic
    - Assumes that you will increment version by a minor version
    - A `--major` or `--patch` flag is required to increment anything other than minor
    - Creates an execution plan in `plan.yaml`
- `qv deploy` reads `plan.yaml` and deploys based on config settings inside `quik.conf`

# Fully Qualified `quick.conf` File

```yaml
version:
    git_url: https://github.com/excircle/scratch-app
build:
    build_management: false
```

# Fully Qualified `plan.yaml` File

```yaml
git_url: https://github.com/excircle/scratch-app
version: 0.5.0
```