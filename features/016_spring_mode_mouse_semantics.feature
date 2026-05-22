# mutation-stamp: sha256=6b66efe4a8c3105334e868500365c6a927c16e10ffa71ec2c1bdd9c04cdaf296
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:10:25-05:00","feature_name":"Spring mode mouse semantics","feature_path":"features/016_spring_mode_mouse_semantics.feature","background_hash":"42e3d546f815ef16f240bf1e11832d95c64beb45af7ff2c3c01122220909aad0","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"spring creation depends on release target","scenario_hash":"93f4e1f4c997bfcbde3fbd9ad3694c92aade23f3c25a1e88e28342f880a9ad59","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:10:25-05:00"},{"index":1,"name":"mouse button controls spring activation behavior","scenario_hash":"d60baef7699484f67d8d3d6b809e1d778b464d93f0dbac117458af68919d5c43","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:10:25-05:00"},{"index":2,"name":"created springs use defaults and creation length","scenario_hash":"b521fa9c2189e94754886b417ee7a621b9c080144f574913d72c8a8e577efe2b","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:10:25-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Spring mode mouse semantics

Background:
  Given the spring mode mouse semantics task is accepted

Scenario Outline: spring creation depends on release target
  Given spring mode is active
  And pointer press is near mass <start_mass>
  When the coder releases the pointer <release_target>
  Then spring creation should <result>

Examples:
  | start_mass | release_target | result                        |
  | 1          | near mass 2    | create spring between 1 and 2 |
  | 1          | away from mass | discard pending spring        |

Scenario Outline: mouse button controls spring activation behavior
  Given spring mode is active
  And pointer press is near mass <start_mass>
  When the coder drags with mouse button <button>
  Then the pending spring behavior should be <behavior>

Examples:
  | start_mass | button | behavior                                    |
  | 1          | left   | actively affects the first mass             |
  | 1          | middle | temporary cursor spring discarded on release |
  | 1          | right  | inactive until the spring is placed         |

Scenario Outline: created springs use defaults and creation length
  Given spring mode is active
  And current Kspring is <kspring>
  And current Kdamp is <kdamp>
  When the coder creates a spring with length <creation_length>
  Then the spring Kspring should be <kspring>
  And the spring Kdamp should be <kdamp>
  And the spring rest length should be <creation_length>

Examples:
  | kspring | kdamp | creation_length |
  | 12.0    | 0.5   | 30.0            |
