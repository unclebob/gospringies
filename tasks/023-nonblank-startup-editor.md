# Task 023: Nonblank Startup Editor

## Goal

Make the application visibly useful immediately when launched.

## Scope

- Show the working simulation editor chrome on first launch.
- Draw the canvas, left toolbar, top command bar, right inspector, and status line as visible regions.
- Start with visible world content, either by loading a default demo scene or by creating an equivalent built-in starter scene.
- Keep TPS/debug information from being the only visible content.
- Keep startup behavior deterministic.

## Acceptance Notes

- The first screen must not be blank except for debug text.
- The visible controls do not need to be fully styled, but they must be recognizable and placed in the intended screen regions.
- The startup scene must contain at least one fixed mass, one movable mass, and one spring.

## Done When

- Launching the app shows editor controls and non-empty simulation content without user action.
- The startup state is covered by acceptance tests and focused app-level tests.
