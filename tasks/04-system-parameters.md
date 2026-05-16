# Task 04: System Parameters

## Goal

Capture editable simulation parameters and defaults.

## Scope

- Add defaults for current mass, elasticity, spring constant, damping, viscosity, stickiness, timestep, precision, grid snap, and show-springs flag.
- Add force configuration for gravity, center attraction, center-of-mass attraction, and wall repulsion.
- Add wall enablement for top, left, right, and bottom walls.

## Acceptance Notes

- Resetting the world restores default system parameters.
- Loading a file replaces system parameters.
- Inserting a file leaves existing system parameters unchanged.

## Done When

- Unit tests cover default values, reset behavior, load replacement, and insert preservation.
