# mutation-stamp: sha256=0b9d4a032774253f3d71c67a1f03740bf81b7d85933bcee0384340c2005cbb1e
Feature: Wall spring barriers

Background:
  Given the wall spring barriers task is accepted

Scenario Outline: wall state controls spring force behavior
  Given spring <spring_id> connects mass <mass_a> to mass <mass_b>
  And spring <spring_id> has Wall value <wall>
  And spring <spring_id> has Kspring <kspring> Kdamp <kdamp> RestLen <rest_len>
  When the coder evaluates spring <spring_id> forces
  Then spring <spring_id> should apply spring force state <spring_force_state>
  And spring <spring_id> should apply damping force state <damping_force_state>

Examples:
  | spring_id | mass_a | mass_b | wall  | kspring | kdamp | rest_len | spring_force_state | damping_force_state |
  | 1         | 1      | 2      | false | 10      | 0.5   | 20       | enabled            | enabled             |
  | 1         | 1      | 2      | true  | 10      | 0.5   | 20       | disabled           | disabled            |

Scenario Outline: wall springs keep a fixed endpoint length
  Given wall spring <spring_id> endpoints start <initial_length> apart with RestLen <rest_len>
  And wall spring endpoint <endpoint_a> fixed state is <fixed_a>
  And wall spring endpoint <endpoint_b> fixed state is <fixed_b>
  When the coder advances wall spring length constraint
  Then wall spring <spring_id> endpoint distance should be <expected_length>
  And wall spring <spring_id> endpoint correction should be <correction_direction>

Examples:
  | spring_id | initial_length | rest_len | endpoint_a | endpoint_b | fixed_a | fixed_b | expected_length | correction_direction |
  | 1         | 120            | 100      | 1          | 2          | false   | false   | 100             | along segment        |
  | 1         | 80             | 100      | 1          | 2          | false   | false   | 100             | along segment        |
  | 1         | 120            | 100      | 1          | 2          | true    | false   | 100             | along segment        |

Scenario Outline: wall springs stop masses from crossing their segment
  Given wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>
  And moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder advances through wall spring collision
  Then mass <mass_id> should remain on the starting side of wall spring <spring_id>
  And mass <mass_id> velocity should be resolved away from wall spring <spring_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | mass_id | mass_x | mass_y | mass_vx | mass_vy |
  | 1         | 0       | 0       | 0       | 100     | 3       | -5     | 50     | 10      | 0       |

Scenario Outline: wall springs collide with masses whose timestep path crosses their segment
  Given wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>
  And fast moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder advances through wall spring collision by <duration>
  Then mass <mass_id> should remain on the starting side of wall spring <spring_id>
  And mass <mass_id> velocity should be resolved away from wall spring <spring_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | mass_id | mass_x | mass_y | mass_vx | mass_vy | duration |
  | 1         | 0       | 0       | 0       | 100     | 3       | -50    | 50     | 1000    | 0       | 1 step   |

Scenario Outline: moving wall springs stop stationary masses from crossing their segment
  Given moving wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2> with velocity <wall_vx>, <wall_vy>
  And stationary mass <mass_id> starts at <mass_x>, <mass_y>
  When the coder advances through moving wall spring collision
  Then mass <mass_id> should remain on the starting side of moving wall spring <spring_id>
  And moving wall spring <spring_id> velocity should be resolved away from mass <mass_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | wall_vx | wall_vy | mass_id | mass_x | mass_y |
  | 1         | -5      | 0       | -5      | 100     | 10      | 0       | 3       | 0      | 50     |

Scenario Outline: wall spring length constraints cannot move endpoints through other wall springs
  Given wall spring <barrier_spring> spans from <barrier_x1>, <barrier_y1> to <barrier_x2>, <barrier_y2>
  And constrained wall spring <moving_spring> endpoint <endpoint_a> starts at <endpoint_a_x>, <endpoint_a_y>
  And constrained wall spring <moving_spring> endpoint <endpoint_b> starts at <endpoint_b_x>, <endpoint_b_y>
  And constrained wall spring <moving_spring> has RestLen <rest_len>
  When the coder advances wall spring length constraints and collisions
  Then wall spring endpoint <endpoint_a> should remain on the starting side of wall spring <barrier_spring>
  And wall spring endpoint <endpoint_b> should remain on the starting side of wall spring <barrier_spring>

