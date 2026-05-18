Feature: Mass collision

Background:
  Given the mass collision task is accepted

Scenario Outline: free masses rebound from each other
  Given a collision world with mass <mass_a> at <x_a>, <y_a> moving <vx_a>, <vy_a>
  And collision mass <mass_a> has mass value <mass_value_a> elasticity <elasticity_a> fixed state <fixed_a>
  And collision mass <mass_b> at <x_b>, <y_b> moving <vx_b>, <vy_b>
  And collision mass <mass_b> has mass value <mass_value_b> elasticity <elasticity_b> fixed state <fixed_b>
  And mass collision is enabled
  When the coder advances through mass collision
  Then collision mass <mass_a> velocity should be <expected_vx_a>, <expected_vy_a>
  And collision mass <mass_b> velocity should be <expected_vx_b>, <expected_vy_b>

Examples:
  | mass_a | x_a | y_a | vx_a | vy_a | mass_value_a | elasticity_a | fixed_a | mass_b | x_b | y_b | vx_b | vy_b | mass_value_b | elasticity_b | fixed_b | expected_vx_a | expected_vy_a | expected_vx_b | expected_vy_b |
  | 1      | 0   | 0   | 1    | 0    | 1            | 1            | false   | 2      | 2   | 0   | -1   | 0    | 1            | 1            | false   | -1            | 0             | 1             | 0             |

Scenario Outline: fixed masses do not move during collision
  Given a collision world with mass <mass_a> at <x_a>, <y_a> moving <vx_a>, <vy_a>
  And collision mass <mass_a> has mass value <mass_value_a> elasticity <elasticity_a> fixed state <fixed_a>
  And collision mass <mass_b> at <x_b>, <y_b> moving <vx_b>, <vy_b>
  And collision mass <mass_b> has mass value <mass_value_b> elasticity <elasticity_b> fixed state <fixed_b>
  And mass collision is enabled
  When the coder advances through mass collision
  Then collision mass <mass_a> velocity should be <expected_vx_a>, <expected_vy_a>
  And collision mass <mass_b> velocity should be <expected_vx_b>, <expected_vy_b>

Examples:
  | mass_a | x_a | y_a | vx_a | vy_a | mass_value_a | elasticity_a | fixed_a | mass_b | x_b | y_b | vx_b | vy_b | mass_value_b | elasticity_b | fixed_b | expected_vx_a | expected_vy_a | expected_vx_b | expected_vy_b |
  | 1      | 0   | 0   | 1    | 0    | 1            | 1            | false   | 2      | 2   | 0   | 0    | 0    | 1            | 1            | true    | -1            | 0             | 0             | 0             |

Scenario Outline: disabled mass collision leaves overlapping masses unchanged
  Given a collision world with mass <mass_a> at <x_a>, <y_a> moving <vx_a>, <vy_a>
  And collision mass <mass_a> has mass value <mass_value_a> elasticity <elasticity_a> fixed state <fixed_a>
  And collision mass <mass_b> at <x_b>, <y_b> moving <vx_b>, <vy_b>
  And collision mass <mass_b> has mass value <mass_value_b> elasticity <elasticity_b> fixed state <fixed_b>
  When the coder advances through mass collision
  Then collision mass <mass_a> velocity should be <expected_vx_a>, <expected_vy_a>
  And collision mass <mass_b> velocity should be <expected_vx_b>, <expected_vy_b>

Examples:
  | mass_a | x_a | y_a | vx_a | vy_a | mass_value_a | elasticity_a | fixed_a | mass_b | x_b | y_b | vx_b | vy_b | mass_value_b | elasticity_b | fixed_b | expected_vx_a | expected_vy_a | expected_vx_b | expected_vy_b |
  | 1      | 0   | 0   | 1    | 0    | 1            | 1            | false   | 2      | 2   | 0   | -1   | 0    | 1            | 1            | false   | 1             | 0             | -1            | 0             |
