Feature: Ebitengine window

Background:
  Given the Ebitengine window task is accepted

Scenario: the application opens without scene data
  When the coder starts the desktop application
  Then the application window should open successfully
  And the world should be empty

Scenario Outline: the application window is resizable
  When the coder resizes the application window to <window_size>
  Then the application should continue running

Examples:
  | window_size |
  | small       |
  | large       |

Scenario Outline: simulation pause state controls stepping
  Given the application simulation pause state is <paused>
  When the coder updates the application loop
  Then simulation stepping should be <stepping>
  And input handling should remain active
  And rendering should remain active

Examples:
  | paused | stepping |
  | true   | stopped  |
  | false  | active   |

Scenario: closing the window exits cleanly
  When the coder closes the application window
  Then the application should exit without error
