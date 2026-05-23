# mutation-stamp: sha256=2fcdf6942fdb06756348933c5640ef918b500ae030b8c53f42443b1a060dba99
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-23T10:02:50-05:00","feature_name":"Wall spring physics","feature_path":"features/028_wall_spring_physics.feature","background_hash":"89cf4779ae04022daf21a3d55c396ddd402783642cfaf49907ee5a94913997e9","implementation_hash":"9f6de76b9cea6baf9caf45d16620768bf1502d7b12bf5acedb3602c53c824b27","scenarios":[{"index":0,"name":"wall state controls spring force behavior","scenario_hash":"a15c75938c45403634d9783ced93bc63553741efffd4dcd87f715bd3739bf173","mutation_count":18,"result":{"Total":18,"Killed":18,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:34:40-05:00"},{"index":1,"name":"wall springs keep a fixed endpoint length","scenario_hash":"c991e16b8eef4f74cfaab4b970f6261693f7b2ef4eceb52d936f711333fec799","mutation_count":27,"result":{"Total":27,"Killed":27,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:34:40-05:00"},{"index":2,"name":"wall springs stop masses from crossing their segment","scenario_hash":"5f3b1031ba886d7017b0777296ce4797210fba085be15526ae9f7f1a9897853e","mutation_count":10,"result":{"Total":10,"Killed":10,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:34:40-05:00"},{"index":3,"name":"wall springs collide with masses whose timestep path crosses their segment","scenario_hash":"6d37ac84efd7f67c9ab68bf8f0d8eb7e320cba12998c3ce984d660cabbab1c27","mutation_count":11,"result":{"Total":11,"Killed":11,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:34:40-05:00"},{"index":4,"name":"wall springs stop masses that start on the wall and move through it","scenario_hash":"7de8c9c8244f3ac601025840ed211deb116c112bd8549a6757739d2715fbe031","mutation_count":10,"result":{"Total":10,"Killed":10,"Survived":0,"Errors":0},"tested_at":"2026-05-22T16:20:43-05:00"},{"index":5,"name":"moving wall springs stop stationary masses from crossing their segment","scenario_hash":"0f617bfb80ad40a24bd7763ac802c9f51665ece7da4ee3d2d900bd9c5eeef420","mutation_count":10,"result":{"Total":10,"Killed":10,"Survived":0,"Errors":0},"tested_at":"2026-05-22T16:20:43-05:00"},{"index":6,"name":"moving floating wall springs stop masses that cross the swept segment","scenario_hash":"e3efeda863cb4bbb96b0031cad661a6059e49949b6debf3f3e2f3333cb4491cc","mutation_count":32,"result":{"Total":32,"Killed":32,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":7,"name":"moving wall springs bounce off fixed wall endpoints","scenario_hash":"eefe8d56c1ec8bb67131ddb7a46170c45390d38ed9e3de531c1073f355c292af","mutation_count":11,"result":{"Total":11,"Killed":11,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":8,"name":"wall spring length constraints cannot move endpoints through other wall springs","scenario_hash":"e90d8a61db04d2d0c5a18f9a8370aea468b86a0b62e39587461e85d25cfafb03","mutation_count":13,"result":{"Total":13,"Killed":13,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":9,"name":"wall spring collision response is shared by endpoint masses","scenario_hash":"b9872e7ea5e9bed8e280313a4054db9d30ea2e5709b3e851fccc1c4983f4bf62","mutation_count":27,"result":{"Total":27,"Killed":27,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":10,"name":"wall spring collision rebound uses configured elasticity","scenario_hash":"88d338141e8a2d3d26ab0a5e0ded0bbad6f7b21d47acf5a74f7d19bf6d37e89f","mutation_count":14,"result":{"Total":14,"Killed":14,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":11,"name":"floating wall collisions conserve momentum with unequal endpoint masses","scenario_hash":"6858e3464ec8df11a1c83c8a4ea3e7a6fe0f644aeb21c31d61fab8259d662412","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-23T09:23:52-05:00"},{"index":12,"name":"floating wall collisions do not add kinetic energy beyond elasticity","scenario_hash":"2fc47e65518f3ca046af2243623f2c01af1d97d59b29c0635da1c1ac8d908632","mutation_count":28,"result":{"Total":28,"Killed":28,"Survived":0,"Errors":0},"tested_at":"2026-05-23T10:02:50-05:00"},{"index":13,"name":"wall spring temperature kicks colliding masses","scenario_hash":"f1e37ca37ab588913471c06ee42709ed21ecf6971a8d6242cb380feda5b33078","mutation_count":12,"result":{"Total":12,"Killed":12,"Survived":0,"Errors":0},"tested_at":"2026-05-23T10:02:50-05:00"},{"index":14,"name":"non-wall spring temperature does not affect collisions","scenario_hash":"1d766c9fc43406bcf9a92371c55cdd881af21e2b86780b98b1f79e0d0d04f805","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-23T10:02:50-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Wall spring physics

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

Scenario Outline: wall springs stop masses that start on the wall and move through it
  Given wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>
  And moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder advances through wall spring collision
  Then mass <mass_id> should remain on the starting side of wall spring <spring_id>
  And mass <mass_id> velocity should be resolved away from wall spring <spring_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | mass_id | mass_x | mass_y | mass_vx | mass_vy |
  | 1         | 500     | 400     | 690     | 400     | 31      | 640    | 400    | 0       | -40     |

Scenario Outline: moving wall springs stop stationary masses from crossing their segment
  Given moving wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2> with velocity <wall_vx>, <wall_vy>
  And stationary mass <mass_id> starts at <mass_x>, <mass_y>
  When the coder advances through moving wall spring collision
  Then mass <mass_id> should remain on the starting side of moving wall spring <spring_id>
  And moving wall spring <spring_id> velocity should be resolved away from mass <mass_id>

Examples:
  | spring_id | wall_x1 | wall_y1 | wall_x2 | wall_y2 | wall_vx | wall_vy | mass_id | mass_x | mass_y |
  | 1         | -5      | 0       | -5      | 100     | 10      | 0       | 3       | 0      | 50     |

Scenario Outline: moving floating wall springs stop masses that cross the swept segment
  Given floating wall spring <spring_id> moves from <previous_wall_x1>, <previous_wall_y1> and <previous_wall_x2>, <previous_wall_y2> to <current_wall_x1>, <current_wall_y1> and <current_wall_x2>, <current_wall_y2>
  And moving mass <mass_id> moves from <previous_mass_x>, <previous_mass_y> to <current_mass_x>, <current_mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder resolves swept floating wall spring collision
  Then mass <mass_id> should remain on the previous side of floating wall spring <spring_id>
  And mass <mass_id> velocity should be resolved away from floating wall spring <spring_id>

Examples:
  | spring_id | previous_wall_x1 | previous_wall_y1 | previous_wall_x2 | previous_wall_y2 | current_wall_x1 | current_wall_y1 | current_wall_x2 | current_wall_y2 | mass_id | previous_mass_x | previous_mass_y | current_mass_x | current_mass_y | mass_vx | mass_vy |
  | 12        | 503.734          | 569.167          | 673.282          | 531.518          | 503.744         | 570.380         | 672.926         | 528.675         | 23      | 620             | 540             | 620            | 539.719        | 0       | -10     |
  | 12        | 503.734          | 569.167          | 673.282          | 531.518          | 503.744         | 570.380         | 672.926         | 528.675         | 26      | 620             | 510             | 620            | 509.719        | 0       | -10     |

Scenario Outline: moving wall springs bounce off fixed wall endpoints
  Given fixed mass <fixed_mass> at <fixed_x>, <fixed_y> is an endpoint of wall spring <fixed_spring>
  And moving wall spring <moving_spring> spans from <moving_x1>, <moving_y1> to <moving_x2>, <moving_y2> with velocity <moving_vx>, <moving_vy>
  When the simulation advances through fixed endpoint collision
  Then moving wall spring <moving_spring> should remain on the starting side of fixed endpoint mass <fixed_mass>
  And moving wall spring <moving_spring> contact point velocity should be resolved away from fixed endpoint mass <fixed_mass>

Examples:
  | fixed_spring | fixed_mass | fixed_x | fixed_y | moving_spring | moving_x1 | moving_y1 | moving_x2 | moving_y2 | moving_vx | moving_vy |
  | 1            | 1          | 0       | 0       | 2             | -10       | -5        | 10        | -5        | 0         | 10        |

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

Scenario Outline: wall spring collision rebound uses configured elasticity
  Given wall spring <spring_id> spans from mass <endpoint_a> to mass <endpoint_b>
  And moving mass <mass_id> with elasticity <elasticity> collides with wall spring <spring_id> at normal speed <normal_speed>
  When the coder resolves the wall spring collision
  Then mass <mass_id> normal rebound speed should be <expected_rebound_speed>
  And wall spring <spring_id> should receive collision impulse for rebound speed <expected_rebound_speed>

Examples:
  | spring_id | endpoint_a | endpoint_b | mass_id | elasticity | normal_speed | expected_rebound_speed |
  | 1         | 1          | 2          | 3       | 0.8        | 10           | 8                      |
  | 1         | 1          | 2          | 3       | 1.0        | 10           | 10                     |

Scenario Outline: floating wall collisions conserve momentum with unequal endpoint masses
  Given a stationary floating wall with endpoint masses <endpoint_a_mass> and <endpoint_b_mass>
  And moving mass <mass_id> with mass <moving_mass> is aimed at the floating wall from <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the simulation advances until the mass collides with the floating wall
  Then the total momentum of the moving mass and floating wall endpoints is unchanged

Examples:
  | endpoint_a_mass | endpoint_b_mass | mass_id | moving_mass | mass_x | mass_y | mass_vx | mass_vy |
  | 2               | 5               | 3       | 1           | -5     | 50     | 10      | 0       |

Scenario Outline: floating wall collisions do not add kinetic energy beyond elasticity
  Given moving floating wall spring <spring_id> has endpoint masses <endpoint_a_mass> and <endpoint_b_mass>
  And moving floating wall spring <spring_id> endpoint velocities are <endpoint_a_vx>, <endpoint_a_vy> and <endpoint_b_vx>, <endpoint_b_vy>
  And moving mass <mass_id> with mass <moving_mass> and elasticity <elasticity> collides with floating wall spring <spring_id> at contact fraction <contact_fraction> with velocity <mass_vx>, <mass_vy>
  When the coder resolves the finite-mass floating wall spring collision
  Then the total kinetic energy of mass <mass_id> and floating wall spring <spring_id> should be <energy_behavior>
  And the total momentum of mass <mass_id> and floating wall spring <spring_id> should be unchanged

Examples:
  | spring_id | endpoint_a_mass | endpoint_b_mass | endpoint_a_vx | endpoint_a_vy | endpoint_b_vx | endpoint_b_vy | mass_id | moving_mass | elasticity | contact_fraction | mass_vx | mass_vy | energy_behavior |
  | 12        | 1               | 1               | -3351.821     | 1287.498      | -322.615      | 129.493       | 32      | 1           | 0.8        | 0.03946          | -7.286  | 9.837   | not increased   |
  | 12        | 1               | 1               | 1000.008      | 103.366       | -87.331       | 64.323        | 24      | 1           | 0.8        | 0.05630          | -427.039 | -155.519 | not increased   |

Scenario Outline: floating wall contact keeps masses inside fixed walls
  Given fixed wall <fixed_wall> is enabled at boundary <wall_boundary>
  And moving floating wall spring <spring_id> has endpoint masses <endpoint_a_mass> and <endpoint_b_mass>
  And moving floating wall spring <spring_id> endpoint velocities are <endpoint_a_vx>, <endpoint_a_vy> and <endpoint_b_vx>, <endpoint_b_vy>
  And moving mass <mass_id> starts between floating wall spring <spring_id> and fixed wall <fixed_wall> at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>
  When the coder resolves simultaneous floating wall and fixed wall contact
  Then mass <mass_id> should remain inside fixed wall <fixed_wall>
  And mass <mass_id> should remain outside floating wall spring <spring_id>
  And the total kinetic energy of mass <mass_id> and floating wall spring <spring_id> should be <energy_behavior>

Examples:
  | fixed_wall | wall_boundary | spring_id | endpoint_a_mass | endpoint_b_mass | endpoint_a_vx | endpoint_a_vy | endpoint_b_vx | endpoint_b_vy | mass_id | mass_x | mass_y | mass_vx | mass_vy | energy_behavior |
  | bottom     | 400           | 12        | 1               | 1               | 0             | -85           | 0             | -85           | 26      | 620    | 403    | 0       | 0       | not increased   |

Scenario Outline: persistent floating wall penetration is resolved without restitution
  Given moving floating wall spring <spring_id> has endpoint masses <endpoint_a_mass> and <endpoint_b_mass>
  And moving floating wall spring <spring_id> endpoint velocities are <endpoint_a_vx>, <endpoint_a_vy> and <endpoint_b_vx>, <endpoint_b_vy>
  And moving mass <mass_id> starts <penetration> inside floating wall spring <spring_id> at contact fraction <contact_fraction> with relative normal velocity <relative_normal_velocity>
  When the coder resolves persistent floating wall spring contact
  Then mass <mass_id> should remain outside floating wall spring <spring_id>
  And mass <mass_id> relative normal velocity should be <normal_velocity_behavior>
  And the total kinetic energy of mass <mass_id> and floating wall spring <spring_id> should be <energy_behavior>

Examples:
  | spring_id | endpoint_a_mass | endpoint_b_mass | endpoint_a_vx | endpoint_a_vy | endpoint_b_vx | endpoint_b_vy | mass_id | penetration | contact_fraction | relative_normal_velocity | normal_velocity_behavior | energy_behavior |
  | 12        | 1               | 1               | 0             | -85           | 0             | -85           | 27      | 1.5         | 0.99             | -2                       | non-penetrating          | not increased   |
  | 12        | 1               | 1               | 0             | -85           | 0             | -85           | 32      | 2.0         | 0.12             | 0                        | unchanged                | not increased   |

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
