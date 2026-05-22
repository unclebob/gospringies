# mutation-stamp: sha256=10ab42f9186df7571e3842ef3e4a55a2eee74baf626ab8caedd8f4e574b89ce0
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:01:39-05:00","feature_name":"XSP load and save","feature_path":"features/007_xsp_load_save.feature","background_hash":"efd05a36a2ac68ea601429cf02db036a71db541d227b873291b5a4110920c124","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"valid XSP files start with the supported format marker","scenario_hash":"438b7c6c27cc9e1b77ae1ef6c8acb0e225b02caa52b0c7b3a0c0a43f0e8b45ce","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:01:39-05:00"},{"index":1,"name":"supported commands load world state","scenario_hash":"a088e8e607ef4aa6090d91dd0e398614cc171e6da9d62184d3e684193b2bc400","mutation_count":16,"result":{"Total":16,"Killed":16,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:01:39-05:00"},{"index":2,"name":"saving is deterministic","scenario_hash":"ae358b93884ab54cee3a46641fa8b833570cf2b5d1af034a54b0980548c3a20a","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:01:39-05:00"},{"index":3,"name":"fixed masses round trip through negative file mass","scenario_hash":"8876c81c23d0524266ded609d6d8c1e97393758138da360be0a74bec2e74825b","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:01:39-05:00"},{"index":4,"name":"malformed XSP input reports useful errors","scenario_hash":"b318c787713a7ace9e7978f991c494d8125a6836edb92eef3d321e4a5a68df01","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:01:39-05:00"}]}
# acceptance-mutation-manifest-end
Feature: XSP load and save

Background:
  Given the XSP load and save task is accepted

Scenario Outline: valid XSP files start with the supported format marker
  Given XSP input starts with <marker>
  When the coder loads the XSP input
  Then loading should <result>

Examples:
  | marker | result |
  | #1.0   | pass   |
  | none   | fail   |

Scenario Outline: supported commands load world state
  Given XSP input contains command <command>
  When the coder loads the XSP input
  Then the loaded world should include <loaded_state>

Examples:
  | command | loaded_state        |
  | cmas    | current mass        |
  | elas    | current elasticity  |
  | kspr    | current spring k    |
  | kdmp    | current damping     |
  | frce    | force configuration |
  | wall    | wall configuration  |
  | mass    | mass                |
  | spng    | spring              |

Scenario Outline: saving is deterministic
  Given a world loaded from <input_file>
  When the coder saves the world twice
  Then both saved outputs should be identical
  And each saved output should end with a newline

Examples:
  | input_file   |
  | simple scene |

Scenario Outline: fixed masses round trip through negative file mass
  Given XSP input contains mass <mass_id> with file mass value <file_mass_value>
  When the coder loads and saves the XSP input
  Then mass <mass_id> fixed state should be <fixed>
  And saved mass <mass_id> should use file mass sign <file_mass_sign>

Examples:
  | mass_id | file_mass_value | fixed | file_mass_sign |
  | 1       | -3.0            | true  | negative       |

Scenario Outline: malformed XSP input reports useful errors
  Given XSP input has problem <problem>
  When the coder loads the XSP input
  Then loading should fail with reason <reason>

Examples:
  | problem                 | reason                   |
  | duplicate mass id       | duplicate id             |
  | duplicate spring id     | duplicate id             |
  | missing spring endpoint | missing spring endpoint  |
  | missing final newline   | missing final newline    |
