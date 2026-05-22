# mutation-stamp: sha256=9cde9919676d958ffbfe27471ba7d5b8e2f815e42ebd4505cb6f4666d146d6bc
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:03:59-05:00","feature_name":"Screen and controls","feature_path":"features/008a_screen_and_controls.feature","background_hash":"d3c1d413f898fcba865e7511ddaac1dc9d64ac8b7df8287f4bc1209396824cdd","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"the first screen is the simulation editor","scenario_hash":"1ee47ae037d015955919c683aa0261ee59d7451c8d20162be0de78577df5e843","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"},{"index":1,"name":"the editor screen contains required regions","scenario_hash":"7f1323593035e3ce3d5d5e3ee8f691656728e803fa6148b3e788fd7ca63e6a36","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"},{"index":2,"name":"commands are visible controls","scenario_hash":"c153995f1f511ce84b1347d3de29a405e253f8ac6709552e945e9de6e8ad79d0","mutation_count":7,"result":{"Total":7,"Killed":7,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"},{"index":3,"name":"visible state reflects application state","scenario_hash":"bf88229ec1896dbf7fcfc9ab111460490f59f2d89f67b004d97f6c05ef12a152","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"},{"index":4,"name":"keyboard shortcuts mirror visible controls","scenario_hash":"31173bec542d6ae4d3031d6a22f538d643fb3b08a858298203bcadb53d4250c7","mutation_count":18,"result":{"Total":18,"Killed":18,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"},{"index":5,"name":"controls remain usable during simulation states","scenario_hash":"be9675a6184306924568d4c132e1a5582347b0bccf268415d6c3cf8be0853752","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:03:59-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Screen and controls

Background:
  Given the screen and controls task is accepted

Scenario: the first screen is the simulation editor
  When the coder starts the desktop application
  Then the first screen should show the simulation editor
  And the first screen should not show a landing page

Scenario Outline: the editor screen contains required regions
  When the coder lays out the editor screen
  Then screen region <region> should be visible
  And screen region <region> should have purpose <purpose>

Examples:
  | region        | purpose                                      |
  | canvas        | edit and view the simulation world           |
  | left toolbar  | run selection commands                       |
  | top bar       | run commands and file commands               |
  | right inspector | edit selected objects and world parameters and show simulation state |

Scenario Outline: commands are visible controls
  When the coder shows the top command bar
  Then command <command> should have a visible control

Examples:
  | command      |
  | pause toggle |
  | reset        |
  | load         |
  | insert       |
  | save         |
  | quit         |

Scenario Outline: visible state reflects application state
  Given application state <state> is active
  When the coder renders the editor controls
  Then visible indicator <indicator> should reflect <state>

Examples:
  | state           | indicator        |
  | paused          | simulation state |
  | running         | simulation state |
  | object selected | selection        |
  | unsaved changes | file state       |

Scenario Outline: keyboard shortcuts mirror visible controls
  Given command <command> has visible control <control>
  When the coder presses keyboard shortcut <shortcut>
  Then command <command> should run

Examples:
  | command      | control      | shortcut |
  | pause toggle | pause toggle | Space    |
  | reset        | reset        | R        |
  | save         | save         | Ctrl+S   |
  | load         | load         | Ctrl+O   |
  | insert       | insert       | Ctrl+I   |
  | quit    | quit    | Q        |

Scenario Outline: controls remain usable during simulation states
  Given simulation state is <simulation_state>
  When the coder renders the editor screen
  Then the canvas should remain visible
  And the visible controls should remain usable

Examples:
  | simulation_state |
  | paused           |
  | running          |
