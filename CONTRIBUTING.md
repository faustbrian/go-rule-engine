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

Do not fork shared verification workflows, add permanent replacements, machine-specific
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

CI and local Make use the checksum-pinned public Tools v2.4.0 verifier without
building that release from source. The immutable shared workflow and tooling
source both select `0e8ff6fee42e88858069a6f6ae6f1080d95c5eae`; PR, main and
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

## Selected Monthly Assurance

The owner-level [selected-assurance caller](.github/workflows/assurance.yml) is
configured for main only, on day 19 of each month at 06:13 UTC. It selects one
of the 2 current declared root-module fuzz targets using the shared UTC-week
rotation, with 30 seconds of active exploration. The selected risks are:
bounded JSON definition parsing with stable canonical round-trip hashes,
and validated path segment identity.
Only the root module is selected. The six declared fuzz targets in the math,
measurement and temporal adapter modules retain their original full/manual
budgets, as do competitor and integration consumers. Selected root fuzzing
does not qualify those adapters or replace concurrent evaluation assertions.

Base-free ordinary checks execute deterministic seeds, not active exploration;
committed-range CI selects affected active smoke. Manual assurance requires an
exact anchored selector from the root fuzz operation in `.golib.yaml`, for
example `^FuzzParseJSON$`.
Empty, unknown and malformed selections refuse. The shared job bounds setup
and execution to ten minutes, active workers and CPU to two, owned memory to
2 GiB, native output to 1 MiB and private-cache lifetime. Failure, cancellation,
exhausted budgets and cleanup refusal remain attributable failures. This lane
does not run mutation or replace Required, original full fuzz budgets,
explicit `make check-full`, or relevant release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.

## Source Documentation

Install the checksum-pinned public verifier declared in `.golib.yaml`.
`make source-docs-check` validates objective Go comment, marker, link and
generated-producer structure; `make check`, `make check-full` and `make ci`
require it before their existing checks. Shared CI runs the same strict
contract and includes its result in Required. This is structural enforcement,
not a substitute for source-truth review or behavioral verification.
