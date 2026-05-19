# mutation-stamp: sha256=3d3b9d2e24ae429a6202a7a0530bc639e31b5827367c79f1aba5536c5dd6828c
Feature: Controls and hotkeys

Background:
  Given the controls and hotkeys task is accepted

Scenario Outline: keyboard shortcuts invoke commands
  Given the application is running
  When the coder presses shortcut <shortcut>
  Then command <command> should run

Examples:
  | shortcut | command      |
  | Q        | quit         |
  | Space    | pause toggle |
  | Delete   | delete       |
  | Ctrl+A   | select all   |
  | R        | reset        |

Scenario Outline: file commands change world state correctly
  Given the world is in state <initial_state>
  When the coder runs file command <command>
  Then the world state should be <expected_state>
  And system parameters should be <parameter_result>

Examples:
  | initial_state | command | expected_state             | parameter_result          |
  | current world | save    | written to XSP file        | unchanged                 |
  | current world | load    | replaced by XSP file       | replaced by XSP file      |
  | current world | insert  | current plus inserted file | existing values preserved |

Scenario: reset clears objects and restores defaults
  Given the world contains objects
  And system parameters have custom values
  When the coder runs the reset command
  Then the world should contain zero masses
  And the world should contain zero springs
  And system parameters should equal defaults

Scenario Outline: parameter controls update editable settings
  Given parameter <parameter> has value <old_value>
  When the coder changes parameter <parameter> to <new_value>
  Then parameter <parameter> should have value <new_value>

Examples:
  | parameter       | old_value | new_value |
  | current mass    | default   | custom    |
  | spring constant | default   | custom    |
  | damping         | default   | custom    |
  | grid snap       | default   | custom    |
