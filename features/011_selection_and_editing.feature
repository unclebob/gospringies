# mutation-stamp: sha256=92ae0566263c87a135dd0bf1d2c1a257b0c886e24ec4f82be412b618d73f89fe
Feature: Selection and editing

Background:
  Given the selection and editing task is accepted

Scenario Outline: objects can be selected individually
  Given the world contains a <object_type> with id <id>
  When the coder selects <object_type> <id>
  Then <object_type> <id> should be selected

Examples:
  | object_type | id |
  | mass        | 1  |
  | spring      | 2  |

Scenario: select all selects every object
  Given the world contains masses and springs
  When the coder selects all objects
  Then every mass should be selected
  And every spring should be selected

Scenario Outline: deleting a selected object removes it
  Given the world contains a <object_type> with id <id>
  And <object_type> <id> is selected
  When the coder deletes selected objects
  Then <object_type> <id> should not exist

Examples:
  | object_type | id |
  | mass        | 1  |
  | spring      | 2  |

Scenario: deleting a mass deletes attached springs
  Given mass 1 is connected to mass 2 by spring 3
  And mass 1 is selected
  When the coder deletes selected objects
  Then mass 1 should not exist
  And spring 3 should not exist
  And mass 2 should still exist

Scenario Outline: duplicating selected objects creates independent objects
  Given selected <object_set> exists
  When the coder duplicates selected objects
  Then duplicated objects should have unique ids
  And duplicated objects should be independent from the original objects

Examples:
  | object_set             |
  | one mass               |
  | two masses and a spring |
