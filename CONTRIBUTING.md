# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's goals and docs.
2. Run `make inventory` and the narrow baseline gate for the module.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Required mutation gates must finish with zero surviving viable mutants.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Run during development:

```bash
make inventory
make check
```

Before submitting a repository-wide change:

```bash
make ci
```

The full scheduled and release gate is `make ci`. Report every unavailable or
failing command; do not describe partial results as release-ready.

The root explicitly selects coverage evidence mode in `.golib.yaml`. It runs
the complete instrumented module tests and requires a valid profile with
nonzero execution for every expected production package. Review must assess
reachable behavior, hostile inputs, and defensive guards independently of
statement counts. Optional adapters retain exact coverage; mutation, race,
fuzz, API, security, consumer, and all other applicable gates remain required.
The defensive JSON marshal-error guard is retained: the private validated,
acyclic DTO has no known admitted-input failure witness. The historical
[exact-coverage run](https://github.com/faustbrian/go-rule-engine/actions/runs/37348864803)
failed at `serialization.go:104.3,105.1` on source
`c23532180913e0a7c646cfd080ba6599a4f035df`; that failure is not a passing
result under this policy.

## Repository Tooling

The repository uses the `go-library-tools` workflow for shared
formatting, module hygiene, safety, tests, coverage, race, fuzz, mutation,
benchmark, API, documentation, and security gates. Repository-specific checks
live under `verification/` and are declared in `.golib.yaml`; do not copy the
shared tooling into this repository or add package-local workflows.

CI deliberately selects the development-only source route at immutable Tools
`606c3e9da5112086217f7108f335c2f88528bc14`: the workflow pin and `tooling_sha`
match, and `source_bootstrap: true` builds that source. Runtime-selected jobs
run full `golib check --all`, not the local shortcut. The stable `Required`
job still requires the shared workflow to succeed. The obsolete exact-coverage
diagnostic route is removed; its failed immutable run remains historical
development evidence, not a release substitute.
All module dependencies use published public versions. `public_dependencies:
true` resolves them through `proxy.golang.org` and `sum.golang.org`, without
the optional private bootstrap archive or a checksum-database bypass.

The retained `tool_version: v1.8.5` and checksum describe the published binary
route, which cannot decode the new optional coverage field. Local contributors
must explicitly select a binary built from the pinned development source via
`make GOLIB=/path/to/source-built/golib check`; the old published CLI is not an
equivalent verifier. Follow the pinned Tools [source-bootstrap contract](https://github.com/faustbrian/go-library-tools/blob/606c3e9da5112086217f7108f335c2f88528bc14/docs/workflows.md)
for isolated builds and cleanup. Tools v2 qualification, stable binary adoption,
and actual public consumers remain pending release boundaries. This preparation
does not publish root v2 or change the published adapters' root-v1 dependencies.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
