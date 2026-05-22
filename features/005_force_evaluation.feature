# mutation-stamp: sha256=6b7852b08f744093f38d13b60dff73a72d266bfce9fd7f7709faf9ddbc108f75
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T11:59:25-05:00","feature_name":"Force evaluation","feature_path":"features/005_force_evaluation.feature","background_hash":"8b3947b9227f6260de810ca0e0c351939670a821452e34fe8a2da19dec2db7e5","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"spring force is equal and opposite","scenario_hash":"6755a3c2f02c3e700b75bd7080e923fd5e435930fa8689654566653ae601de83","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:59:25-05:00"},{"index":1,"name":"spring damping acts along the spring direction","scenario_hash":"4dae27f31acfa521b57e1f3bfd1386dd521be42db919128ac71f475fb61ce9b5","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:59:25-05:00"},{"index":2,"name":"environmental forces can be evaluated independently","scenario_hash":"c79aff3ea2c7305e1f6534d610d80ed043680cbb2c0484cfb1584e641ad17525","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:59:25-05:00"},{"index":3,"name":"fixed masses do not accumulate acceleration","scenario_hash":"530ade0ea63e89c9acc553ecb2d0c1a614deb16aed15002909bc4715bc45d5c0","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:59:25-05:00"},{"index":4,"name":"wall force repels masses from inside boundaries","scenario_hash":"971b76612a53f8b7cc2d83deb4c2b461b9e6722e8fe28aad0c87dc9fe9c54b18","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:59:25-05:00"}]}
# acceptance-mutation-manifest-end
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

Scenario Outline: wall force repels masses from inside boundaries
  Given wall <wall> is enabled
  And mass <mass_id> is near the inside of the <wall> boundary
  When the coder evaluates forces without advancing time
  Then mass <mass_id> should receive force toward the inside of the world

Examples:
  | wall   | mass_id |
  | top    | 1       |
  | left   | 2       |
  | right  | 3       |
  | bottom | 4       |
