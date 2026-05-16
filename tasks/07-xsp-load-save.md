# Task 07: XSP Load And Save

## Goal

Load and save the original human-readable `.xsp` file format.

## Scope

- Parse files starting with `#1.0`.
- Support known commands for parameters, forces, walls, masses, and springs.
- Save worlds back to deterministic `.xsp` text.
- Preserve fixed masses through the negative-mass file representation.
- Reject malformed input with useful errors.

## Acceptance Notes

- Loading then saving a simple scene yields deterministic output.
- Invalid duplicate mass ids and missing spring endpoints are reported.
- Files must end in a newline.

## Done When

- Unit tests cover successful parsing, deterministic saving, round trips, and malformed input.
