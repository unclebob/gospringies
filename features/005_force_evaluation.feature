Feature: Force evaluation

Background:
  Given the force evaluation task is accepted

Scenario Outline: spring force is equal and opposite
  Given mass <mass_a> is connected to mass <mass_b> by a spring
  And the spring has rest length <rest_length>
  And the spring has spring constant <spring_constant>
  When the coder evaluates forces without advancing time
  Then force on mass <mass_a> should be equal and opposite to force on mass <mass_b>

Examples:
  | mass_a | mass_b | rest_length | spring_constant |
  | 1      | 2      | 10.0        | 12.0            |

Scenario Outline: spring damping acts along the spring direction
  Given mass <mass_a> is connected to mass <mass_b> by a spring
  And mass <mass_a> has velocity <velocity_a>
  And mass <mass_b> has velocity <velocity_b>
  And the spring has damping constant <damping_constant>
  When the coder evaluates forces without advancing time
  Then spring damping should affect only the spring direction

Examples:
  | mass_a | mass_b | velocity_a | velocity_b | damping_constant |
  | 1      | 2      | moving     | still      | 0.5              |

Scenario Outline: environmental forces can be evaluated independently
  Given a world with force <force> enabled
  And a movable mass is affected by <force>
  When the coder evaluates forces without advancing time
  Then the mass should receive a force from <force>

Examples:
  | force                     |
  | gravity                   |
  | viscosity                 |
  | wall repulsion            |
  | center attraction         |
  | center of mass attraction |

Scenario Outline: fixed masses do not accumulate acceleration
  Given mass <mass_id> fixed state is <fixed>
  And mass <mass_id> is affected by force <force>
  When the coder evaluates forces without advancing time
  Then mass <mass_id> acceleration should be <acceleration>

Examples:
  | mass_id | fixed | force   | acceleration |
  | 1       | true  | gravity | zero         |

Scenario Outline: wall force pushes masses back into bounds
  Given wall <wall> is enabled
  And mass <mass_id> is outside the <wall> boundary
  When the coder evaluates forces without advancing time
  Then mass <mass_id> should receive force toward the inside of the world

Examples:
  | wall   | mass_id |
  | top    | 1       |
  | left   | 2       |
  | right  | 3       |
  | bottom | 4       |
