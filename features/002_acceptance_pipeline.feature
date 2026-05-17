Feature: Acceptance pipeline

Background:
  Given the acceptance pipeline task is accepted

Scenario: acceptance tests run through the required pipeline stages
  When the coder runs the acceptance test command
  Then the Gherkin parser should run successfully
  And the acceptance test generator should run successfully
  And the generated executable acceptance tests should run successfully

Scenario Outline: generated acceptance artifacts stay separate from hand-written tests
  When the coder generates acceptance tests
  Then generated acceptance <artifact> should be written under <generated_location>
  And hand-written <test_type> tests should remain outside <generated_location>

Examples:
  | artifact        | generated_location   | test_type  |
  | test source     | acceptance/generated | unit       |
  | parsed feature  | build/acceptance     | unit       |

Scenario: a minimal smoke feature proves the pipeline works
  When the coder adds a minimal smoke feature
  Then the smoke feature should parse successfully
  And the smoke feature should generate an executable acceptance test
  And the generated smoke acceptance test should pass

Scenario: acceptance verification works from a clean checkout
  When the coder checks out the committed project
  Then the acceptance test command should pass without uncommitted setup steps

Scenario Outline: acceptance mutation stamps and skips verified features
  Given feature file <feature_file> has mutation stamp state <stamp_state>
  When the coder runs acceptance mutation for <feature_file>
  Then acceptance mutation should <mutation_behavior> <feature_file>
  And feature file <feature_file> should have mutation stamp state <expected_stamp_state>

Examples:
  | feature_file                     | stamp_state | mutation_behavior | expected_stamp_state |
  | features/pipeline_smoke.feature  | unstamped   | run and stamp      | stamped              |
  | features/pipeline_smoke.feature  | stamped     | skip               | stamped              |
