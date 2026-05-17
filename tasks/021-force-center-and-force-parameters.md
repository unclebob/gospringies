# Task 021: Force Center And Force Parameters

## Goal

Specify the documented special force parameters and center-point behavior.

## Scope

- Gravity has Magnitude and Direction.
- Gravity direction uses degrees with `0.0` down and increasing counter-clockwise.
- Center-of-mass attraction has Magnitude and Damping.
- Center attraction has Magnitude and Exponent.
- Negative center-attraction magnitude repels masses from the center.
- Wall repulsion has Magnitude and Exponent.
- Set Center uses the single selected mass when exactly one mass is selected.
- Set Center uses screen center when no mass is selected.
- The center mass is visually marked.
- Center forces do not apply reciprocal response to the center mass.

## Acceptance Notes

- Only one force's parameter controls are active at a time.
- Enabling a force selects that force's parameter controls.

## Done When

- Force parameter behavior, center selection, center visual marking, and non-reciprocal center force behavior are covered.
