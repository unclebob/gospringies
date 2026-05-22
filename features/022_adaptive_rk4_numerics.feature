# mutation-stamp: sha256=673c438298d9c8b4754a851b54c3a6efcb13e8ab3a2574909e9fcecbbe34e3ea
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:16:36-05:00","feature_name":"Adaptive RK4 numerics","feature_path":"features/022_adaptive_rk4_numerics.feature","background_hash":"09918a33e20854b1010e3880c5dcec3b86b1e4cbe25e30becbd6f1539eae7c50","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"fixed timestep RK4 advances deterministically","scenario_hash":"d8685b3d02b87776661b77714ccdd56b1a80e87fd4756a61d9fd803b967ca1b2","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:36-05:00"},{"index":1,"name":"adaptive timestep uses precision to choose smaller steps","scenario_hash":"954c89b3a06c0b88a42ae0e111310ca35b555742c9521256a0e3b01bb1c8e403","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:36-05:00"},{"index":2,"name":"adaptive stepping preserves requested duration","scenario_hash":"35d551e076a3d7eb181c42774a7cc8897baa6ef908f7227a00344f1eeb0b6852","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:36-05:00"},{"index":3,"name":"numerics remain independent of rendering frame rate","scenario_hash":"f0c0d02a7f4558d9c2ac987632faf4cce1c60652eb824cb96c2c6c50c58fa1dd","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:36-05:00"}]}
# acceptance-mutation-manifest-end
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
