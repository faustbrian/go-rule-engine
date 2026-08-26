# Changelog

All notable changes to this project are documented here. The format follows
Keep a Changelog and semantic versioning.

## Unreleased

### Documentation

- Replace archived monorepo links and completed execution artifacts with a
  standalone, human-oriented documentation structure.

## 1.0.0 - 2026-08-25

### Fixed

- Bind the reviewed zero-mutant `jsonast` delegation facade to its exact
  standalone source identity.

### Changed

- Upgrade the competitor benchmark's Git, cryptography, and network
  dependencies to current security-fixed releases.

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Link the package README to package-owned documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-rule-engine` identity while preserving its documented API and behavior.
- Delegate core and adapter mutation checks to the canonical exact-100
  repository runner instead of package-specific thresholds and exclusions.
- Keep standalone module tidiness in the release gate instead of requiring an
  unpublished canonical tag before running local competitor benchmarks.
- Verify optional domain adapters through their independently attributable
  module gates instead of duplicating them in the core integration gate.
- Let isolated compilers canonically serialize and parse definitions that use
  their registered custom operators while preserving built-in-only package
  helpers.

### Added

- Typed immutable facts, propositions, compiler, and execution plans.
- Deterministic conflict strategies and bounded forward chaining.
- Canonical JSON AST serialization and SHA-256 hashing.
- Explicit typed operators, fact resolvers, and bounded plan caching.
- Truth-table, hostile-input, race, fuzz, and benchmark suites.
