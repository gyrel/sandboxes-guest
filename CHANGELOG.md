# Changelog

## v0.0.6 — 2026-11-16

### Changed

- Remove the mandatory external release-review gate while retaining candidate validation, CI, and immutable release-tag protections.
- Accept an `Unreleased` heading before the newest version during release validation.
- Pin CI and release builds to Go 1.27.4.

## v0.0.3 — 2026-11-03

### Changed

- Normalize the Go module path to the lowercase `github.com/quoiner/sandboxes-guest` namespace.
- Publish a CycloneDX module SBOM as a CI artifact.

