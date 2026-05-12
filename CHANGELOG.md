# Changelog

All notable changes to lime are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
lime uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- **`uv` package manager** — installs Python tools via `uv tool install` (fast, isolated environments); uses PEP 508 `pkg==version` syntax for versioned installs; searches via `uv pip index versions`; preferred over `pip` by default
- **`brew-cask` package manager** — installs macOS GUI applications via `brew install --cask`; searches via `brew search --casks`; shares the `brew` binary so availability mirrors Homebrew
- **Language ecosystem managers**: `pip` (PyPI via `pip3 install --user`), `cargo` (crates.io), `npm` (npm registry via `npm install -g`), `go` (Go modules via `go install`)
- Version-specific install support (`--version` / `pkg==version` / `pkg@version` depending on manager) for all managers that support it
- Cooldown windows — configurable per-manager or global minimum time between installs of the same package (`default_cooldown`, `manager_cooldowns` in config)
- Initial release of `lime`
- `lime install` — install packages via the best available package manager
- `lime search` — search packages across managers via Repology API
- `lime config` — open config file in `$EDITOR`
- `lime config path` — print config file path
- XDG Base Directory spec compliance for config location
- Rate-limited, retry-capable HTTP client (Repology API)
- `--dry-run` flag — shows commands without executing
- `--json` flag — machine-readable output
- `-y` / `--yes` flag — skip confirmation prompt
- `--manager` flag — force a specific package manager
- Colored terminal output (respects `NO_COLOR`)
- Fallback to next available manager when preferred is unavailable
- Graceful degradation when Repology is unreachable
- GitHub Actions CI: test, race detector, coverage, lint, cross-compile matrix

---

## Template for future releases

## [X.Y.Z] — YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
