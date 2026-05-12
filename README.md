# lime — Let's Install My Ecosystem

[![CI](https://github.com/freeoss-space/lime/actions/workflows/ci.yml/badge.svg)](https://github.com/freeoss-space/lime/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/freeoss-space/lime.svg)](https://pkg.go.dev/github.com/freeoss-space/lime)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)

`lime` is a cross-platform package installer and search abstraction layer.
You say **what** to install; lime figures out **how** based on what's available on your system.

```
$ lime install ripgrep

Install package?

  Manager : apt
  Package : ripgrep
  Command : sudo apt-get install -y ripgrep

[y/N] y
✓ installed ripgrep via apt
```

---

## Why lime?

Different systems have different package managers. Shell scripts that install
software must special-case `brew`, `apt`, `dnf`, `pacman`… lime eliminates
that boilerplate.

- **One command across distros** — `lime install ripgrep` works on Ubuntu, Arch, Fedora, macOS, Windows.
- **Smart fallback** — if your preferred manager doesn't have a package, lime tries the next one.
- **Repology-powered** — uses [Repology](https://repology.org) to find the correct package name per manager.
- **Non-interactive friendly** — use `-y` and `--dry-run` in scripts and CI.

---

## Features

- Search packages across all package managers via Repology
- Install with the best available manager, with ordered preference
- Graceful fallback when a manager is unavailable
- Dry-run mode (`--dry-run`) to preview commands
- JSON output (`--json`) for scripting
- Rate-limited, retry-capable Repology API client
- XDG-compliant config (`~/.config/lime/config.toml`)
- Colored output (respects `NO_COLOR`)
- Cross-platform: Linux, macOS, Windows

---

## Supported package managers

| Manager   | Platform          | Binary    |
|-----------|-------------------|-----------|
| `brew`    | macOS / Linux     | `brew`    |
| `apt`     | Debian / Ubuntu   | `apt-get` |
| `dnf`     | Fedora / RHEL     | `dnf`     |
| `pacman`  | Arch Linux        | `pacman`  |
| `zypper`  | openSUSE          | `zypper`  |
| `apk`     | Alpine Linux      | `apk`     |
| `pkg`     | FreeBSD           | `pkg`     |
| `winget`  | Windows           | `winget`  |
| `choco`   | Windows           | `choco`   |
| `scoop`   | Windows           | `scoop`   |

---

## Installation

### From source (recommended)

```bash
go install github.com/freeoss-space/lime/cmd/lime@latest
```

### Build locally

```bash
git clone https://github.com/freeoss-space/lime
cd lime
make build
# binary at: bin/lime
```

### Download release binary

Pre-built binaries for Linux, macOS, and Windows are available on the
[Releases](https://github.com/freeoss-space/lime/releases) page.

---

## Usage

### Install packages

```bash
# Install a single package
lime install ripgrep

# Install multiple packages
lime install ripgrep fd bat

# Skip confirmation prompt
lime install -y ripgrep

# Force a specific manager
lime install --manager brew ripgrep

# Preview without executing
lime install --dry-run ripgrep
```

### Search packages

```bash
# Search Repology for packages matching "ripgrep"
lime search ripgrep

# JSON output for scripting
lime search --json ripgrep
```

Example search output:

```
MANAGER  PACKAGE   VERSION    REPO           AVAIL
apt      ripgrep   14.0.3-1   debian_stable  ✓
brew     ripgrep   14.0.3     homebrew
pacman   ripgrep   14.0.3-1   arch
dnf      ripgrep   14.0.3     fedora_40
```

### Config

```bash
# Open config in $EDITOR
lime config

# Print config file path
lime config path
```

### Global flags

| Flag         | Description                          |
|--------------|--------------------------------------|
| `--dry-run`  | Print commands without executing     |
| `--json`     | Output results as JSON               |
| `-v`         | Verbose output                       |
| `--version`  | Show version information             |

---

## Configuration

Config file location:

| Platform      | Path                                    |
|---------------|-----------------------------------------|
| Linux / macOS | `$XDG_CONFIG_HOME/lime/config.toml`      |
|               | `~/.config/lime/config.toml` (fallback)  |
| Windows       | `%APPDATA%\lime\config.toml`             |

Example config:

```toml
# Ordered list of preferred package managers.
preferred_managers = [
  "brew",
  "apt",
  "dnf",
  "pacman",
  "apk",
]

# Skip the confirmation prompt (equivalent to always passing -y).
auto_confirm = false

# Repology API rate limit (requests per second). Do not exceed 1.
rate_limit_per_second = 1

# Per-request HTTP timeout in seconds.
http_timeout_seconds = 10
```

See [`examples/config.toml`](examples/config.toml) for a fully annotated example.

---

## Repology

lime uses the [Repology API](https://repology.org/api) to resolve package names
across repositories. Repology aggregates package metadata from hundreds of
repositories worldwide.

**API usage:**
- lime sends a custom `User-Agent` header: `lime/<version> (https://github.com/freeoss-space/lime)`
- Rate limiting is enforced client-side (default: 1 request/second)
- Failed requests are retried with exponential backoff
- lime respects the [Repology API guidelines](https://repology.org/api)

If you find lime useful, consider [supporting Repology](https://repology.org/donate).

---

## Architecture

```
cmd/lime/          — main entry point (ldflags version injection)
internal/
  cli/            — Cobra commands (install, search, config)
  config/         — XDG config loading/saving (TOML)
  repology/       — Repology API client with types
  managers/       — PackageManager interface + 10 implementations
  install/        — install orchestration (fallback, confirmation, dry-run)
  search/         — search orchestration (annotation, sorting)
pkg/
  httpclient/     — rate-limited, retry-capable HTTP client
tests/            — integration tests
```

Key design principles:
- **Dependency injection** — all managers accept a `Commander` interface, enabling test doubles without subprocess calls.
- **Interface-driven** — `PackageManager`, `Commander`, `RepologyClient` are all injectable interfaces.
- **Business logic isolated from CLI** — `install` and `search` packages are fully testable without Cobra.
- **TDD** — every package was written test-first (red → green → refactor).

---

## Development

```bash
# Run tests
make test

# Run tests with race detector
make test-race

# Coverage report
make coverage

# HTML coverage
make coverage-html

# Lint
make lint

# Format
make fmt

# Build
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

---

## License

AGPL-3.0 — see [LICENSE](LICENSE).
