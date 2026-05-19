# GoSpringies

In the late 80s I used to play with XSpringies on a Sun SPARCstation. I decided to have swarm-forge whip this up for me in Go, complete with Gherkin tests, unit tests, CRAP analysis, mutation tests, Gherkin mutation, and the rest of the engineering machinery. I have not reviewed any of the Go code. So far, so good.

GoSpringies is an experimental Go remake inspired by the original XSpringies spring-mass simulation. It is a small playground for masses, springs, forces, walls, and demo scenes loaded from XSpringies-style files.

This project was built in Go with help from [swarm-forge](https://github.com/unclebob/swarm-forge).

The application is still incomplete. Some UI controls are present but not fully wired yet, and some examples can still show mathematical instabilities that need to be smoothed out in the simulation and timestep handling.

## UI Gestures

- Click an empty spot on the canvas to add a mass there.
- Click a mass to select it. Shift-click adds the mass to the current selection.
- Drag a mass to move it. If the mass is part of a selection, the whole selected group moves.
- Hold T when releasing a dragged mass to give the dragged mass or selected group the velocity vector of the drag.
- Drag on empty canvas to rubber-band a selection rectangle. Shift-drag adds enclosed objects to the current selection.
- Press Escape to clear the selection.
- Control-drag, or Command-drag on macOS, from one mass to another to create a spring.
- Right-click a mass to open a menu for fixing/freeing it or setting its mass.
- Right-click near a spring to set its spring constant.
- Use Edit > Cut, Copy, and Paste, or their keyboard shortcuts, for selected objects. Paste places copied objects at the current mouse location.

## File Save and Load

- Click Save, or press Ctrl+S, to open the save filename dialog.
- The save filename field starts as `.xsp`, with the cursor before the extension. Type the filename before `.xsp`; for example, typing `simple hex` saves `saves/simple hex.xsp`.
- Saved simulations are written to the local `saves/` directory.
- Click Load, or press Ctrl+O, to open the load picker.
- The load picker refreshes the available files each time it opens. Saved files from `saves/` are listed first, followed by a separator, then starter demos from `demos/` and original demos from `demos/original/`.
- Choose a saved or demo `.xsp` file from the picker to replace the current world with that file.

## Right Control Panel

- Mass, elasticity, fixed, spring constant, damping, and rest length controls apply to newly created objects or the selected object when applicable.
- Gravity enables gravity; the Gravity slider adjusts its magnitude.
- Center and CMass enable center attraction and center-of-mass attraction.
- WallRep enables wall repulsion; Collide enables mass-to-mass collision.
- SetCtr makes the selected mass the center-force target.
- Top, Bot, Left, and Right toggle the simulation walls.
- Grid toggles grid snapping; Springs toggles spring visibility.
- Viscosity adjusts drag from the surrounding medium.
- Stick adjusts wall stickiness.
- Speed controls simulation speed; zero pauses advancement.
- Step, Prec, and Adapt control timestep, precision, and adaptive timestep behavior.
