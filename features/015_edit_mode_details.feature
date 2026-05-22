# mutation-stamp: sha256=49ed3dfb4f9a8aedb1e979ab0a3b80f29ed8614c5d7c38d61717126e16e3054c
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:09:50-05:00","feature_name":"Edit mode details","feature_path":"features/015_edit_mode_details.feature","background_hash":"ebb229056455760dbd7aa1909d22fb87fc9cbe92455a3e308f2d615528d35c9b","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"left click selection controls selection set","scenario_hash":"38976cc9a3bd0a9543e008810952f0d44d6198b164436500608c4a8027472d93","mutation_count":12,"result":{"Total":12,"Killed":12,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:09:50-05:00"},{"index":1,"name":"selection box selects enclosed objects","scenario_hash":"890f9f3bead69262e66e547a7f433b74a5c1d01aca2900decef6e6d6d4e1b6c5","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:09:50-05:00"},{"index":2,"name":"middle drag moves selected objects","scenario_hash":"4afbe13e449071995557d98cfd5ef68b0338375261d0435362fc55ca2837e010","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:09:50-05:00"},{"index":3,"name":"right drag throws or stops selected masses","scenario_hash":"c7e6b417e21f7644a351d84d9b6eaf9801c17840a131c50ad6111a39ffd47bc5","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:09:50-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Edit mode details

Background:
  Given the edit mode details task is accepted

Scenario Outline: left click selection controls selection set
  Given edit mode is active
  And object <object_id> is near the pointer
  And selection initially contains <initial_selection>
  When the coder <click_action> object <object_id>
  Then selection should contain <expected_selection>

Examples:
  | object_id | initial_selection | click_action       | expected_selection |
  | 1         | none              | left clicks        | 1                  |
  | 2         | 1                 | shift left clicks  | 1,2                |
  | 1         | 1,2               | shift left clicks  | 2                  |

Scenario Outline: selection box selects enclosed objects
  Given edit mode is active
  And objects <inside_objects> are inside the selection box
  And objects <outside_objects> are outside the selection box
  And selection initially contains <initial_selection>
  When the coder drags an empty-space selection box with <modifier>
  Then selection should contain <expected_selection>

Examples:
  | inside_objects | outside_objects | initial_selection | modifier | expected_selection |
  | 1,2            | 3               | none              | none     | 1,2                |
  | 2              | 3               | 1                 | shift    | 1,2                |

Scenario Outline: middle drag moves selected objects
  Given edit mode is active
  And selected object <object_id> starts at <start_position>
  When the coder middle-drags selected objects by <drag_delta>
  Then object <object_id> position should be <expected_position>

Examples:
  | object_id | start_position | drag_delta | expected_position |
  | 1         | 10,10          | 5,-3       | 15,7              |

Scenario Outline: right drag throws or stops selected masses
  Given edit mode is active
  And selected mass <mass_id> fixed state is <fixed>
  When the coder right-drags selected masses with release velocity <release_velocity>
  Then mass <mass_id> velocity should be <expected_velocity>

Examples:
  | mass_id | fixed | release_velocity | expected_velocity |
  | 1       | false | 4,-2             | 4,-2              |
  | 2       | false | 0,0              | 0,0               |
  | 3       | true  | 4,-2             | unchanged         |
