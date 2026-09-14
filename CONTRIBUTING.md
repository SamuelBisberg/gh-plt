# Contributing to gh-plt

Thanks for taking the time to contribute!

## Prerequisites

- [Go](https://go.dev) 1.25+ (see `go.mod`)
- [git](https://git-scm.com)
- [GitHub CLI](https://cli.github.com) (`gh`), to run the extension locally

## Getting started

```sh
git clone https://github.com/SamuelBisberg/gh-plt.git
cd gh-plt
go build ./...
```

Install it as a local `gh` extension to try your changes end-to-end:

```sh
gh extension install .
gh plt search
```

> [!TIP]
> After changing code, `gh extension remove plt && gh extension install .` picks up a fresh build.

## Making changes

```sh
go build ./...
go vet ./...
gofmt -l .          # should print nothing
go test -race ./...
```

All four must pass before opening a PR - CI runs the same checks.

### Project layout

| Package         | Responsibility                                                                                     |
| ---------------- | --------------------------------------------------------------------------------------------------- |
| `cmd/`           | The `gh plt` command tree ([cobra](https://github.com/spf13/cobra))                                 |
| `pkg/`           | Catalog and ecosystem parsing, project detection, config, rendering, version resolution, prompts, styling |
| `templates/`     | The embedded template catalog and language/framework ecosystem definitions ([`ecosystems.yaml`](templates/ecosystems.yaml)) |

### Adding a template

A template is a directory under `templates/` holding a `template.yaml` manifest plus its content files - no Go code required:

- `templates/<kind>/` for a template that applies regardless of language (e.g. `templates/release/`)
- `templates/<lang>/<kind>/` for a language-specific, framework-less template (e.g. `templates/js/ci/`)
- `templates/<lang>/<framework>/<kind>/` for a framework-specific template (e.g. `templates/php/laravel/ci/`)

Each file listed in the manifest declares its `destination`, an optional write `strategy` (`override`, `append`, or `merge`), any `requires` (dependencies to offer installing first), and any `variables` to prompt for before rendering. See an existing `template.yaml` for the shape.

### Tests

`pkg`'s tests exercise the catalog parser, ecosystem detection, config load/save, and file-writing strategies (`mergedContent`) directly against temp directories and the embedded catalog - no mocking of the filesystem or `gh` itself. Follow that pattern for new catalog- or detection-related behavior.

> [!WARNING]
> `pkg.ResolveVersions` calls the live GitHub API through `gh`'s authenticated client. Tests that exercise version resolution need real network access and a `gh auth login` session - keep that behavior isolated rather than pulling it into unrelated tests.

## Submitting a pull request

1. Fork the repo and create a branch off `main`.
2. Make your change, with tests.
3. Run the checks above.
4. Open a PR - the template will guide you through the rest.

## Reporting bugs / requesting features

Please use the issue templates; they ask for just enough context to act on a report quickly.
