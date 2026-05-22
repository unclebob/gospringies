# mutation-stamp: sha256=a79d52380b6b4862a0ac2263b5b8e0b1e752dd0254d5083e2bc90ad4d7b39dc3
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T11:58:50-05:00","feature_name":"System parameters","feature_path":"features/004_system_parameters.feature","background_hash":"73214ec4478360c15b673ff8a9859bade38d2458639899fd020d3a642a8b2c40","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"a new world starts with editable defaults","scenario_hash":"d98a885948f5e16027a559cd189c23fdbee7de4a190f38449e9d82611ce75530","mutation_count":20,"result":{"Total":20,"Killed":20,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:58:50-05:00"},{"index":1,"name":"forces have editable configuration","scenario_hash":"9772886322cd53950f79281a751bb52ec1f1d5e398a30b35962e137321b863be","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:58:50-05:00"},{"index":2,"name":"walls have editable enabled state","scenario_hash":"25c6132476af0d2ed906d3e7d29b6087cee3ed53971372ec06fac9f2986813c5","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:58:50-05:00"},{"index":3,"name":"world operations preserve or replace parameters correctly","scenario_hash":"b8dfd74f504fe32f5ac62551b3447d89110f34bee5823154b9a7633d60c4e8d5","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:58:50-05:00"}]}
# acceptance-mutation-manifest-end
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
