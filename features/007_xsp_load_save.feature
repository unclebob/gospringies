Feature: XSP load and save

Background:
  Given the XSP load and save task is accepted

Scenario Outline: valid XSP files start with the supported format marker
  Given XSP input starts with <marker>
  When the coder loads the XSP input
  Then loading should <result>

Examples:
  | marker | result |
  | #1.0   | pass   |
  | none   | fail   |

Scenario Outline: supported commands load world state
  Given XSP input contains command <command>
  When the coder loads the XSP input
  Then the loaded world should include <loaded_state>

Examples:
  | command | loaded_state        |
  | cmas    | current mass        |
  | elas    | current elasticity  |
  | kspr    | current spring k    |
  | kdmp    | current damping     |
  | frce    | force configuration |
  | wall    | wall configuration  |
  | mass    | mass                |
  | spng    | spring              |

Scenario Outline: saving is deterministic
  Given a world loaded from <input_file>
  When the coder saves the world twice
  Then both saved outputs should be identical
  And each saved output should end with a newline

Examples:
  | input_file   |
  | simple scene |

Scenario Outline: fixed masses round trip through negative file mass
  Given XSP input contains mass <mass_id> with file mass value <file_mass_value>
  When the coder loads and saves the XSP input
  Then mass <mass_id> fixed state should be <fixed>
  And saved mass <mass_id> should use file mass sign <file_mass_sign>

Examples:
  | mass_id | file_mass_value | fixed | file_mass_sign |
  | 1       | -3.0            | true  | negative       |

Scenario Outline: malformed XSP input reports useful errors
  Given XSP input has problem <problem>
  When the coder loads the XSP input
  Then loading should fail with reason <reason>

Examples:
  | problem                 | reason                   |
  | duplicate mass id       | duplicate id             |
  | duplicate spring id     | duplicate id             |
  | missing spring endpoint | missing spring endpoint  |
  | missing final newline   | missing final newline    |
