Feature: System parameters

Background:
  Given the system parameters task is accepted

Scenario Outline: a new world starts with editable defaults
  When the coder creates a new world
  Then parameter <parameter> should have default value <value>

Examples:
  | parameter       | value |
  | current mass    | set   |
  | elasticity      | set   |
  | spring constant | set   |
  | damping         | set   |
  | viscosity       | set   |
  | stickiness      | set   |
  | timestep        | set   |
  | precision       | set   |
  | grid snap       | set   |
  | show springs    | set   |

Scenario Outline: forces have editable configuration
  When the coder creates a new world
  Then force <force> should have enabled state <enabled>
  And force <force> should have editable parameters

Examples:
  | force                     | enabled |
  | gravity                   | set     |
  | center attraction         | set     |
  | center of mass attraction | set     |
  | wall repulsion            | set     |

Scenario Outline: walls have editable enabled state
  When the coder creates a new world
  Then wall <wall> should have enabled state <enabled>

Examples:
  | wall   | enabled |
  | top    | set     |
  | left   | set     |
  | right  | set     |
  | bottom | set     |

Scenario Outline: world operations preserve or replace parameters correctly
  Given a world with parameter <parameter> changed to <changed_value>
  When the coder performs <operation>
  Then parameter <parameter> should be <expected_value_source>

Examples:
  | parameter | changed_value | operation     | expected_value_source     |
  | viscosity | custom        | reset         | default value             |
  | timestep  | custom        | load file     | value from loaded file    |
  | damping   | custom        | insert file   | existing world value      |
