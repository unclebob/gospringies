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

Examples:
  | mass_id | fixed | start_position | start_velocity | duration |
  | 1       | true  | initial        | zero           | 10 steps |

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
