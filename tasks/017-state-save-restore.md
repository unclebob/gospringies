# Task 017: State Save Restore

## Goal

Add XSpringies-style in-memory Save State and Restore State behavior distinct from file save/load.

## Scope

- Save State records current masses, springs, and system parameters in memory.
- Restore State restores the latest saved state.
- If no state has been saved, Restore State restores the initial state.
- Reset still clears objects and restores default parameters.
- File save/load behavior remains separate from Save State and Restore State.

## Acceptance Notes

- Saved state is not a file operation.
- Restoring can be repeated without consuming the saved state.

## Done When

- Save State, Restore State, restore-without-save, and separation from file operations are covered.
