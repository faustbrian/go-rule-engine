# Changelog

All notable changes to this project are documented here. The format follows
Keep a Changelog and semantic versioning.

## Unreleased

### Documentation

- Document published root v2.0.0 and math v2.0.0, measurement v3.0.0 and
  temporal v2.0.0 adapters, with coordinated migration and installation guidance.
  Earlier preparation entries below record their state before publication.

### Security

- Update the development-only documentation parser `smol-toml` to 1.9.0
  for GHSA-r4xh-jqrq-34v2. The root Go runtime has no external dependencies.

### Changed

- Resume the adapter successor-major and competitor migrations against the
  published root v2.0.0 dependency. Adapter tags remain pending; historical
  root-v1-compatible adapter releases and persisted encodings are unchanged.

- Prepare root v2: cache reuse requires complete compiler-limit equality,
  canonicalization uses the compiler registry, and one bounded operation
  context covers cache compilation and resolver evaluation. Caller deadlines
  and cancellation remain authoritative. Public canonical helpers retain
  their background-context API.
- Prepare math adapter v2, measurement adapter v3, and temporal adapter v2
  for root v2 nominal types; algorithms and published domain dependencies
  remain unchanged. None of these four module versions had been published
  when this preparation was recorded; actual availability follows public tags.
- Order delivery root first: retain published adapter and competitor sources
  on root v1 until root v2 is public, then resume the retained successor-major
  migrations. No pending adapter release is declared completed by this batch.

### Changed

- Adopt the checksum-verified Golib v1.8.5 CLI and immutable v1.7.2 workflow
  so independently versioned adapters can select a module for release
  rehearsal. A blank selector retains the all-module default and required CI
  checks remain strict.

- Raise the minimum supported Go version from 1.26.6 to 1.27.0.
- Adopt the checksum-verified `go-library-tools` v1.4.0 CLI and immutable
  W14 workflow so repository verification prefers canonical public module
  identities before source-built fallback, and record canonical public
  module checksums in every reverse consumer.

- Adopt the checksum-verified `go-library-tools` v1.3.0 CLI, schema-v2
  cohesion metadata for all four releasable modules, a local cohesion gate,
  and immutable reusable-workflow enforcement.

- Adopt the released shared `go-library-tools` workflow and configuration for
  repository checks while retaining the rule-engine-specific integration,
  competitor, and documentation checks.

### Documentation

- Complete the stable-v1 installation, package-selection, lifecycle,
  troubleshooting, performance, and project-navigation contract for the root
  module and all three optional adapters.
- Publish a dedicated measurement performance and troubleshooting guide and
  bind module metadata and documentation checks to the complete guide set.
- Correct support and private security-reporting routes.

- Add a direct engineering entry point for the non-releasable competitor
  benchmark harness and replace its placeholder inventory purpose.

- Link the core and adapter documentation to the immutable v1.4.0 ecosystem
  index and publish package selection, ownership, lifecycle, compatibility,
  and delivery metadata.

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
