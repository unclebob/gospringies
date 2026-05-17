# Task 024: Original Style Human Interface

## Goal

Make the visible interface operable by a person while staying close to the original XSpringies interaction model.

## Scope

- Draw the interface using custom Ebitengine rendering.
- Use an original XSpringies-like layout and control vocabulary.
- Keep the canvas as the central working area.
- Draw a visible left mode toolbar for Select, Mass, Spring, and Drag.
- Draw a visible top command bar for Run, Pause, Reset, Save State, Restore State, Load, Insert, Save, and Quit.
- Draw a visible right inspector with mass controls, spring controls, force controls, wall toggles, grid snap, show springs, timestep, precision, and adaptive timestep.
- Draw a visible bottom status line with mode, run state, object counts, selected object count, current file, dirty state, and last error.
- Use keyboard path entry for Load, Insert, and Save commands.
- Show validation and file errors in the status line.
- Keep controls clickable and mirror their keyboard shortcuts.

## Acceptance Notes

- Do not add a separate GUI toolkit; controls are custom-drawn in Ebitengine.
- The interface should be close to original XSpringies in behavior and control vocabulary, not necessarily pixel-identical.
- A user must be able to operate the visible controls without knowing hidden test-only APIs.
- Prioritize visible usability before further deep physics compatibility work.

## Done When

- Launching the app presents a recognizable XSpringies-like editor.
- A user can switch modes, pause/run, reset, load, insert, save, save/restore state, and edit common parameters through visible controls.
- File path entry for load, insert, and save works from the keyboard and reports errors visibly.
