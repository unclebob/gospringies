Feature: Selected object parameter editing

Background:
  Given the selected object parameter editing task is accepted

Scenario Outline: mass controls update selected masses
  Given selected mass <mass_id> exists
  When the coder changes mass control <control> to <value>
  Then mass <mass_id> should have <control> value <value>

Examples:
  | mass_id | control    | value |
  | 1       | mass       | 2.0   |
  | 1       | elasticity | 0.5   |
  | 1       | fixed      | true  |

Scenario Outline: spring controls update selected springs
  Given selected spring <spring_id> exists
  When the coder changes spring control <control> to <value>
  Then spring <spring_id> should have <control> value <value>

Examples:
  | spring_id | control | value |
  | 1         | Kspring | 15.0  |
  | 1         | Kdamp   | 0.8   |

Scenario Outline: set rest length uses current spring length
  Given selected spring <spring_id> has current length <current_length>
  When the coder sets rest length
  Then spring <spring_id> rest length should be <current_length>

Examples:
  | spring_id | current_length |
  | 1         | 42.0           |

Scenario Outline: controls become creation defaults without compatible selection
  Given no selected object is compatible with control <control>
  When the coder changes control <control> to <value>
  Then future <object_type> objects should use <control> value <value>

Examples:
  | control | value | object_type |
  | mass    | 3.0   | mass        |
  | Kspring | 20.0  | spring      |
