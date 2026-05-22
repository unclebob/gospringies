# mutation-stamp: sha256=723b66cefc3441d7aa0b59e4ff0c262855247b85aabb947b488f724f7358de10
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T06:36:52-05:00","feature_name":"Wall spring editor controls","feature_path":"features/028_wall_spring_ui.feature","background_hash":"89cf4779ae04022daf21a3d55c396ddd402783642cfaf49907ee5a94913997e9","implementation_hash":"b26576c66147fff378f3b2547a7beb73116580539815405da08190ff9c09aa4b","scenarios":[{"index":0,"name":"visible spring controls edit wall state","scenario_hash":"80d74ee1214a5335debea065ca4ab6066ac27089df0d0374e217ce8657e138f9","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:36:52-05:00"},{"index":1,"name":"visible spring controls edit every selected spring wall state","scenario_hash":"a9aa550876a9ea67073f0d2cb1cc479a696b98861445ed803580ed06fac30afe","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:36:52-05:00"},{"index":2,"name":"spring right-click menu toggles wall state","scenario_hash":"206f85146ea00f13d6958786843a22f13f1e6def0d4dbc9c0938197d523c8219","mutation_count":24,"result":{"Total":24,"Killed":24,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:36:52-05:00"},{"index":3,"name":"spring right-click menu edits temperature with a slider dialog","scenario_hash":"17d94ff86530bb15ed445879f7f8b6b796bc6cfd8459a779631f048cd637db1e","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:36:52-05:00"},{"index":4,"name":"wall springs render differently from normal springs","scenario_hash":"0583b27af4b2825406c321e0af2b4beac2c7162c75ab834b70d70c2abe33f2e1","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:36:52-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Wall spring editor controls

Background:
  Given the wall spring barriers task is accepted

Scenario Outline: visible spring controls edit wall state
  Given selected spring <spring_id> has Wall value <old_wall>
  When the coder changes spring control Wall to <new_wall>
  Then spring <spring_id> should have Wall value <new_wall>

Examples:
  | spring_id | old_wall | new_wall |
  | 1         | false    | true     |
  | 1         | true     | false    |

Scenario Outline: visible spring controls edit every selected spring wall state
  Given selected springs <spring_ids> have Wall values <old_walls>
  When the coder changes spring control Wall to <new_wall>
  Then selected springs <spring_ids> should have Wall values <new_walls>

Examples:
  | spring_ids | old_walls          | new_wall | new_walls       |
  | 1, 2, 3    | false, false, true | true     | true, true, true |

Scenario Outline: spring right-click menu toggles wall state
  Given spring <spring_id> has Wall value <old_wall>
  And spring <spring_id> right-click menu includes item <menu_item>
  When the coder selects spring menu item Wall for spring <spring_id>
  Then spring <spring_id> should have Wall value <new_wall>

Examples:
  | spring_id | old_wall | menu_item | new_wall |
  | 1         | false    | Kspring   | true     |
  | 1         | false    | Kdamp     | true     |
  | 1         | false    | RestLen   | true     |
  | 1         | false    | Wall      | true     |
  | 1         | false    | Temperature | true   |
  | 1         | true     | Wall      | false    |

Scenario Outline: spring right-click menu edits temperature with a slider dialog
  Given spring <spring_id> has Temperature value <old_temperature>
  When the coder selects spring menu item Temperature for spring <spring_id>
  Then spring Temperature dialog should open with range <minimum> to <maximum>
  When the coder changes the spring Temperature dialog value to <new_temperature>
  Then spring <spring_id> should have Temperature value <new_temperature>

Examples:
  | spring_id | old_temperature | minimum | maximum | new_temperature |
  | 1         | 0               | 0       | 10      | 7.5             |

Scenario Outline: wall springs render differently from normal springs
  Given spring <spring_id> has Wall value <wall>
  When the coder renders spring <spring_id>
  Then spring <spring_id> should use spring rendering style <rendering_style>

Examples:
  | spring_id | wall  | rendering_style |
  | 1         | false | normal          |
  | 1         | true  | wall            |
