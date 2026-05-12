# Changelog

All notable changes to jil are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
jil uses [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added
- Initial release of `jil`
- `jil install` — install packages via the best available package manager
- `jil search` — search packages across managers via Repology API
- `jil config` — open config file in `$EDITOR`
- `jil config path` — print config file path
- Support for 10 package managers: apt, brew, dnf, pacman, zypper, apk, pkg, winget, choco, scoop
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
