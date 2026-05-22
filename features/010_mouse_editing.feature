# mutation-stamp: sha256=76be1526058d774a2bb98dcc8c98025e9de6f2afa824b6cc82a29023c498fcaf
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:05:42-05:00","feature_name":"Mouse editing","feature_path":"features/010_mouse_editing.feature","background_hash":"60b0db83657bdf035445ccab1581a3b40dea4a3467f9418757ca65531e64d22f","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"clicking in add mass mode creates a mass","scenario_hash":"5e3425e2ca59af8fbeec9a9ec3a57a2ea0c540cfebf409f013f68d260684ad37","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:05:42-05:00"},{"index":1,"name":"grid snap constrains mass placement to grid points","scenario_hash":"f785403d0a41b40d23770b26d1305c23e4b34d74b98a9c934c7d02a407ce7837","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:05:42-05:00"},{"index":2,"name":"grid snap constrains mass dragging to grid points","scenario_hash":"947e2a7c47d5cfde7dc641e55218634171411da0aaaf8fdada44a1b3018aef91","mutation_count":7,"result":{"Total":7,"Killed":7,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:05:42-05:00"},{"index":3,"name":"spring placement connects existing masses","scenario_hash":"bdd3c34c409ce0689ad71ee8cb0cc3aef1d19bb545312274e20643e8ff3fa94e","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:05:42-05:00"},{"index":4,"name":"dragging moves only movable masses","scenario_hash":"69e4cee54a344fe5de7a5db2a3f4cab97f249c2da7e64106f20309cd38959647","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:05:42-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Mouse editing

Background:
  Given the mouse editing task is accepted

Scenario Outline: clicking in add mass mode creates a mass
  Given the editor mode is <mode>
  And the current mass defaults are configured
  When the coder clicks at <pointer_position>
  Then a mass should be created at <expected_position>
  And the mass should use the current mass defaults

Examples:
  | mode     | pointer_position | expected_position |
  | add mass | 120,80           | 120,80            |

Scenario Outline: grid snap constrains mass placement to grid points
  Given grid snap is <grid_snap>
  And the grid snap size is <snap_size>
  And the editor mode is add mass
  When the coder clicks at <pointer_position>
  Then a mass should be created at <expected_position>
  And mass placement should be constrained to grid state <grid_snap>

Examples:
  | grid_snap | snap_size | pointer_position | expected_position |
  | enabled   | 10        | 123,87           | 120,90            |
  | disabled  | 10        | 123,87           | 123,87            |

Scenario Outline: grid snap constrains mass dragging to grid points
  Given grid snap is enabled
  And the grid snap size is <snap_size>
  And mass <mass_id> fixed state is false
  And mass <mass_id> starts at <start_position>
  When the coder drags mass <mass_id> through <drag_position> to <target_position>
  Then mass <mass_id> drag position should be <snapped_drag_position>
  And mass <mass_id> position should be <expected_position>

Examples:
  | mass_id | snap_size | start_position | drag_position | target_position | snapped_drag_position | expected_position |
  | 1       | 10        | 10,10          | 123,87        | 146,113         | 120,90                | 150,110           |

Scenario Outline: spring placement connects existing masses
  Given the editor mode is <mode>
  And mass <mass_a> exists
  And mass <mass_b> exists
  When the coder creates a spring from mass <mass_a> to mass <mass_b>
  Then a spring should connect mass <mass_a> to mass <mass_b>
  And the spring should use the current spring defaults

Examples:
  | mode       | mass_a | mass_b |
  | add spring | 1      | 2      |

Scenario Outline: dragging moves only movable masses
  Given mass <mass_id> fixed state is <fixed>
  And mass <mass_id> starts at <start_position>
  When the coder drags mass <mass_id> to <target_position>
  Then mass <mass_id> position should be <expected_position>
  And mass <mass_id> id should remain <mass_id>

Examples:
  | mass_id | fixed | start_position | target_position | expected_position |
  | 1       | false | 10,10          | 40,50           | 40,50             |
  | 2       | true  | 10,10          | 40,50           | 10,10             |
