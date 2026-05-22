# mutation-stamp: sha256=b8fdc074919c6b5773c0225dca3fd637678223f5260cf3b633ea900358d7581a
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:07:28-05:00","feature_name":"Controls and hotkeys","feature_path":"features/012_controls_and_hotkeys.feature","background_hash":"83d899efaf3e7c994cfc3bb72e750341cb76a325fa0b7a8a89fec0de86851ec2","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"keyboard shortcuts invoke commands","scenario_hash":"eee2c19c6a343849a2f255dea41db92600e7fa1a55d050440807ba99af08b86d","mutation_count":10,"result":{"Total":10,"Killed":10,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:07:28-05:00"},{"index":1,"name":"file commands change world state correctly","scenario_hash":"5e0f95ce53b8b49848bc18a71e521414202a37ea5553ddfb6f349e76e2a3aa0e","mutation_count":9,"result":{"Total":9,"Killed":9,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:07:28-05:00"},{"index":2,"name":"reset clears objects and restores defaults","scenario_hash":"088c459b39793c21b8dabe3a9738ce713ec3cc2966030c96da99144d527803da","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:07:28-05:00"},{"index":3,"name":"parameter controls update editable settings","scenario_hash":"3a96ce69da0d0ba8456a935c33e1cc9ce9e1645f3605c9e188661b2cc4d06a92","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:07:28-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Controls and hotkeys

Background:
  Given the controls and hotkeys task is accepted

Scenario Outline: keyboard shortcuts invoke commands
  Given the application is running
  When the coder presses shortcut <shortcut>
  Then command <command> should run

Examples:
  | shortcut | command      |
  | Q        | quit         |
  | Space    | pause toggle |
  | Delete   | delete       |
  | Ctrl+A   | select all   |
  | R        | reset        |

Scenario Outline: file commands change world state correctly
  Given the world is in state <initial_state>
  When the coder runs file command <command>
  Then the world state should be <expected_state>
  And system parameters should be <parameter_result>

Examples:
  | initial_state | command | expected_state             | parameter_result          |
  | current world | save    | written to XSP file        | unchanged                 |
  | current world | load    | replaced by XSP file       | replaced by XSP file      |
  | current world | insert  | current plus inserted file | existing values preserved |

Scenario: reset clears objects and restores defaults
  Given the world contains objects
  And system parameters have custom values
  When the coder runs the reset command
  Then the world should contain zero masses
  And the world should contain zero springs
  And system parameters should equal defaults

Scenario Outline: parameter controls update editable settings
  Given parameter <parameter> has value <old_value>
  When the coder changes parameter <parameter> to <new_value>
  Then parameter <parameter> should have value <new_value>

Examples:
  | parameter       | old_value | new_value |
  | current mass    | default   | custom    |
  | spring constant | default   | custom    |
  | damping         | default   | custom    |
  | grid snap       | default   | custom    |