Examples:
  | barrier_spring | barrier_x1 | barrier_y1 | barrier_x2 | barrier_y2 | moving_spring | endpoint_a | endpoint_a_x | endpoint_a_y | endpoint_b | endpoint_b_x | endpoint_b_y | rest_len |
  | 1              | 0          | 0          | 0          | 100        | 2             | 3          | -5           | 40           | 4          | -80          | 40           | 150      |

Scenario Outline: wall spring collision response is shared by endpoint masses
  Given wall spring <spring_id> spans from mass <endpoint_a> to mass <endpoint_b>
  And wall spring endpoint <endpoint_a> fixed state is <fixed_a>
  And wall spring endpoint <endpoint_b> fixed state is <fixed_b>
  And moving mass <mass_id> collides with wall spring <spring_id> at contact fraction <contact_fraction>
  When the coder resolves the wall spring collision
  Then wall spring endpoint <endpoint_a> should receive impulse share <impulse_share_a>
  And wall spring endpoint <endpoint_b> should receive impulse share <impulse_share_b>

Examples:
  | spring_id | endpoint_a | endpoint_b | fixed_a | fixed_b | mass_id | contact_fraction | impulse_share_a | impulse_share_b |
  | 1         | 1          | 2          | false   | false   | 3       | 0.25             | 0.75            | 0.25            |
  | 1         | 1          | 2          | false   | false   | 3       | 0.50             | 0.50            | 0.50            |
  | 1         | 1          | 2          | true    | false   | 3       | 0.25             | absorbed        | 0.25            |

Scenario Outline: wall spring temperature kicks colliding masses
  Given wall spring <spring_id> has Temperature <temperature>
  And temperature random seed is <seed>
  And moving mass <mass_id> collides with wall spring <spring_id> at contact fraction <contact_fraction>
  When the coder resolves the wall spring collision
  Then mass <mass_id> should receive temperature kick <kick_behavior>

Examples:
  | spring_id | temperature | seed | mass_id | contact_fraction | kick_behavior                              |
  | 1         | 0           | 11   | 3       | 0.50             | none                                       |
  | 1         | 10          | 11   | 3       | 0.50             | full screen height against gravity 10     |

Scenario Outline: non-wall spring temperature does not affect collisions
  Given spring <spring_id> has Wall value false
  And spring <spring_id> has Temperature <temperature>
  And temperature random seed is <seed>
  And moving mass <mass_id> collides with spring <spring_id>
  When the coder resolves spring collision
  Then mass <mass_id> should receive temperature kick <kick_behavior>

Examples:
  | spring_id | temperature | seed | mass_id | kick_behavior |
  | 1         | 10          | 11   | 3       | none          |

Scenario Outline: wall attribute persists through XSP files
  Given XSP input contains spring <spring_id> with Wall value <input_wall>
  When the coder loads and saves the XSP input
  Then loaded spring <spring_id> should have Wall value <loaded_wall>
  And saved spring <spring_id> should include Wall value <saved_wall>

Examples:
  | spring_id | input_wall | loaded_wall | saved_wall |
  | 1         | true       | true        | true       |
  | 1         | absent     | false       | false      |

Scenario Outline: temperature attribute persists through XSP files
  Given XSP input contains spring <spring_id> with Temperature value <input_temperature>
  When the coder loads and saves the XSP input
  Then loaded spring <spring_id> should have Temperature value <loaded_temperature>
  And saved spring <spring_id> should include Temperature value <saved_temperature>

Examples:
  | spring_id | input_temperature | loaded_temperature | saved_temperature |
  | 1         | 7.5               | 7.5                | 7.5               |
  | 1         | absent            | 0                  | 0                 |

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
