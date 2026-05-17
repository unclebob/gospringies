Feature: Nonblank startup editor

Background:
  Given the nonblank startup editor task is accepted

Scenario: startup screen shows more than debug text
  When the coder starts the desktop application
  Then the first screen should show visible editor chrome
  And the first screen should show visible world content
  And debug text should not be the only visible content

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
