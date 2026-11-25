# Changelog

## Unreleased

### Changed

- Rename the Go module path from `github.com/rheic/sandboxes-guest` to
  `github.com/gyrel/sandboxes-guest`. Breaking change; pre-1.0 alpha policy.

## v0.0.8 — 2026-11-19

### Fixed

- Remove stale links whose store entries disappeared on the next start. Agent
  files and foreign links are left untouched, matching Docker's read-only
  linking semantics.


## v0.0.7 — 2026-11-19

### Added

- Add the guest `link-skills` command and bounded linker for a read-only shared
  skills store. Existing agent files are never overwritten; conflicting links
  are counted and left untouched.
- Resolve the current user's home inside the guest from the account database
  instead of trusting a host path.
- Map `kimi` and `dsh` to the shared `.agents/skills` directory.

### Security

- Reject unsafe store entries, symlinked store entries, clean-root paths and
  unsupported agents. The linker itself does not mount or verify the store; the
  runtime must provide a trusted read-only source.


## v0.0.6 — 2026-11-16

### Changed

- Remove the mandatory external release-review gate while retaining candidate validation, CI, and immutable release-tag protections.
- Accept an `Unreleased` heading before the newest version during release validation.
- Pin CI and release builds to Go 1.27.4.

## v0.0.3 — 2026-11-03

### Changed

- Normalize the Go module path to the lowercase `github.com/quoiner/sandboxes-guest` namespace.
- Publish a CycloneDX module SBOM as a CI artifact.

