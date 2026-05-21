# mutation-stamp: sha256=99fcb35cf9f1193d5549fe6cbb7a962e46f39facb23d702a1a0c83f89adad1e2
Feature: Screen and controls

Background:
  Given the screen and controls task is accepted

Scenario: the first screen is the simulation editor
  When the coder starts the desktop application
  Then the first screen should show the simulation editor
  And the first screen should not show a landing page

Scenario Outline: the editor screen contains required regions
  When the coder lays out the editor screen
  Then screen region <region> should be visible
  And screen region <region> should have purpose <purpose>

Examples:
  | region        | purpose                                      |
  | canvas        | edit and view the simulation world           |
  | left toolbar  | run selection commands                       |
  | top bar       | run commands and file commands               |
  | right inspector | edit selected objects and world parameters and show simulation state |

Scenario Outline: commands are visible controls
  When the coder shows the top command bar
  Then command <command> should have a visible control

Examples:
  | command      |
  | pause toggle |
  | reset        |
  | load         |
  | insert       |
  | save         |
  | quit         |

Scenario Outline: visible state reflects application state
  Given application state <state> is active
  When the coder renders the editor controls
  Then visible indicator <indicator> should reflect <state>

Examples:
  | state           | indicator        |
  | paused          | simulation state |
  | running         | simulation state |
  | object selected | selection        |
  | unsaved changes | file state       |

Scenario Outline: keyboard shortcuts mirror visible controls
  Given command <command> has visible control <control>
  When the coder presses keyboard shortcut <shortcut>
  Then command <command> should run

Examples:
  | command      | control      | shortcut |
  | pause toggle | pause toggle | Space    |
  | reset        | reset        | R        |
  | save         | save         | Ctrl+S   |
  | load         | load         | Ctrl+O   |
  | insert       | insert       | Ctrl+I   |
  | quit    | quit    | Q        |

Scenario Outline: controls remain usable during simulation states
  Given simulation state is <simulation_state>
  When the coder renders the editor screen
  Then the canvas should remain visible
  And the visible controls should remain usable

Examples:
  | simulation_state |
  | paused           |
  | running          |
