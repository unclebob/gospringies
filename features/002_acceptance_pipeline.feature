# mutation-stamp: sha256=d06718928a25f275753d0506423e560a634a85bcc4541dcebda018298716358f
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T11:56:32-05:00","feature_name":"Acceptance pipeline","feature_path":"features/002_acceptance_pipeline.feature","background_hash":"c8db61d7612ec6fe7faba896194ef765291f973591585548eebf771d7beb78e8","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"acceptance tests run through the required pipeline stages","scenario_hash":"25295daaed4995b2593fc90389ca724e3822ce9bf1f0a298127eca0235cbc64b","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:56:32-05:00"},{"index":1,"name":"generated acceptance artifacts stay separate from hand-written tests","scenario_hash":"d08ac806cddbb069f9f35132784f0e3833271da79d713a839cb5684eefb55cb5","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:56:32-05:00"},{"index":2,"name":"a minimal smoke feature proves the pipeline works","scenario_hash":"6c53688a8f25ac29fa1ca9419be138c5a1459d6ff8e5dcc7711ecd4b5af08a01","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:56:32-05:00"},{"index":3,"name":"acceptance verification works from a clean checkout","scenario_hash":"bf3f032ed7ea0ad11925f693ef02fba1e7cba4ea74e3967175c105aaae4d1a23","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:56:32-05:00"},{"index":4,"name":"acceptance mutation stamps and skips verified features","scenario_hash":"04faeccf465d63a5d93eac75cfdd249ed77178ae71ebc2635910bb7bc0c6c223","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:56:32-05:00"}]}
# acceptance-mutation-manifest-end
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
