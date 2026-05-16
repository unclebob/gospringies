# Task 09: Render World

## Goal

Draw masses, springs, walls, and selection state.

## Scope

- Render masses as visible circular markers.
- Render springs as lines when show-springs is enabled.
- Render enabled walls at window boundaries.
- Render fixed masses distinctly from movable masses.
- Keep drawing code in the UI/app boundary.

## Acceptance Notes

- A loaded world is visible in the window.
- Hiding springs removes spring lines while leaving masses visible.
- Fixed masses are visually distinguishable.

## Done When

- Rendering code can draw an empty world and a non-empty world without errors.
- Visual smoke checks confirm masses and springs appear.
