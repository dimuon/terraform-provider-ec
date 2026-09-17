# Requirements verification — consistency, mapping, tests

## Consistency checks

### Identity and Import

Identity `id` format must match Import’s accepted id. Imported value must be what Create would store.

### Schema vs requirements

- “When [attribute] is configured” ⇒ that attribute is optional or optional+computed in the schema, not required-only.
- “The resource SHALL set [attribute] in state” ⇒ computed or optional+computed.
- Type-gated validation (“when type is X”) ⇒ `type` and the gated fields exist in the schema.

### Lifecycle

No pair that says attribute X both RequiresReplace and updates in place.

### Client

Create/Update/Read/Delete (or data-source Read) must agree on Stateful vs Serverless. Do not mix `cloud-sdk-go` and the generated serverless client for one type unless the code actually does.

### State and Plan/State

For one attribute, do not require both “preserve null” and “store empty list” when the API returns empty. `UseStateForUnknown` / defaults should match optional+computed or computed fields.

### API

Create/Update reference the APIs the code calls. Read and Delete match GET/DELETE (and association teardown if specified).

## Requirement → implementation mapping

| Category | Typical location | What to check |
| --- | --- | --- |
| **Type / client** | `schema.go` `Metadata` / `Configure` | TypeName; `ConvertProviderData`; Stateful vs Serverless. |
| **Unconfigured client** | `resourceReady` / equivalent in schema or resource file | Diagnostic summary `Unconfigured API Client`; no API call. |
| **Identity** | `create.go` (id set); plan modifier on `id` | API id stored; kept on later plans. |
| **Import** | `ImportState` in `schema.go` | Passthrough or custom; error on bad id if validated. |
| **Lifecycle** | `schema.go` plan modifiers | `RequiresReplace` vs in-place. |
| **Create/Update** | `create.go`, `update.go` | API call; read-after-mutate; missing-after error text. |
| **Read** | `read.go` | GET; not-found removes state; other errors keep state. |
| **Delete** | `delete.go` | Associations; not-found as success when required. |
| **Validation** | `ValidateConfig`, schema validators | Type-gated attributes; skip when unknown. |
| **Mapping** | `expanders.go`, `flatteners.go` | Empty string → null; omit unknown nested ids. |
| **Plan/State** | plan modifiers in `schema.go` | Defaults; unknown nested computed ids when the set changes. |

Data sources: Read, Type/client, Unconfigured client, Mapping, Plan/State only.

This provider has **no** `UpgradeState` / schema `Version` on resources. Do not look for StateUpgrade requirements or invent them.

## Test opportunity patterns

### Unit (no live API) — agent may run `make unit`

| Kind | Example | Verifies |
| --- | --- | --- |
| Expand/flatten | Empty API description → null in state | Mapping |
| ValidateConfig | Remote-cluster fields on `type=ip` | Type-gated validation |
| Import | ImportState sets id | Identity/Import |
| Unconfigured client | CRUD with nil client → diagnostic | Ready guard |
| Plan modifiers | Default `include_by_default` false | Plan/State |

### Acceptance (live Elastic Cloud API) — human / Buildkite only

Tests live in `ec/acc/` as `TestAcc…`. **Never** run them from this skill.

| Kind | Example | Verifies |
| --- | --- | --- |
| CRUD round-trip | `TestAccDeploymentTrafficFilter_basic` | Create/read/update/destroy |
| Type variant | `TestAccDeploymentTrafficFilter_remoteCluster` | Type-gated payload |
| Import | Acc ImportState step | Import + read |

Suggest: requirement heading, type (unit / acceptance), description, who runs (agent vs human/Buildkite). If `ec/acc/` already covers the case, say so rather than inventing a duplicate.
