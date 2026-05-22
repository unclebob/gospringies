# mutation-stamp: sha256=f853bbef73cb10e27fbf66709fa37c5d1bddb330fe80fc52be42303f2a2d3e06
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:22:11-05:00","feature_name":"Original style human interface","feature_path":"features/024_original_style_human_interface.feature","background_hash":"77c79884bffeb08e710ef969e5ef421319e8204858b5349b08804640e7b86c73","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"the interface is custom-drawn in Ebitengine","scenario_hash":"af835c600052a40bd73842f4611ce5e5c796b531d885815a886575a91454e712","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"},{"index":1,"name":"original command controls are visible and clickable","scenario_hash":"d12c3d3a7fdbb910980a08b96ffa0c9be9830223a2d33b5e15236a18768d2609","mutation_count":9,"result":{"Total":9,"Killed":9,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"},{"index":2,"name":"inspector controls expose editable settings","scenario_hash":"411be94c5ec4998279bd85e17ace54f7ccb93be24d316b7b893b9aea826edbd4","mutation_count":16,"result":{"Total":16,"Killed":16,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"},{"index":3,"name":"right inspector reports current application state","scenario_hash":"2d6e8d0d65662aa69e6ebab1e6d9e1da0e1aee3d01a439d9c5ccf236cafc81f1","mutation_count":14,"result":{"Total":14,"Killed":14,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"},{"index":4,"name":"file commands use keyboard path entry","scenario_hash":"46840125deddebdf143ea5245953a130c62758721405779df274353230af1b59","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"},{"index":5,"name":"visible controls mirror keyboard shortcuts","scenario_hash":"c323b73d18ef16730614f89f6ab5433a29365a0a8045c130d39f6c004c067b4c","mutation_count":18,"result":{"Total":18,"Killed":18,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:22:11-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Original style human interface

Background:
  Given the original style human interface task is accepted

Scenario Outline: the interface is custom-drawn in Ebitengine
  When the coder renders the editor interface
  Then interface element <element> should be drawn by the Ebitengine app

Examples:
  | element         |
  | canvas          |
  | left toolbar    |
  | top command bar |
  | right inspector |

Scenario Outline: original command controls are visible and clickable
  When the coder renders the top command bar
  Then command control <command> should be visible
  And command control <command> should be clickable

Examples:
  | command       |
  | Run           |
  | Pause         |
  | Reset         |
  | Save State    |
  | Restore State |
  | Load          |
  | Insert        |
  | Save          |
  | Quit          |

Scenario Outline: inspector controls expose editable settings
  When the coder renders the right inspector
  Then inspector control <control> should be visible
  And inspector control <control> should be editable

Examples:
  | control                    |
  | Mass                       |
  | Elasticity                 |
  | Fixed Mass                 |
  | Kspring                    |
  | Kdamp                      |
  | Set Rest Length            |
  | Gravity                    |
  | Center Attraction          |
  | Center Of Mass Attraction  |
  | Wall Repulsion             |
  | Wall Toggles               |
  | Grid Snap                  |
  | Show Springs               |
  | Time Step                  |
  | Precision                  |
  | Adaptive Time Step         |

Scenario Outline: right inspector reports current application state
  Given application state <state> is active
  When the coder renders the right inspector
  Then status field <field> should show <state>

Examples:
  | state              | field                 |
  | running            | run state             |
  | paused             | run state             |
  | object counts      | object counts         |
  | selected count     | selected object count |
  | current file path  | current file          |
  | unsaved changes    | dirty state           |
  | file error message | last error            |

Scenario Outline: file commands use keyboard path entry
  When the coder activates file command <command>
  Then keyboard path entry should open for <command>
  When the coder submits path <path>
  Then file command <command> should use path <path>

Examples:
  | command | path              |
  | Load    | demos/pendulum.xsp |
  | Insert  | demos/spring-chain.xsp |
  | Save    | out/current.xsp   |

Scenario Outline: visible controls mirror keyboard shortcuts
  Given visible control <control> invokes command <command>
  When the coder presses shortcut <shortcut>
  Then command <command> should run

Examples:
  | control | command | shortcut |
  | Pause   | Pause   | Space    |
  | Reset   | Reset   | R        |
  | Load    | Load    | Ctrl+O   |
  | Insert  | Insert  | Ctrl+I   |
  | Save    | Save    | Ctrl+S   |
  | Quit    | Quit    | Q        |
