# mutation-stamp: sha256=4d2e026646e65295a2b254fa5e15d887df8190a063d95bafbd0db96d518f77f1
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:16:01-05:00","feature_name":"Force center and force parameters","feature_path":"features/021_force_center_and_force_parameters.feature","background_hash":"7f440a992441c50e73f4aec9cb761ff4ca9b3b57910907c8ed52554d10f8a188","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"forces expose documented parameters","scenario_hash":"11a7da83eaad17e75f7ff3ec3a0a6f1401cbb9bf7598cd8d01f827b370a4c112","mutation_count":12,"result":{"Total":12,"Killed":12,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:01-05:00"},{"index":1,"name":"gravity direction uses XSpringies degrees","scenario_hash":"9ec100087972f19f215b8a4aa0662e7eb3bc79f3b533f1e7ac33d3fc64ad7cce","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:01-05:00"},{"index":2,"name":"set center chooses selected mass or screen center","scenario_hash":"d3361c87a0a49266cdc64bb920756d0d8dabf6a417f853a30496f354c5358b81","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:01-05:00"},{"index":3,"name":"center behavior is visible and non-reciprocal","scenario_hash":"f5bca64aa207446e44d70c3471289f491415459e97c8155eb7089ddfae3472b6","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:01-05:00"},{"index":4,"name":"enabling a force selects its parameter controls","scenario_hash":"eea6ffbdc5ba6a5bf83acd7ccdfac50f8f439b8fbfda61d503731ba1daa9a810","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:16:01-05:00"}]}
# acceptance-mutation-manifest-end
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
