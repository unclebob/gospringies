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
