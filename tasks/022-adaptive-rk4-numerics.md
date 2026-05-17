# Task 022: Adaptive RK4 Numerics

## Goal

Match XSpringies' documented RK4 and adaptive timestep simulation controls.

## Scope

- Use RK4 as the simulation integrator.
- Time Step controls fixed-step RK4 integration.
- Adaptive Time Step can be enabled or disabled.
- Precision controls adaptive RK4 error tolerance.
- Lower precision values produce smaller adaptive steps.
- Adaptive stepping must advance the requested simulation duration deterministically.

## Acceptance Notes

- The simulation remains independent of rendering frame rate.
- Adaptive stepping is a numerical behavior, not a UI rendering behavior.

## Done When

- Fixed RK4 stepping, adaptive enablement, precision behavior, and deterministic duration advancement are covered by tests.
