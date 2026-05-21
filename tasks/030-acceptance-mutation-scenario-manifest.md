# Acceptance Mutation Scenario Manifest

Acceptance mutation is getting slow when a feature file contains many scenarios. Add a scenario-level manifest so unchanged scenarios can be skipped in the same spirit that mutate4go skips unchanged functions.

Do not add Gherkin acceptance coverage for this task. This task is specified textually.

## Goal

`gherkin-mutator` should avoid re-running acceptance mutations for scenarios that have already passed mutation and have not changed since the last successful mutation run.

The unit of reuse is a scenario, not a whole feature file.

## Manifest Location

Store the scenario manifest in the feature file, using comment markers similar to the mutate4go manifest block.

Use stable markers:

```text
# acceptance-mutation-manifest-begin
# <json manifest line or lines>
# acceptance-mutation-manifest-end
```

The manifest must be ignored by the Gherkin parser and must not affect generated acceptance tests.

## Manifest Shape

The manifest should be JSON and deterministic. It must include:

- `version`
- `tested_at`
- feature identity, including feature name and feature file path
- a hash of the feature background
- a hash that represents the acceptance mutation implementation and execution contract
- one entry per scenario

Each scenario entry must include:

- stable scenario index
- scenario name
- scenario hash
- mutation count executed for that scenario
- result summary for that scenario

The scenario hash must include:

- scenario name
- step text and order
- example headers
- example row values

The background hash must include all background step text and order.

## Skip Rules

On startup, `gherkin-mutator` should parse the feature and the embedded acceptance mutation manifest.

For each scenario, skip that scenario's mutations only when all of these are true:

- manifest version is supported
- feature name matches
- scenario index and scenario name match
- scenario hash matches
- background hash matches
- acceptance mutation implementation hash matches
- the previous scenario result summary had no survived mutations and no errors

If any condition fails, rerun that scenario's mutations.

The manifest must be granular: changing one scenario reruns that scenario while unchanged scenarios in the same feature can be skipped.

Changing the background reruns all scenarios in the feature.

Changing acceptance mutation code, the generator contract, or the runtime contract reruns all scenarios covered by that manifest.

## Mutation IDs And Reports

Mutation IDs and paths must remain deterministic for a fixed feature.

Skipped scenarios must not be reported as newly killed mutations. The text report should make skipped work visible without pretending it executed. A concise summary line is enough, for example:

```text
skipped_scenarios=5 skipped_mutations=42
```

The command exit code must still reflect only mutations that actually ran during the current invocation:

- exit `0` when all executed mutations were killed and there were no errors
- exit `1` when any executed mutation survived or errored
- exit `2` for usage errors

When every scenario is skipped, exit `0` and print that the scenario manifest is valid for the skipped scenarios.

## Writing The Manifest

After a successful run in which every executed mutation is killed and no executed mutation errors, update the embedded manifest.

Keep still-valid skipped scenario entries.

Replace entries for scenarios that were executed in the current run.

Remove entries for scenarios that no longer exist.

The write should be deterministic so repeated successful runs without source changes do not create needless diffs beyond timestamp policy. If timestamps would cause churn, preserve an existing `tested_at` for skipped scenarios and update it only for executed scenarios.

## Backward Compatibility

Existing feature-level mutation stamps may remain supported as a fast path only when the whole feature is unchanged.

If both a feature-level mutation stamp and a scenario manifest exist, the implementation may use the whole-feature stamp to skip the entire feature. If the whole-feature stamp is missing or stale, the scenario manifest should still be used for per-scenario skipping.

Existing command-line options should continue to work.

## Verification Expectations

Add focused unit tests for:

- manifest parsing and writing
- scenario hash changes when scenario steps or examples change
- background hash changes invalidating every scenario
- unchanged scenarios are skipped while changed scenarios execute
- stale implementation hash invalidates manifest entries
- survived or errored scenario results are not skipped on the next run
- removed scenarios are removed from the rewritten manifest

Run the relevant `gherkin-mutator` tests and one direct mutation command on a feature with at least two scenarios, demonstrating that the second run skips unchanged scenarios.
