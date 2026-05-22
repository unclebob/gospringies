# mutation-stamp: sha256=4137f171c438af55ece444e863edf91df5931579d9629bc97c6b64af6fef4ddd
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:08:05-05:00","feature_name":"Demo Files","feature_path":"features/013_demo_files.feature","background_hash":"5a71fd8bbc5fc76e27aa92b8dfbd3e7e1b6e9221df6cfd98adb899ee26d2b2c0","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"Demo files are valid and human readable","scenario_hash":"abd7075c3cb2d0bf663984cb9f6cbc38fdf2fe69d7603892acb12b70be115003","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:05-05:00"},{"index":1,"name":"Demo files load successfully","scenario_hash":"12d427916dab2376ffbbe0730cce99ee33491fdc7c0a41577f41972e1cf49eb5","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:05-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Demo Files
  Demo files provide ready-to-load examples for manual exploration and regression coverage.

  Background:
    Given the demo files task is accepted

  Scenario: Demo files are valid and human readable
    When the coder adds demo file <demo_file>
    Then demo file <demo_file> should be valid XSP
    And demo file <demo_file> should be human readable

    Examples:
      | demo_file        |
      | pendulum.xsp     |
      | spring-chain.xsp |
      | small-mesh.xsp   |

  Scenario: Demo files load successfully
    Given demo file <demo_file> exists
    When the coder loads demo file <demo_file>
    Then the loaded world should include <required_feature>

    Examples:
      | demo_file        | required_feature |
      | pendulum.xsp     | fixed mass       |
      | spring-chain.xsp | multiple springs |
      | small-mesh.xsp   | multiple springs |
