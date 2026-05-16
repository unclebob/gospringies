# Task 08a: Screen And Controls

## Goal

Define the first usable desktop screen layout and visible controls before detailed rendering and editing behavior are implemented.

## Scope

- Use a single-window application layout.
- Put the simulation canvas in the center and let it fill the remaining space.
- Add a left toolbar for editing modes: select, add mass, add spring, and drag.
- Add a top command bar for run/pause, reset, load, insert, save, and quit.
- Add a right inspector for selected object properties and world parameters.
- Add a bottom status line for current mode, simulation state, object counts, and file state.
- Mirror important visible commands with keyboard shortcuts.
- Keep screen layout code in the UI/app boundary.

## Acceptance Notes

- The first screen is the working simulation editor, not a landing page.
- Controls remain visible and usable while the simulation is paused or running.
- The canvas remains the primary focus of the screen.
- The UI must not prescribe physics or file-format implementation details.

## Done When

- The app presents the screen regions and controls listed above.
- The visible control state reflects current mode, pause state, selection, and dirty file state.
- Keyboard shortcuts invoke the same commands as the visible controls.
