# mutation-stamp: sha256=2cffad1deedf9449024295ee6ccfad66182ecb1485f2dad5bef73ba3189cdeb2
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:11:02-05:00","feature_name":"State save restore","feature_path":"features/017_state_save_restore.feature","background_hash":"f65b29e8cafa1eefa441ffdf2b32c2d8c212d4c24eab7b54ce0d7a05aa32ac2f","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"saved state can be restored repeatedly","scenario_hash":"d9ec30a2034cf069b6002255e765ae4571238e24bd628a5926bca032a472aa3b","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:02-05:00"},{"index":1,"name":"restore without saved state restores initial state","scenario_hash":"b332b6a6449d10236155379039ba8563352225f783a96b6b1e46ec21c483af85","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:02-05:00"},{"index":2,"name":"state save restore is separate from file operations","scenario_hash":"545d6e2dcb157d8fd2236a6b3be86ce468118e45ca04e5d57c32adf1bdd801fa","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:02-05:00"}]}
# acceptance-mutation-manifest-end
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
