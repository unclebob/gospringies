# mutation-stamp: sha256=63612d7a928a3c17097ec08cf624d114acb479afc3fbe3ea0fbf35064d6d6959
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:18:24-05:00","feature_name":"Render visible controls","feature_path":"features/024_1_render_visible_controls.feature","background_hash":"4242c802dc13dc42f2484db31b454d989583a088c4aaac675810c1b384c35659","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"editor chrome regions produce visible pixels","scenario_hash":"99109b83f32032cfbea3b9631ef958f4289d993605cd435d3821af1239cd8713","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"},{"index":1,"name":"visible controls have readable labels","scenario_hash":"8790e7b5b969991b2675196bc4f9f6d2a997101ece349a2c2b311494870e8400","mutation_count":14,"result":{"Total":14,"Killed":14,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"},{"index":2,"name":"inspector sections are visibly rendered","scenario_hash":"d8cb961835694d4b30a558c7d4d9880200b3616e2fff2a424ee4703c3dcad679","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"},{"index":3,"name":"status fields are visibly rendered","scenario_hash":"943fc3a2b42a31cbe6208b7d21a7b19b57322d631dbaa0e8d82bc6366a21f16b","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"},{"index":4,"name":"world content remains visible with editor chrome","scenario_hash":"e1a100a9afc181fbbe5132c3e82983e564badc628bafc321f1cbaa1347992faa","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"},{"index":5,"name":"default window size has no clipped control labels","scenario_hash":"196e4c9c4601d1f854e95f34c687f0f58b3034e52c863f0ea55b7b841ac4decc","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:18:24-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Render visible controls

Background:
  Given the render visible controls task is accepted

Scenario Outline: editor chrome regions produce visible pixels
  When the coder draws the application frame
  Then screen region <region> should contain non-background pixels
  And screen region <region> should not contain only debug text

Examples:
  | region          |
  | left toolbar    |
  | top command bar |
  | right inspector |

Scenario Outline: visible controls have readable labels
  When the coder draws the application frame
  Then visible control <control> should have readable label <label>

Examples:
  | control       | label         |
  | run command   | Run           |
  | pause command | Pause         |
  | reset command | Reset         |
  | load command  | Load          |
  | insert command | Insert       |
  | save command  | Save          |
  | quit command  | Quit          |

Scenario Outline: inspector sections are visibly rendered
  When the coder draws the application frame
  Then inspector section <section> should be visible

Examples:
  | section                      |
  | ----- Selected Mass(es) -----  |
  | ----- Selected Spring(s) ----- |
  | ----- Forces -----             |
  | ----- Simulation -----         |
  | ----- Display -----            |

Scenario Outline: status fields are visibly rendered
  Given application state <state> is active
  When the coder draws the application frame
  Then status field <field> should be visible
  And status field <field> should show <state>

Examples:
  | state            | field          |
  | running          | run state      |
  | object counts    | object counts  |
  | saved            | file state     |

Scenario: world content remains visible with editor chrome
  When the coder draws the application frame
  Then the canvas should contain visible world content
  And editor chrome should not cover all world content

Scenario: default window size has no clipped control labels
  When the coder draws the application frame at the default window size
  Then visible control labels should fit inside their regions
