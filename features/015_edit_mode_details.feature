# mutation-stamp: sha256=7f9679b50d3253122fb6f1b961bc2b0d5159b639001160db10dec0234c883f97
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
