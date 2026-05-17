Feature: Adaptive RK4 numerics

Background:
  Given the adaptive RK4 numerics task is accepted

Scenario Outline: fixed timestep RK4 advances deterministically
  Given adaptive timestep is <adaptive>
  And time step is <time_step>
  When the coder advances the simulation by <duration>
  Then RK4 integration should advance deterministically by <duration>

Examples:
  | adaptive | time_step | duration |
  | false    | 0.01      | 1.0      |

Scenario Outline: adaptive timestep uses precision to choose smaller steps
  Given adaptive timestep is <adaptive>
  And precision is <precision>
  When the coder advances an unstable simulation by <duration>
  Then adaptive RK4 should choose <step_behavior>

Examples:
  | adaptive | precision | duration | step_behavior |
  | true     | low       | 1.0      | smaller steps |
  | true     | high      | 1.0      | larger steps  |

Scenario Outline: adaptive stepping preserves requested duration
  Given adaptive timestep is <adaptive>
  When the coder advances the simulation by <duration>
  Then simulation time should advance by <duration>

Examples:
  | adaptive | duration |
  | true     | 1.0      |

Scenario Outline: numerics remain independent of rendering frame rate
  Given render frame rate is <frame_rate>
  And adaptive timestep is <adaptive>
  When the coder advances the simulation by <duration>
  Then simulation state should not depend on render frame rate <frame_rate>

Examples:
  | frame_rate | adaptive | duration |
  | 30 fps     | false    | 1.0      |
  | 60 fps     | true     | 1.0      |
