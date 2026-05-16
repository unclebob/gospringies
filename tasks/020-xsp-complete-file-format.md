# Task 020: XSP Complete File Format

## Goal

Complete support for the documented XSpringies `.xsp` file format.

## Scope

- Parse and save all documented commands: `cmas`, `elas`, `kspr`, `kdmp`, `fixm`, `shws`, `cent`, `frce`, `visc`, `stck`, `step`, `prec`, `adpt`, `gsnp`, `wall`, `mass`, and `spng`.
- Preserve boolean zero/non-zero semantics.
- Preserve positive integer id semantics for masses, springs, and center mass.
- Use `cent -1` for screen center.
- Enforce no blank lines.
- Enforce final newline.
- Add `.xsp` extension automatically for file operations when omitted.
- Respect `SPRINGDIR` as the default file directory when set.

## Acceptance Notes

- Loading a complete file replaces world state and parameters.
- Inserting a file loads only masses and springs and preserves current parameters.
- Saving emits deterministic, human-readable `.xsp`.

## Done When

- Every documented command has load/save coverage, error handling coverage, and round-trip coverage where applicable.
