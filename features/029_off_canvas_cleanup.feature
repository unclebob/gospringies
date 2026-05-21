# mutation-stamp: sha256=bdce169436a55773c6b4c5886bd153d6476b460cca9edfe740ce7661c350392e
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
