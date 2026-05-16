Feature: Demo files

Background:
  Given the demo files task is accepted

Scenario Outline: demo files are provided
  When the coder adds demo file <demo_file>
  Then demo file <demo_file> should be valid XSP
  And demo file <demo_file> should be human readable

Examples:
  | demo_file     |
  | pendulum      |
  | spring chain  |
  | small mesh    |

Scenario Outline: every demo file loads successfully
  Given demo file <demo_file> exists
  When the coder loads demo file <demo_file>
  Then loading should pass

Examples:
  | demo_file     |
  | pendulum      |
  | spring chain  |
  | small mesh    |

Scenario Outline: demos exercise required scene features
  Given demo file <demo_file> exists
  When the coder loads demo file <demo_file>
  Then the loaded world should include <required_feature>

Examples:
  | demo_file     | required_feature |
  | pendulum      | fixed mass       |
  | spring chain  | multiple springs |
  | small mesh    | multiple springs |
