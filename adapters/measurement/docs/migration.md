# Migration

## Pending module v3

Update the adapter import to
`github.com/faustbrian/go-rule-engine/adapters/measurement/v3` and construct
quantities with `github.com/faustbrian/go-measurement/v2`. The parent
`github.com/faustbrian/go-rule-engine/v2` supplies the new nominal root types.
V3 is pending publication; published adapter v2.0.1 remains unchanged.
Existing quantity:v1
persisted values and operator names require no rewrite for this module upgrade.

## Historical unversioned encoding

The current encoding replaces the historical unversioned form:

```text
quantity:<amount> <unit>
```

with:

```text
quantity:v1|<amount>|<unit>
```

Regenerate persisted rule literals and facts from validated
`measurement.Quantity` values by calling `Quantity`. Do not rewrite arbitrary
strings in place: parse legacy values with an explicit canonical measurement
profile, validate the unit and amount, then encode the resulting quantity.

Deploy readers that understand v1 before writing v1 values. The adapter does
not accept both formats because permissive dual parsing would make persisted
identity and retirement of the legacy grammar ambiguous.
