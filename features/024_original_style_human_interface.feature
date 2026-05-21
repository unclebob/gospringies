# mutation-stamp: sha256=bb95bad9cf91da334e9a3f72829f94c92dd066dd0d08dd42750e81dc301dbc02
Feature: Original style human interface

Background:
  Given the original style human interface task is accepted

Scenario Outline: the interface is custom-drawn in Ebitengine
  When the coder renders the editor interface
  Then interface element <element> should be drawn by the Ebitengine app

Examples:
  | element         |
  | canvas          |
  | left toolbar    |
  | top command bar |
  | right inspector |

Scenario Outline: original command controls are visible and clickable
  When the coder renders the top command bar
  Then command control <command> should be visible
  And command control <command> should be clickable

Examples:
  | command      |
  | pause toggle |
  | reset        |
  | load         |
  | insert       |
  | save         |
  | quit         |

Scenario Outline: inspector controls expose editable settings
  When the coder renders the right inspector
  Then inspector control <control> should be visible
  And inspector control <control> should be editable

Examples:
  | control                    |
  | Mass                       |
  | Elasticity                 |
  | Fixed Mass                 |
  | Kspring                    |
  | Kdamp                      |
  | Set Rest Length            |
  | Gravity                    |
  | Center Attraction          |
  | Center Of Mass Attraction  |
  | Wall Repulsion             |
  | Wall Toggles               |
  | Grid Snap                  |
  | Show Springs               |
  | Time Step                  |
  | Precision                  |
  | Adaptive Time Step         |

Scenario Outline: right inspector reports current application state
  Given application state <state> is active
  When the coder renders the right inspector
  Then status field <field> should show <state>

Examples:
  | state   | field         |
  | Masses: | object counts |
  | File:   | current file  |
  | saved   | file state    |

Scenario Outline: file commands use keyboard path entry
  When the coder activates file command <command>
  Then keyboard path entry should open for <command>
  When the coder submits path <path>
  Then file command <command> should use path <path>

Examples:
  | command | path              |
  | Load    | demos/pendulum.xsp |
  | Insert  | demos/spring-chain.xsp |
  | Save    | out/current.xsp   |

Scenario Outline: visible controls mirror keyboard shortcuts
  Given visible control <control> invokes command <command>
  When the coder presses shortcut <shortcut>
  Then command <command> should run

Examples:
  | control | command      | shortcut |
  | Pause   | pause toggle | Space    |
  | Reset   | Reset        | R        |
  | Load    | Load         | Ctrl+O   |
  | Insert  | Insert       | Ctrl+I   |
  | Save    | Save    | Ctrl+S   |
  | Quit    | Quit    | Q        |
