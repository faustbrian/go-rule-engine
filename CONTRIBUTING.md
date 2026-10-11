# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's docs.
2. Run the standalone Go baseline below.
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

Explicit full and release mutation gates must finish with zero surviving viable
mutants. Local mutation is optional, not a universal contribution prerequisite.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use the declared Go 1.27.2 producer; module minima remain Go 1.27.0. From the
affected module directory, test without sibling checkouts or the repository
workspace:

```bash
GOWORK=off go test ./...
```

All modules declare no services or test/build tags. This command runs
deterministic fuzz seeds and saved corpus regressions, not active exploration.
With the checksum-pinned public verifier selected by `.golib.yaml`, run from
the repository root:

```bash
make inventory
make check
make ci
```

Both Make routes run bounded local checks across the declared modules. A root
race selection checks coherent concurrent context lookups, plan-cache access,
custom-operator evaluations and the exact call count, once within two minutes.
Base-free checks run seeds/corpora; changed-source CI selects bounded active
smoke for causally affected targets. For short active exploration choose one
target; optional local mutation chooses one package:

```bash
GOMAXPROCS=2 GOWORK=off go test -run '^$' -fuzz '^FuzzParseJSON$' -fuzztime=2s -parallel=1 -timeout=2m .
golib mutation --module . --package .
```

`make check-full` preserves the original full integration, coverage, race,
mutation, API, security, documentation and benchmark campaigns, including the
isolated competitor harness and all eight native two-second fuzz targets in the
root and three adapters. Run the relevant full campaigns for behavior changes
at those boundaries. Select full qualification explicitly with `make check-full`
or a workflow dispatch with `release_dry_run: true`; scheduled shared CI uses
bounded checks, not the full campaign.
Report unavailable or failing prerequisites; partial results are not release
evidence.

The full route explicitly selects root coverage evidence mode in `.golib.yaml`.
It runs the complete instrumented module tests and requires a valid profile with
nonzero execution for every expected production package. Review must assess
reachable behavior, hostile inputs, and defensive guards independently of
statement counts. Optional adapters retain exact coverage; full mutation, race,
fuzz, API, security and consumer contracts remain unchanged.
The defensive JSON marshal-error guard is retained: the private validated,
acyclic DTO has no known admitted-input failure witness. The historical
[exact-coverage run](https://github.com/faustbrian/go-rule-engine/actions/runs/37348864803)
failed at `serialization.go:104.3,105.1` on source
`c23532180913e0a7c646cfd080ba6599a4f035df`; that failure is not a passing
result under this policy.

## Repository Tooling

CI and local Make use the checksum-pinned public Tools v2.1.1 verifier without
building that release from source. The immutable shared workflow and tooling
source both select `eff8b2127c361ee2f422123824a1d337cc277ac4`; PR, main and
scheduled shared CI use the bounded local route. Full qualification is selected
explicitly by `make check-full` or the `release_dry_run: true` workflow input.
Required fails unless the shared workflow succeeds.
Shared setup still checks out its pinned tooling; no sibling library or goal
ledger is a prerequisite. Public dependencies resolve through
`proxy.golang.org` and `sum.golang.org`, without private bootstrap or checksum
bypass. Repository-specific full operations remain under `verification/`.
The failed immutable exact-coverage run above remains historical evidence,
not a release substitute. This CI adoption does not publish a library release
or change any adapter dependency.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
