# mutation-stamp: sha256=1af7dcc185fd51ce9c8904fc71737358287c9ad9dafe97ca2d93131480bf055a
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:31:23-05:00","feature_name":"Mass collision","feature_path":"features/026_mass_collision.feature","background_hash":"1e8dcfd36956f9d20b18b6996e8d4174aff8b093af7646800a253784108c6745","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"free masses rebound from each other","scenario_hash":"8feaef0c2e6eab722545f3b4c9006d1dac63b004538befa95295cd28c1ed1ffc","mutation_count":20,"result":{"Total":20,"Killed":20,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:31:23-05:00"},{"index":1,"name":"fixed masses do not move during collision","scenario_hash":"6c6d86b61380e78a0d919343456fccb300c14858a7b7d4f0ba2b77a85841dba1","mutation_count":20,"result":{"Total":20,"Killed":20,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:31:23-05:00"},{"index":2,"name":"disabled mass collision leaves overlapping masses unchanged","scenario_hash":"b744afda533e1fe1231f2e3d226cad1fdbb6e780b3196759d71cdd791b8efd28","mutation_count":20,"result":{"Total":20,"Killed":20,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:31:23-05:00"}]}
# acceptance-mutation-manifest-end
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
