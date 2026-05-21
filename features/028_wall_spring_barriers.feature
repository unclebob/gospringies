# mutation-stamp: sha256=f5541c9d0e4ef017374f8977dd0c2a81bf0c2e16923165409637598ac9a17463
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

Scenario Outline: wall springs stop masses from crossing their segment
  Given wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>
  And moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder advances through wall spring collision
  Then mass <mass_id> should remain on the starting side of wall spring <spring_id>
  And mass <mass_id> velocity should be resolved away from wall spring <spring_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | mass_id | mass_x | mass_y | mass_vx | mass_vy |
  | 1         | 0       | 0       | 0       | 100     | 3       | -5     | 50     | 10      | 0       |

Scenario Outline: wall spring collision response is shared by endpoint masses
  Given wall spring <spring_id> spans from mass <endpoint_a> to mass <endpoint_b>
  And wall spring endpoint <endpoint_a> fixed state is <fixed_a>
  And wall spring endpoint <endpoint_b> fixed state is <fixed_b>
  And moving mass <mass_id> collides with wall spring <spring_id>
  When the coder resolves the wall spring collision
  Then wall spring endpoint <endpoint_a> should receive impulse share <impulse_share_a>
  And wall spring endpoint <endpoint_b> should receive impulse share <impulse_share_b>

Examples:
  | spring_id | endpoint_a | endpoint_b | fixed_a | fixed_b | mass_id | impulse_share_a | impulse_share_b |
  | 1         | 1          | 2          | false   | false   | 3       | half            | half            |
  | 1         | 1          | 2          | true    | false   | 3       | none            | half            |

Scenario Outline: wall attribute persists through XSP files
  Given XSP input contains spring <spring_id> with Wall value <input_wall>
  When the coder loads and saves the XSP input
  Then loaded spring <spring_id> should have Wall value <loaded_wall>
  And saved spring <spring_id> should include Wall value <saved_wall>

Examples:
  | spring_id | input_wall | loaded_wall | saved_wall |
  | 1         | true       | true        | true       |
  | 1         | absent     | false       | false      |

Scenario Outline: visible spring controls edit wall state
  Given selected spring <spring_id> has Wall value <old_wall>
  When the coder changes spring control Wall to <new_wall>
  Then spring <spring_id> should have Wall value <new_wall>

Examples:
  | spring_id | old_wall | new_wall |
  | 1         | false    | true     |
  | 1         | true     | false    |

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
  | 1         | true     | Wall      | false    |

Scenario Outline: wall springs render differently from normal springs
  Given spring <spring_id> has Wall value <wall>
  When the coder renders spring <spring_id>
  Then spring <spring_id> should use spring rendering style <rendering_style>

Examples:
  | spring_id | wall  | rendering_style |
  | 1         | false | normal          |
  | 1         | true  | wall            |
