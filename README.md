# GoSpringies

In the late 80s I used to play with XSpringies on a Sun SPARCstation. I decided to have swarm-forge whip this up for me in Go, complete with Gherkin tests, unit tests, CRAP analysis, mutation tests, Gherkin mutation, and the rest of the engineering machinery. I have not reviewed any of the Go code. So far, so good.

GoSpringies is an experimental Go remake inspired by the original XSpringies spring-mass simulation. It is a small playground for masses, springs, forces, walls, and demo scenes loaded from XSpringies-style files.

This project was built in Go with help from [swarm-forge](https://github.com/unclebob/swarm-forge).

The application is still incomplete. Some UI controls are present but not fully wired yet, and some examples can still show mathematical instabilities that need to be smoothed out in the simulation and timestep handling.

## UI Gestures

- Click an empty spot on the canvas to add a mass there.
- Mass and spring placement clicks outside the drawing canvas are ignored. Dragging an existing mass is constrained to the canvas edge.
- Click a mass to select it. Shift-click adds the mass to the current selection.
- Drag a mass to move it. If the mass is part of a selection, the whole selected group moves.
- Hold T when releasing a dragged mass to give the dragged mass or selected group the velocity vector of the drag.
- Drag on empty canvas to rubber-band a selection rectangle. Shift-drag adds enclosed objects to the current selection.
- Press Escape to clear the selection.
- Control-drag, or Command-drag on macOS, from one mass to another to create a spring.
- Control-click, or Command-click on macOS, on empty canvas creates a mass and starts rubber-banding a spring from it. The next click places the next mass and connects the spring. Keep Control down to continue placing a connected chain of masses; release Control before the next click to end the chain. Clicking an existing mass while chaining connects to that mass and ends the chain.
- Right-click a mass to open a menu for fixing/freeing it or setting its mass.
- Right-click near a spring to open spring settings for Kspring, Kdamp, and RestLen.
- Use Edit > Cut, Copy, and Paste, or their keyboard shortcuts, for selected objects. Paste places copied objects at the current mouse location.

## File Save and Load

- Click Save, or press Ctrl+S, to open the save filename dialog.
- The save filename field starts as `.xsp`, with the cursor before the extension. Type the filename before `.xsp`; for example, typing `simple hex` saves `saves/simple hex.xsp`.
- Saved simulations are written to the local `saves/` directory.
- Click Load, or press Ctrl+O, to open the load picker.
- The load picker refreshes the available files each time it opens. Saved files from `saves/` are listed first, followed by a separator, then starter demos from `demos/` and original demos from `demos/original/`.
- Choose a saved or demo `.xsp` file from the picker to replace the current world with that file.

## Right Control Panel

- The panel is grouped into Selected Mass(es), Selected Spring(s), Forces, Simulation, and Display sections.
- Mass and Elasticity apply to newly created masses or selected masses when applicable.
- Fixed, Set Center, and Collide share one row. Fixed toggles selected mass fixation, Set Center makes the selected mass the center-force target, and Collide toggles mass-to-mass collision.
- Kspring, Kdamp, and RestLen apply to newly created springs or selected springs when applicable. RestLen can also be set from the spring context menu.
- Numeric sliders show the current committed value. The `<` and `>` buttons on the left and right of each slider decrement or increment the value by `0.1`; holding either button for half a second repeats the change every tenth of a second.
- A slider text box shows the committed setting value. Clicking the text box selects the whole value for editing. Typed changes are local to the text box until Enter or Return is pressed, then the value is committed and the selection highlight is removed.
- Gravity, Center Attraction, CM Attraction, and Wall Repulsion have checkboxes to the left of their slider labels. The checkbox enables the force; the slider sets its magnitude.
- The `T`, `B`, `L`, and `R` checkboxes beside Wall Repulsion toggle the top, bottom, left, and right simulation walls.
- The Display section contains Grid, which toggles grid snapping, and Springs, which toggles spring visibility.
- Viscosity adjusts drag from the surrounding medium.
- Stick adjusts wall stickiness.
- Speed controls simulation speed; zero pauses advancement.
- Time Step controls timestep size. The Adapt checkbox beside Precision toggles adaptive timestep behavior, and the Precision slider controls the adaptive precision value.
- The status indicators are at the bottom of the right panel in three rows: simulation/object state, current file, and dirty/error state.
