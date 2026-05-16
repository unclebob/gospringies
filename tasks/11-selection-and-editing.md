# Task 11: Selection And Editing

## Goal

Support selecting, duplicating, and deleting masses and springs.

## Scope

- Select individual masses and springs.
- Select all objects.
- Delete selected objects.
- Duplicate selected objects with new ids.
- When deleting a mass, delete attached springs.

## Acceptance Notes

- Deleting a mass removes springs attached to that mass.
- Duplicating selected objects creates independent objects with unique ids.
- Select all selects every mass and spring in the world.

## Done When

- Unit tests cover select, delete, cascading spring deletion, and duplication.
