# mutation-stamp: sha256=b3f485e68bb336761847627114728ef1377210f0754cba8921a25f94a53cf872
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:12:51-05:00","feature_name":"Wall collision and stickiness","feature_path":"features/019_wall_collision_and_stickiness.feature","background_hash":"8945f0994c3ea2dc7587cab6ca118b6b9de2a133875f3ce069761ea267409620","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"enabled walls bounce masses moving toward them","scenario_hash":"5f201fdcf52480c0bb161f9baa2121b425ef1425ab45064b703d4714313300a5","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:12:51-05:00"},{"index":1,"name":"one-way walls allow outside masses to enter","scenario_hash":"da283fee5d2a8d2e886f3fefd0e891c5c8896a8536fcf5de456479aa4ced12f0","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:12:51-05:00"},{"index":2,"name":"enabled walls collide with masses whose timestep path crosses the boundary","scenario_hash":"6e1b47a2588096e8d9c38985329e7c4de7341bbd8e3ea17e3fdc2039ae4d3c05","mutation_count":12,"result":{"Total":12,"Killed":12,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:12:51-05:00"},{"index":3,"name":"stickiness can hold and release a mass","scenario_hash":"489d7592492d95d3b28438bd1a0e4fd638fbf307575885369ba0e3b8559cf8d0","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:12:51-05:00"},{"index":4,"name":"disabled walls do not collide","scenario_hash":"41d1c098de2e772fc8a3a0fe3ce641dfbee1279357cf7350830daefd4876667c","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:12:51-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Wall collision and stickiness

Background:
  Given the wall collision and stickiness task is accepted

Scenario Outline: enabled walls bounce masses moving toward them
  Given wall <wall> is enabled
  And mass <mass_id> has elasticity <elasticity>
  And mass <mass_id> moves from inside the screen toward wall <wall>
  When the coder advances through the wall collision
  Then mass <mass_id> wall-normal velocity should be reversed
  And mass <mass_id> wall-normal velocity magnitude should be scaled by <elasticity>

Examples:
  | wall | mass_id | elasticity |
  | left | 1       | 0.5        |
  | top  | 2       | 1.0        |

Scenario Outline: one-way walls allow outside masses to enter
  Given wall <wall> is enabled
  And mass <mass_id> moves from off-screen toward the screen through wall <wall>
  When the coder advances through the wall boundary
  Then mass <mass_id> should pass through wall <wall>

Examples:
  | wall  | mass_id |
  | left  | 1       |
  | right | 2       |

Scenario Outline: enabled walls collide with masses whose timestep path crosses the boundary
  Given wall <wall> is enabled
  And mass <mass_id> starts at <start_x>, <start_y> with velocity <velocity_x>, <velocity_y>
  When the coder advances through the wall boundary by <duration>
  Then mass <mass_id> should remain on the starting side of wall <wall>
  And mass <mass_id> wall-normal velocity should be resolved toward the starting side of wall <wall>

Examples:
  | wall  | mass_id | start_x | start_y | velocity_x | velocity_y | duration |
  | right | 1       | 790     | 400     | 300        | 0          | 1 step   |
  | top   | 2       | 400     | 590     | 0          | 300        | 1 step   |

Scenario Outline: stickiness can hold and release a mass
  Given stickiness is <stickiness>
  And mass <mass_id> collides with wall <wall>
  When the wall collision removes all wall-normal velocity
  Then mass <mass_id> should stick to wall <wall>
  When force <release_force> pulls mass <mass_id> away from wall <wall>
  Then mass <mass_id> should be <release_result>

Examples:
  | stickiness | mass_id | wall | release_force | release_result |
  | high       | 1       | left | insufficient  | stuck          |
  | high       | 1       | left | sufficient    | released       |

Scenario Outline: disabled walls do not collide
  Given wall <wall> is disabled
  And mass <mass_id> moves toward wall <wall>
  When the coder advances through the wall boundary
  Then mass <mass_id> should not bounce from wall <wall>

Examples:
  | wall   | mass_id |
  | bottom | 1       |
