# GoSpringies

GoSpringies is an experimental Go remake inspired by the original XSpringies spring-mass simulation. It is a small playground for masses, springs, forces, walls, and demo scenes loaded from XSpringies-style files.

This project was built in Go with help from [swarm-forge](https://github.com/unclebob/swarm-forge).

The application is still incomplete. Some UI controls are present but not fully wired yet, and some examples can still show mathematical instabilities that need to be smoothed out in the simulation and timestep handling.

## UI Gestures

- Click an empty spot on the canvas to add a mass there.
- Click a mass to select it. Shift-click adds the mass to the current selection.
- Drag a mass to move it. If the mass is part of a selection, the whole selected group moves.
- Drag on empty canvas to rubber-band a selection rectangle. Shift-drag adds enclosed objects to the current selection.
- Press Escape to clear the selection.
- Control-drag, or Command-drag on macOS, from one mass to another to create a spring.
- Right-click a mass to open a menu for fixing/freeing it or setting its mass.
- Right-click near a spring to set its spring constant.
- Use Edit > Cut, Copy, and Paste, or their keyboard shortcuts, for selected objects. Paste places copied objects at the current mouse location.
