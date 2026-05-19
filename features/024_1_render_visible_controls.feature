# mutation-stamp: sha256=15e8a314f21bfc699e6880e406f73ed6e8a9d2818f24c46d1d151c199ec78155
Feature: Render visible controls

Background:
  Given the render visible controls task is accepted

Scenario Outline: editor chrome regions produce visible pixels
  When the coder draws the application frame
  Then screen region <region> should contain non-background pixels
  And screen region <region> should not contain only debug text

Examples:
  | region          |
  | left toolbar    |
  | top command bar |
  | right inspector |
  | status line     |

Scenario Outline: visible controls have readable labels
  When the coder draws the application frame
  Then visible control <control> should have readable label <label>

Examples:
  | control       | label         |
  | run command   | Run           |
  | pause command | Pause         |
  | reset command | Reset         |
  | load command  | Load          |
  | insert command | Insert       |
  | save command  | Save          |
  | quit command  | Quit          |

Scenario Outline: inspector sections are visibly rendered
  When the coder draws the application frame
  Then inspector section <section> should be visible

Examples:
  | section    |
  | Mass       |
  | Spring     |
  | Forces     |
  | Walls      |
  | Simulation |

Scenario Outline: status fields are visibly rendered
  Given application state <state> is active
  When the coder draws the application frame
  Then status field <field> should be visible
  And status field <field> should show <state>

Examples:
  | state            | field          |
  | running          | run state      |
  | object counts    | object counts  |
  | saved            | file state     |

Scenario: world content remains visible with editor chrome
  When the coder draws the application frame
  Then the canvas should contain visible world content
  And editor chrome should not cover all world content

Scenario: default window size has no clipped control labels
  When the coder draws the application frame at the default window size
  Then visible control labels should fit inside their regions
