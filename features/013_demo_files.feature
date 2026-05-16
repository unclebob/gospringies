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
