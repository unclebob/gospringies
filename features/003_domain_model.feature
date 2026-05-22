# mutation-stamp: sha256=2662680f33be29d465a157304c29182fd607a1d27b08de4459b4b66cdd86a934
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T11:57:10-05:00","feature_name":"Domain model","feature_path":"features/003_domain_model.feature","background_hash":"1e4fa602f36cfb97e67e01f5e3ae0d3cce331eb002b07773fe39ea98370d603d","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"a new world can be empty","scenario_hash":"d32bd36e88e82ee00e86a2fd89f6c34dcc597da7457f4147af074efc4674a547","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"},{"index":1,"name":"masses expose their modeled properties","scenario_hash":"03626ca6ee278c3bd3fef45af1e66f526d81ee157c4bc01d7c920ac0f0502df8","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"},{"index":2,"name":"springs expose their modeled properties","scenario_hash":"f3dc036153ec5d26331a67a280c2eebce5ef54e5b13fc01607bc65fd54068e65","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"},{"index":3,"name":"duplicate ids are invalid","scenario_hash":"00332095ab8319271bdaad1ce2931069821317a8674851e11af986e394c3e108","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"},{"index":4,"name":"springs require existing endpoint masses","scenario_hash":"25ece75aa6e188ff3c5f61fd513cbcab583073f8eebec58b8e106899cb30038e","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"},{"index":5,"name":"fixed mass is explicit domain state","scenario_hash":"ae7b67fac0df79645bfadd8bd34d1cb68abd7f3a92d43fe5e2041e44fad02cbb","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:57:10-05:00"}]}
# acceptance-mutation-manifest-end
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
