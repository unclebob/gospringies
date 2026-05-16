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
  | start_mass | button | behavior                                     |
  | 1          | left   | actively affects the first mass              |
  | 1          | middle | temporary cursor spring discarded on release |
  | 1          | right  | inactive until the spring is placed          |

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
