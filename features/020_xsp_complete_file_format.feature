# mutation-stamp: sha256=7d5029c34a4a7667a65ad5dda22af49bd81e035e2db3559a1a729da34d0c0ef3
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:14:55-05:00","feature_name":"XSP complete file format","feature_path":"features/020_xsp_complete_file_format.feature","background_hash":"5be46197ba2f3a13c8a7337de4ec8f07c61fb660802203970c74ba923996b228","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"documented commands load and save","scenario_hash":"6ad6196a69693066f38ae878f36d95d3e6932749c375b27d5f552e51f84f5953","mutation_count":17,"result":{"Total":17,"Killed":17,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:14:55-05:00"},{"index":1,"name":"named force tokens are stable in saved XSP","scenario_hash":"feb720b36904b65b51d068f153dee9211d65aa9fd3cda467d5b9f6d4db65e954","mutation_count":16,"result":{"Total":16,"Killed":16,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:14:55-05:00"},{"index":2,"name":"file format validation rejects documented invalid structure","scenario_hash":"227f77e3bde959ba9dc650c8d5e32cfb0623e2a3fd76bd7a27928e39edb631ff","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:14:55-05:00"},{"index":3,"name":"file operations normalize file paths","scenario_hash":"598f81163988ab9b6e35c25d022a436171f2ecddf380cd894c2e1e8bf1c4e7b7","mutation_count":9,"result":{"Total":9,"Killed":9,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:14:55-05:00"},{"index":4,"name":"insert loads only objects","scenario_hash":"e9981d374d558c88519652f07a0e6c74b73baac9b5ba3764b8e8fe13971c8562","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:14:55-05:00"}]}
# acceptance-mutation-manifest-end
Feature: XSP complete file format

Background:
  Given the XSP complete file format task is accepted

Scenario Outline: documented commands load and save
  Given XSP input contains command <command>
  When the coder loads and saves the XSP input
  Then saved XSP output should include command <command>

Examples:
  | command |
  | cmas    |
  | elas    |
  | kspr    |
  | kdmp    |
  | fixm    |
  | shws    |
  | cent    |
  | frce    |
  | visc    |
  | stck    |
  | step    |
  | prec    |
  | adpt    |
  | gsnp    |
  | wall    |
  | mass    |
  | spng    |

Scenario Outline: named force tokens are stable in saved XSP
  Given world force <force_name> is configured with <enabled_state> and <force_parameters>
  When the coder saves the world as XSP
  Then saved XSP output should include force token <force_token>
  When the coder loads XSP input with force token <force_token>
  Then loaded force token <force_token> should map to force <force_name>
  Then loaded force <force_name> should be configured with <enabled_state> and <force_parameters>

Examples:
  | force_name                | force_token               | enabled_state | force_parameters       |
  | center attraction         | center-attraction         | false         | magnitude=0 exponent=2 |
  | center of mass attraction | center-of-mass-attraction | false         | magnitude=0 damping=0  |
  | wall repulsion            | wall-repulsion            | false         | magnitude=0 exponent=2 |
  | mass collision            | mass-collision            | false         | none                   |

Scenario Outline: file format validation rejects documented invalid structure
  Given XSP input has problem <problem>
  When the coder loads the XSP input
  Then loading should fail with reason <reason>

Examples:
  | problem              | reason                |
  | blank line           | blank lines not allowed |
  | missing final newline | missing final newline |
  | non-positive id      | ids must be positive  |

Scenario Outline: file operations normalize file paths
  Given filename input is <filename>
  And environment variable SPRINGDIR is <springdir>
  When the coder resolves an XSP filename
  Then resolved filename should be <resolved_filename>

Examples:
  | filename | springdir | resolved_filename        |
  | demo     | unset     | demo.xsp                 |
  | demo.xsp | unset     | demo.xsp                 |
  | demo     | examples  | examples/demo.xsp        |

Scenario Outline: insert loads only objects
  Given current parameters are <current_parameters>
  When the coder inserts XSP file <input_file>
  Then inserted masses and springs should be added
  And parameters should remain <current_parameters>

Examples:
  | current_parameters | input_file |
  | custom             | complete   |
