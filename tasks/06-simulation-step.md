# Task 06: Simulation Step

## Goal

Advance the world through time in a deterministic way.

## Scope

- Implement a simple initial integrator suitable for interactive behavior.
- Respect timestep and fixed masses.
- Keep simulation stepping independent of rendering frame rate.
- Add a stable public function that advances by a requested duration.

## Acceptance Notes

- A mass under gravity changes velocity and position over time.
- A fixed mass remains stationary after any number of steps.
- Smaller timesteps produce stable, deterministic results for simple scenes.

## Done When

- Unit tests cover gravity, fixed masses, springs, and repeated steps.
