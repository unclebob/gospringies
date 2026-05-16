# Task 015: Edit Mode Details

## Goal

Match XSpringies edit-mode object selection and pointer manipulation behavior.

## Scope

- Left click selects the nearest object and clears other selections.
- Shift-left click toggles the nearest object's selection without clearing other selections.
- Dragging empty space creates a selection box.
- Shift-selection box adds enclosed objects to the current selection.
- Middle-button drag moves selected objects while preserving their relative positions.
- Right-button drag throws selected masses by applying release pointer velocity.
- Right-click without movement stops selected masses.

## Acceptance Notes

- Selection behavior applies to masses and springs where applicable.
- Throw behavior applies velocity to selected movable masses.
- Fixed masses remain fixed during move and throw actions.

## Done When

- Edit-mode selection, box selection, move, throw, and stop behaviors are covered by acceptance and focused unit tests.
