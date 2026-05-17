Feature: State save restore

Background:
  Given the state save restore task is accepted

Scenario Outline: saved state can be restored repeatedly
  Given the world is in state <saved_state>
  When the coder saves state
  And the world changes to state <changed_state>
  And the coder restores state <restore_count> times
  Then the world should be in state <saved_state>

Examples:
  | saved_state | changed_state | restore_count |
  | A           | B             | 1             |
  | A           | B             | 2             |

Scenario: restore without saved state restores initial state
  Given no state has been saved
  And the world has changed from the initial state
  When the coder restores state
  Then the world should be in the initial state

Scenario Outline: state save restore is separate from file operations
  Given the world is in state <memory_state>
  When the coder saves state
  And the coder performs file operation <file_operation>
  And the coder restores state
  Then the world should be in state <memory_state>

Examples:
  | memory_state | file_operation |
  | A            | save file      |
  | A            | load file      |
