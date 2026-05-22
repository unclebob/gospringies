# mutation-stamp: sha256=db3f281a99a07dd4094d51e787c459da438d9afde6638e3a4c4f16cbee024320
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:00:34-05:00","feature_name":"Simulation step","feature_path":"features/006_simulation_step.feature","background_hash":"24ee7649a62297f60800ab748970844fd47fe8d2a4767a00cdac3223c48a296e","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"gravity changes movable mass state over time","scenario_hash":"35ec5c03fcaa15322ec5b99c7ba105e046838f476965ab42946be89a174e52fc","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:00:34-05:00"},{"index":1,"name":"fixed masses remain stationary","scenario_hash":"1cdc9e822af05a23711e3b34c2377414bdecea5239b05abbc760afef0bbf6d03","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:00:34-05:00"},{"index":2,"name":"advancing by duration is deterministic","scenario_hash":"6272e7858b3b4cd4a726938e60cd13357c6076c63a0514874a56bc3ea3343b16","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:00:34-05:00"},{"index":3,"name":"simulation stepping is independent of render frame rate","scenario_hash":"658bd99d6c54e5cf481a01a2d877b09678dd184a92dc6c84e1233aaddc007efd","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:00:34-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Simulation step

Background:
  Given the simulation step task is accepted

Scenario Outline: gravity changes movable mass state over time
  Given a movable mass starts at position <start_position>
  And gravity is enabled
  When the coder advances the simulation by <duration>
  Then the mass position should differ from <start_position>
  And the mass velocity should differ from <start_velocity>

Examples:
  | start_position | start_velocity | duration |
  | initial        | zero           | 1 step   |

Scenario Outline: fixed masses remain stationary
  Given mass <mass_id> fixed state is <fixed>
  And mass <mass_id> starts at position <start_position>
  When the coder advances the simulation by <duration>
  Then mass <mass_id> position should remain <start_position>
  And mass <mass_id> velocity should remain <start_velocity>
  And the coder looks up mass <id>
  And mass <id> fixed state should be <fixed>

Examples:
  | mass_id | id | fixed | start_position | start_velocity | duration |
  | 1       | 1  | true  | initial        | zero           | 10 steps |

Scenario Outline: advancing by duration is deterministic
  Given a world in state <initial_state>
  When the coder advances the simulation by <duration>
  Then the resulting state should be the same on every run

Examples:
  | initial_state   | duration |
  | simple spring   | 1 second |
  | gravity only    | 1 second |

Scenario Outline: simulation stepping is independent of render frame rate
  Given a world in state <initial_state>
  When the coder advances the simulation by <duration> using render frame rate <frame_rate>
  Then the resulting simulation time should be <duration>

Examples:
  | initial_state | duration | frame_rate |
  | gravity only  | 1 second | 30 fps     |
  | gravity only  | 1 second | 60 fps     |
