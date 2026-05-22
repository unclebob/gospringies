# mutation-stamp: sha256=505e27be8b26965f39de49b3150b2335e380cc0a8ee2f95d0e35f9bbfa865a61
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:42:58-05:00","feature_name":"Off-canvas cleanup","feature_path":"features/029_off_canvas_cleanup.feature","background_hash":"a1063b7ba59c13868ebd2874284d4932b72ef9be255e421664dd149489b572ca","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"objects beyond one screen height from the canvas are deleted","scenario_hash":"7241bcc239d8d5e260f6db4ee57d6e962c9d23e500678d1528aa32fcd54603d6","mutation_count":14,"result":{"Total":14,"Killed":14,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:42:58-05:00"},{"index":1,"name":"objects at the cleanup boundary are retained","scenario_hash":"3e441050b97a65546f67a06c0e0d27068e12b99224fde0c23d2cf22940311d9d","mutation_count":13,"result":{"Total":13,"Killed":13,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:42:58-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Off-canvas cleanup

Background:
  Given the off-canvas cleanup task is accepted

Scenario Outline: objects beyond one screen height from the canvas are deleted
  Given a cleanup canvas with width <canvas_width> and height <canvas_height>
  And cleanup mass <mass_a> starts at <x_a>, <y_a>
  And cleanup mass <mass_b> starts at <x_b>, <y_b>
  And cleanup spring <spring_id> connects mass <spring_mass_a> to mass <spring_mass_b>
  When the coder advances off-canvas cleanup
  Then the cleanup world should contain <expected_mass_count> masses
  And the cleanup world should contain <expected_spring_count> springs
  And cleanup mass <remaining_mass> should remain present

Examples:
  | canvas_width | canvas_height | mass_a | x_a | y_a | mass_b | x_b | y_b | spring_id | spring_mass_a | spring_mass_b | remaining_mass | expected_mass_count | expected_spring_count |
  | 200          | 100           | 1      | 50  | 50  | 2      | 50  | 201 | 1         | 1             | 2             | 1              | 1                   | 0                     |

Scenario Outline: objects at the cleanup boundary are retained
  Given a cleanup canvas with width <canvas_width> and height <canvas_height>
  And cleanup mass <mass_a> starts at <x_a>, <y_a>
  And cleanup mass <mass_b> starts at <x_b>, <y_b>
  And cleanup spring <spring_id> connects mass <spring_mass_a> to mass <spring_mass_b>
  When the coder advances off-canvas cleanup
  Then the cleanup world should contain <expected_mass_count> masses
  And the cleanup world should contain <expected_spring_count> springs

Examples:
  | canvas_width | canvas_height | mass_a | x_a | y_a | mass_b | x_b | y_b | spring_id | spring_mass_a | spring_mass_b | expected_mass_count | expected_spring_count |
  | 200          | 100           | 1      | 50  | 50  | 2      | 50  | 200 | 1         | 1             | 2             | 2                   | 1                     |
