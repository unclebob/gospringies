# Task 018: Selected Object Parameter Editing

## Goal

Allow selected masses and springs to be edited through the visible controls.

## Scope

- Changing Mass updates selected masses.
- Changing Elasticity updates selected masses.
- Fixed Mass checkbox fixes or unfixes selected masses.
- Changing Kspring updates selected springs.
- Changing Kdamp updates selected springs.
- Set Rest Length updates selected springs to their current geometric length.
- Changing defaults with no compatible selection affects future object creation.

## Acceptance Notes

- Mass controls affect only selected masses.
- Spring controls affect only selected springs.
- Mixed selections update only compatible objects for each control.

## Done When

- Selected-object edits and default-edit behavior are covered by acceptance and focused unit tests.
