# mutation-stamp: 6f9772a7acefccea4402711d6d16e34409718d53c1a87c7c26dd0b5afff58d27
Feature: Clickable visible controls

Background:
  Given the clickable visible controls task is accepted

Scenario Outline: clicking command controls runs commands
  When the coder clicks inside rendered bounds of visible control <control>
  Then command <command> should run

Examples:
  | control | command |
  | Pause   | pause   |
  | Run     | run     |
  | Reset   | reset   |
  | Quit    | quit    |

Scenario: clicking Load opens the demo picker
  When the coder clicks inside rendered bounds of visible control Load
  Then the demo picker should open

Scenario Outline: clicking path-based file controls opens keyboard path entry
  When the coder clicks inside rendered bounds of visible control <control>
  Then keyboard path entry should open for <command>

Examples:
  | control | command |
  | Insert  | Insert  |
  | Save    | Save    |

Scenario Outline: clicked controls match keyboard shortcut behavior
  Given visible control <control> maps to shortcut <shortcut>
  When the coder clicks inside rendered bounds of visible control <control>
  Then the result should match pressing shortcut <shortcut>

Examples:
  | control | shortcut |
  | Pause   | Space    |
  | Reset   | R        |
  | Load    | Ctrl+O   |
  | Insert  | Ctrl+I   |
  | Save    | Ctrl+S   |
  | Quit    | Q        |

Scenario: clicking outside visible controls does nothing
  Given the application state is recorded
  When the coder clicks outside all visible controls
  Then the application state should remain unchanged

Scenario Outline: clicking run and pause controls changes simulation state
  Given simulation state is <old_state>
  When the coder clicks inside rendered bounds of visible control <control>
  Then simulation state should be <new_state>

Examples:
  | old_state | control | new_state |
  | running   | Pause   | paused    |
  | paused    | Run     | running   |
