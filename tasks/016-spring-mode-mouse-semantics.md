# Task 016: Spring Mode Mouse Semantics

## Goal

Match XSpringies spring-mode creation and temporary tugging behavior.

## Scope

- Starting near a mass anchors the first spring endpoint to that mass.
- Releasing near a second mass creates a spring between the two masses.
- Releasing away from a second mass discards the pending spring.
- Left-button creation actively affects the first mass while dragging.
- Middle-button drag creates a temporary spring from the first mass to the cursor and discards it on release.
- Right-button creation does not affect the first mass until the spring is placed.
- Created springs use current Kspring and Kdamp defaults.
- Created spring rest length equals the spring length at release.

## Acceptance Notes

- No spring is created unless both endpoints are valid masses.
- Temporary springs can affect simulation while they exist but are not saved into the world.

## Done When

- Spring creation, discard, temporary tugging, delayed activation, and rest-length behavior are covered.
