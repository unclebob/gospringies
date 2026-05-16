Feature: Force center and force parameters

Background:
  Given the force center and force parameters task is accepted

Scenario Outline: forces expose documented parameters
  When the coder selects force <force>
  Then force <force> should expose parameter <parameter_one>
  And force <force> should expose parameter <parameter_two>

Examples:
  | force                     | parameter_one | parameter_two |
  | gravity                   | Magnitude     | Direction     |
  | center of mass attraction | Magnitude     | Damping       |
  | center attraction         | Magnitude     | Exponent      |
  | wall repulsion            | Magnitude     | Exponent      |

Scenario Outline: gravity direction uses XSpringies degrees
  Given gravity direction is <direction_degrees>
  When the coder evaluates gravity
  Then gravity should point <expected_direction>

Examples:
  | direction_degrees | expected_direction |
  | 0.0               | down               |
  | 90.0              | right              |
  | 180.0             | up                 |
  | 270.0             | left               |

Scenario Outline: set center chooses selected mass or screen center
  Given selected masses are <selected_masses>
  When the coder sets the force center
  Then force center should be <expected_center>

Examples:
  | selected_masses | expected_center |
  | none            | screen center   |
  | 1               | mass 1          |

Scenario Outline: center behavior is visible and non-reciprocal
  Given force center is mass <center_mass>
  And force <force> is enabled
  When the coder evaluates center forces
  Then mass <center_mass> should be visually marked as the center
  And mass <center_mass> should not receive reciprocal force response from <force>

Examples:
  | center_mass | force                     |
  | 1           | center attraction         |
  | 1           | center of mass attraction |

Scenario Outline: enabling a force selects its parameter controls
  When the coder enables force <force>
  Then parameter controls for force <force> should be active

Examples:
  | force             |
  | gravity           |
  | wall repulsion    |
