# Task 05: Force Evaluation

## Goal

Compute forces on masses without advancing time.

## Scope

- Compute spring force using Hooke's law.
- Compute spring damping along the spring direction.
- Compute gravity, viscosity, wall repulsion, and optional center forces.
- Ignore movement forces for fixed masses while still allowing movable connected masses to react.

## Acceptance Notes

- Equal and opposite spring forces are applied to spring endpoints.
- Fixed masses do not accumulate acceleration.
- Wall force pushes masses back into bounds when walls are enabled.

## Done When

- Unit tests cover each force type independently.
- Combined-force tests cover common interactions such as a pendulum mass attached to a fixed mass.
