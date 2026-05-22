# mutation-stamp: sha256=c987334f92e00efee9ebb20229a49c113fbb5e4f329d02f0ecff225cb7e80c7b
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:11:41-05:00","feature_name":"Selected object parameter editing","feature_path":"features/018_selected_object_parameter_editing.feature","background_hash":"a4b2bcfa3706f8d04396089047a4c953f1abae4f00a3a8f416636ef4071f453e","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"mass controls update selected masses","scenario_hash":"575298335b22fb6e63bd2bf11322ad50a1925f4a6ee9a55d64b4ea90a540478f","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:41-05:00"},{"index":1,"name":"spring controls update selected springs","scenario_hash":"6005c5b1fe1ff45c20a8b85d10e49a41ffb1a257e65dd8147d8f44649a4782d4","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:41-05:00"},{"index":2,"name":"set rest length uses current spring length","scenario_hash":"300e6355101915e9d40e9723c6eb88aa854d1b44b06d0d5e19d9f9c7bddd83e4","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:41-05:00"},{"index":3,"name":"controls become creation defaults without compatible selection","scenario_hash":"1063732128fbbcee9e90bc40a268a3521f166375e54e00e29ca73d4108684fd5","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:11:41-05:00"}]}
# acceptance-mutation-manifest-end
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
