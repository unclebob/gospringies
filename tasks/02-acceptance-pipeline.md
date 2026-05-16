# Task 02: Acceptance Pipeline

## Goal

Put the acceptance pipeline from `github.com/unclebob/Acceptance-Pipeline-Specification` in place.

## Scope

- Add parser, generator, mutator, and scripts described by the acceptance pipeline.
- Keep generated acceptance tests separate from unit tests.
- Add one minimal feature file and generated executable test as a smoke check.

## Acceptance Notes

- Running acceptance tests means parser, generator, then generated executable tests.
- Generated files have a predictable location and are not mixed into hand-written unit tests.

## Done When

- The acceptance script runs from a clean checkout.
- The smoke feature passes.
