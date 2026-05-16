# Task 08: Ebitengine Window

## Goal

Create the desktop application shell using Ebitengine.

## Scope

- Open a resizable application window.
- Run an update/draw loop.
- Maintain an app state that references the plain Go world.
- Add pause/resume state for simulation.

## Acceptance Notes

- The app opens without requiring scene data.
- Closing the window exits cleanly.
- Simulation can be paused while input and rendering continue.

## Done When

- The app command runs locally.
- Domain tests still run without Ebitengine.
