# Task 12: Controls And Hotkeys

## Goal

Expose essential controls for simulation and editing.

## Scope

- Add keyboard shortcuts for quit, pause/resume, reset, select all, and delete.
- Add basic on-screen controls or simple keyboard-driven parameter changes for early usability.
- Add load, save, and insert file commands.
- Keep behavior deterministic and documented.

## Acceptance Notes

- Reset clears objects and restores default parameters.
- Save writes the current world to `.xsp`.
- Load replaces the current world.
- Insert adds objects from a file without replacing current parameters.

## Done When

- Acceptance coverage exists for load, save, insert, and reset behavior.
- Manual smoke test covers key controls.
