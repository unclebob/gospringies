# Task 03: Domain Model

## Goal

Represent the world state independently of graphics.

## Scope

- Define masses with id, position, velocity, mass value, elasticity, and fixed state.
- Define springs with id, two endpoint mass ids, spring constant, damping constant, and rest length.
- Define the world as masses, springs, walls, forces, and current editor defaults.
- Enforce basic validation for duplicate ids and missing spring endpoints.

## Acceptance Notes

- A world can contain zero or more masses and springs.
- A fixed mass is represented in the domain explicitly, even if file format uses a negative mass value.

## Done When

- Unit tests cover creation, lookup, validation, and fixed-mass representation.
