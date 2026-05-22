# mutation-stamp: sha256=42a0ebe99993198a7de3a2f22bc7534970122d1b03d81a1ab46f46ccedaaae02
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:17:16-05:00","feature_name":"Nonblank startup editor 23.1","feature_path":"features/023_1_nonblank_startup_editor.feature","background_hash":"232b04b385d29c4cac47aaa83daba093bf70de4c4fa4649bce923dfedf4a356e","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"startup screen shows more than debug text","scenario_hash":"cdcbb4b18a18a3d04a2ec07c69e1756d47f819229215692a20826796126b9ef0","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:17:16-05:00"},{"index":1,"name":"editor chrome is visible on startup","scenario_hash":"b6e4a3ba2cf86e8c164fd09a9f424a5455ce0ce61e70f33af8b3a33c374ef52e","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:17:16-05:00"},{"index":2,"name":"startup world contains visible simulation objects","scenario_hash":"6a19eb65660fb80d77fa926d66730ba735957bec776a4ae2e274c285e7c927f6","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:17:16-05:00"},{"index":3,"name":"startup state is deterministic","scenario_hash":"63947ea2b1f3d80c46c6e5b3e11a30a53b34ad5487eb63a9041f61708262d571","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:17:16-05:00"},{"index":4,"name":"startup uses the default demo scene","scenario_hash":"bce27bf0f3654efa3fcb2b1bdf48d2c975b5da216b3841cfea40a599dce76e22","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:17:16-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Nonblank startup editor 23.1

Background:
  Given the nonblank startup editor 23.1 task is accepted

Scenario Outline: startup screen shows more than debug text
  When the coder starts the desktop application
  Then the first screen should show visible editor chrome
  And the first screen should show visible world content
  And debug text should not be the only visible content
  And the startup world should be loaded from <default_demo>

Examples:
  | default_demo       |
  | demos/pendulum.xsp |

Scenario Outline: editor chrome is visible on startup
  When the coder starts the desktop application
  Then startup screen region <region> should be visible

Examples:
  | region          |
  | canvas          |
  | left toolbar    |
  | top bar         |
  | right inspector |

Scenario Outline: startup world contains visible simulation objects
  When the coder starts the desktop application
  Then the startup world should contain <object_count> <object_type>

Examples:
  | object_type  | object_count |
  | fixed mass   | at least 1   |
  | movable mass | at least 1   |
  | spring       | at least 1   |

Scenario: startup state is deterministic
  When the coder starts the desktop application twice
  Then both startup worlds should be equivalent
  And both startup screens should show the same editor chrome

Scenario Outline: startup uses the default demo scene
  When the coder starts the desktop application
  Then the startup world should match demo file <default_demo>

Examples:
  | default_demo       |
  | demos/pendulum.xsp |
