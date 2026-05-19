# mutation-stamp: sha256=f40c2adfbeeecf6ab9334674e82dabe6d514a47785e15739c50825469e43813d
Feature: Domain model

Background:
  Given the domain model task is accepted

Scenario Outline: a new world can be empty
  When the coder creates a new world
  Then the world should contain <mass_count> masses
  And the world should contain <spring_count> springs

Examples:
  | mass_count | spring_count |
  | 0          | 0            |

Scenario Outline: masses expose their modeled properties
  Given a world with mass <id> at <x>, <y>
  And mass <id> has velocity <vx>, <vy>
  And mass <id> has mass value <mass_value>
  And mass <id> has elasticity <elasticity>
  And mass <id> fixed state is <fixed>
  When the coder looks up mass <id>
  Then mass <id> should have position <x>, <y>
  And mass <id> should have velocity <vx>, <vy>
  And mass <id> should have mass value <mass_value>
  And mass <id> should have elasticity <elasticity>
  And mass <id> fixed state should be <fixed>

Examples:
  | id | x    | y    | vx  | vy   | mass_value | elasticity | fixed |
  | 1  | 10.0 | 20.0 | 0.0 | 0.0  | 1.0        | 0.8        | false |
  | 2  | 30.0 | 40.0 | 1.5 | -2.0 | 2.5        | 0.4        | true  |

Scenario Outline: springs expose their modeled properties
  Given a world with mass <mass_a> at <x_a>, <y_a>
  And a world with mass <mass_b> at <x_b>, <y_b>
  And a spring <spring_id> connects mass <mass_a> to mass <mass_b>
  And spring <spring_id> has spring constant <spring_constant>
  And spring <spring_id> has damping constant <damping_constant>
  And spring <spring_id> has rest length <rest_length>
  When the coder looks up spring <spring_id>
  Then spring <spring_id> should connect mass <mass_a> to mass <mass_b>
  And spring <spring_id> should have spring constant <spring_constant>
  And spring <spring_id> should have damping constant <damping_constant>
  And spring <spring_id> should have rest length <rest_length>

Examples:
  | spring_id | mass_a | mass_b | x_a | y_a | x_b  | y_b  | spring_constant | damping_constant | rest_length |
  | 7         | 1      | 2      | 0.0 | 0.0 | 10.0 | 0.0  | 12.5            | 0.7              | 10.0        |
  | 8         | 2      | 3      | 1.0 | 2.0 | 5.0  | 8.0  | 20.0            | 1.3              | 7.2         |

Scenario Outline: duplicate ids are invalid
  Given a world already contains a <object_type> with id <id>
  When the coder adds another <object_type> with id <id>
  Then validation should fail with reason <reason>

Examples:
  | object_type | id | reason       |
  | mass        | 1  | duplicate id |
  | spring      | 5  | duplicate id |

Scenario Outline: springs require existing endpoint masses
  Given a world with mass <existing_mass> at <x>, <y>
  When the coder adds spring <spring_id> connecting mass <mass_a> to mass <mass_b>
  Then validation should fail with reason <reason>

Examples:
  | existing_mass | x   | y   | spring_id | mass_a | mass_b | reason                   |
  | 1             | 0.0 | 0.0 | 2         | 1      | 9      | missing spring endpoint  |
  | 1             | 0.0 | 0.0 | 3         | 8      | 1      | missing spring endpoint  |

Scenario Outline: fixed mass is explicit domain state
  Given a world with mass <id> at <x>, <y>
  And mass <id> has mass value <mass_value>
  And mass <id> fixed state is <fixed>
  When the coder reads mass <id> from the domain model
  Then mass <id> fixed state should be <fixed>
  And mass <id> mass value should remain <mass_value>

Examples:
  | id | x   | y   | fixed | mass_value |
  | 4  | 5.0 | 6.0 | true  | 3.0        |
