# mutation-stamp: sha256=33ddc3dbebef42f8f1f22622a0c65871b7a0c8ebe582d1b2150b508f6b26725d
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
