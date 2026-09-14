<div align="center">

# gh-plt

_Best-practice GitHub Actions workflows, scaffolded for the project you're actually in_

[![CI](https://github.com/SamuelBisberg/gh-plt/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelBisberg/gh-plt/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/SamuelBisberg/gh-plt?sort=semver)](https://github.com/SamuelBisberg/gh-plt/releases)

[Install](#install) • [Usage](#usage) • [Configuration](#configuration) • [How it works](#how-it-works)

</div>

A [GitHub CLI](https://cli.github.com) extension that generates GitHub Actions workflows and composite actions from a curated, context-aware template catalog. gh-plt detects the language and framework in your working directory, suggests templates that actually apply, and pins every `owner/repo@latest` action reference it renders to a real version before writing anything to disk. A single static binary: no Node, no Python, just your existing github extension.

## Install

```sh
gh extension install SamuelBisberg/gh-plt
```

Upgrade later with `gh extension upgrade plt`.

## Usage

### `gh plt add`

Detects (or prompts for) the project's language and framework, narrows the catalog to templates that apply, and lets you pick one.

```console
$ gh plt add
✓ Detected: php (laravel)
? Select a template (detected: php) php/laravel/ci — Run the Laravel test suite with PHPUnit/Pest
✓ Wrote .github/workflows/laravel-ci.yml
```

Pass a catalog path to generate a specific template directly, bypassing suggestions:

```sh
gh plt add php/laravel/ci
```

### `gh plt search`

Browse the full catalog with a filterable, fuzzy picker - templates matching your project's detected language are bubbled to the top, but nothing is filtered out.

```sh
gh plt search
```

### `gh plt config`

Steps through every configurable setting - preferred package manager per language, action versioning mode - one prompt at a time and saves your choices.

```sh
gh plt config
```

> [!TIP]
> Every generated file is confirmed before it's written. Overwriting, appending to, or merging into a file that already exists always defaults to "no" - only creating a brand-new file defaults to "yes".

## Configuration

Settings live in `~/.config/gh/plt.yaml` (or your platform's config dir equivalent) and are managed with `gh plt config`.

| Key                  | Values               | Default                    |                                                                                                                                  |
| -------------------- | -------------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `installers.<lang>`  | package manager name | language's flagged default | Which package manager to install a template's dependencies with, per language.                                                   |
| `actions.versioning` | `tag`, `hash`        | `tag`                      | How resolved `owner/repo@latest` action references are pinned - a semver tag or a commit hash with a version-annotating comment. |

## How it works

- **The template catalog** lives under [`templates/`](templates), compiled into the binary with `go:embed`. Each template is a directory - `templates/<kind>`, `templates/<lang>/<kind>`, or `templates/<lang>/<framework>/<kind>` - holding a `template.yaml` manifest plus the content files it renders. No Go code changes are needed to add, edit, or remove a template.
- **Detection** (`gh plt add` with no path) walks [`templates/ecosystems.yaml`](templates/ecosystems.yaml) looking for marker files (`composer.json`, `package.json`, `pyproject.toml`, framework markers like `artisan`, lockfiles like `pnpm-lock.yaml`) to infer the project's language, framework, and package manager. Anything undetected is prompted for instead.
- **Rendering** runs each template file through Go's `text/template` with your answers to its declared variables, then writes it to disk per that file's strategy: `override` (default) replaces the destination outright, `append` adds to the end of an existing file, and `merge` deep-merges it in as JSON - for a partial file like `package.json` where only some keys are gh-plt's to set.
- **Version resolution** scans rendered content for `owner/repo@latest` placeholders and replaces each with the actual latest release, resolved concurrently through the GitHub API via your `gh auth login` session - never a hardcoded or stale pin.
- **Dependency installation** prompts once per package a template requires, then shells out to the language's configured package manager (`composer`, `npm`/`pnpm`/`yarn`, `uv`/`pip`) with output streamed straight to your terminal.
- The interactive prompts are [charmbracelet/huh](https://github.com/charmbracelet/huh); styling is [lipgloss](https://github.com/charmbracelet/lipgloss). Everything renders in-process - no subprocess.

Releases are built by [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) whenever a version tag is pushed, producing binaries for Linux, macOS, and Windows - that's what `gh extension install`/`upgrade` fetch.

Want to contribute? See [CONTRIBUTING.md](CONTRIBUTING.md).
