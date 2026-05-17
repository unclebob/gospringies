Feature: Nonblank startup editor 23.1

Background:
  Given the nonblank startup editor 23.1 task is accepted

Scenario Outline: startup screen shows more than debug text
  When the coder starts the desktop application
  Then the first screen should show visible editor chrome
  And the first screen should show visible world content
  And debug text should not be the only visible content
  And the startup world should be loaded from <default_demo>

Examples:
  | default_demo       |
  | demos/pendulum.xsp |

Scenario Outline: editor chrome is visible on startup
  When the coder starts the desktop application
  Then startup screen region <region> should be visible

Examples:
  | region          |
  | canvas          |
  | left toolbar    |
  | top bar         |
  | right inspector |
  | status line     |

Scenario Outline: startup world contains visible simulation objects
  When the coder starts the desktop application
  Then the startup world should contain <object_count> <object_type>

Examples:
  | object_type  | object_count |
  | fixed mass   | at least 1   |
  | movable mass | at least 1   |
  | spring       | at least 1   |

Scenario: startup state is deterministic
  When the coder starts the desktop application twice
  Then both startup worlds should be equivalent
  And both startup screens should show the same editor chrome

Scenario Outline: startup uses the default demo scene
  When the coder starts the desktop application
  Then the startup world should match demo file <default_demo>

Examples:
  | default_demo       |
  | demos/pendulum.xsp |
