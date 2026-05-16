# Task 10: Mouse Editing

## Goal

Support direct mouse creation and manipulation of scene objects.

## Scope

- Add mass placement at the pointer.
- Add spring placement between two selected or clicked masses.
- Add dragging for movable masses.
- Respect grid snap when enabled.

## Acceptance Notes

- Clicking in add-mass mode creates a mass using current mass defaults.
- Creating a spring connects existing masses and uses current spring defaults.
- Dragging a movable mass updates its position without changing its id.

## Done When

- Unit tests cover grid snap and edit operations in plain Go where possible.
- Manual UI smoke test confirms mouse workflows.
