Feature: Clickable visible controls

Background:
  Given the clickable visible controls task is accepted

Scenario Outline: clicking mode controls changes active mode
  Given the editor mode is <old_mode>
  When the coder clicks visible control <control>
  Then the editor mode should be <new_mode>
  And visible control <control> should show active state

Examples:
  | old_mode | control     | new_mode   |
  | select   | Mass        | add mass   |
  | select   | Spring      | add spring |
  | select   | Drag        | drag       |
  | add mass | Select      | select     |

Scenario Outline: clicking command controls runs commands
  When the coder clicks visible control <control>
  Then command <command> should run

Examples:
  | control | command |
  | Pause   | pause   |
  | Run     | run     |
  | Reset   | reset   |
  | Quit    | quit    |

Scenario Outline: clicking file controls opens keyboard path entry
  When the coder clicks visible control <control>
  Then keyboard path entry should open for <command>

Examples:
  | control | command |
  | Load    | Load    |
  | Insert  | Insert  |
  | Save    | Save    |

Scenario Outline: clicked controls match keyboard shortcut behavior
  Given visible control <control> maps to shortcut <shortcut>
  When the coder clicks visible control <control>
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
  When the coder clicks visible control <control>
  Then simulation state should be <new_state>

Examples:
  | old_state | control | new_state |
  | running   | Pause   | paused    |
  | paused    | Run     | running   |
