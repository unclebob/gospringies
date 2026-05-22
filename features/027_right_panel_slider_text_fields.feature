# mutation-stamp: sha256=3434687a58037fd814eca9f46861bbd5b30d4e05134bdf2048d2a0acb1ecd611
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:33:58-05:00","feature_name":"Right panel slider text fields","feature_path":"features/027_right_panel_slider_text_fields.feature","background_hash":"ff46d35663aad2e69ad03c3949c0dceae2c3d98c38924843d67ef525cf253d93","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"right panel numeric settings use sliders with text fields","scenario_hash":"661162eb27b3a73e51157f63d4b257459bbced5288d652a745bfa5423254e661","mutation_count":26,"result":{"Total":26,"Killed":26,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"},{"index":1,"name":"right panel numeric controls fit without overlap","scenario_hash":"672c11976de752a489d68ef8bf3f0d09425b08e21faa3deca3d782cf535a55ba","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"},{"index":2,"name":"slider text fields use a compact row layout","scenario_hash":"4a08318721508842b6e09b4cd3a05d24c7dfaad592f5026e5a8585a228afbe8d","mutation_count":26,"result":{"Total":26,"Killed":26,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"},{"index":3,"name":"stickiness is adjusted with a slider","scenario_hash":"052d97582ed4684442459ea8a5d135412e790a94e0fbb8f6548f02cf0a82035b","mutation_count":1,"result":{"Total":1,"Killed":1,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"},{"index":4,"name":"slider text fields accept direct keyboard entry","scenario_hash":"09871f509ea75d82d871c2a415565d1d61ad037ba94a28f0fc695168038a8393","mutation_count":9,"result":{"Total":9,"Killed":9,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"},{"index":5,"name":"slider and text field stay synchronized","scenario_hash":"13ee3fde8362e3d19edfdf6006ef061ecce12451cd3795f8e9663b6550a245e5","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:33:58-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Right panel slider text fields

Background:
  Given the right panel slider text fields task is accepted

Scenario Outline: right panel numeric settings use sliders with text fields
  When the coder renders the right inspector
  Then numeric setting <setting> should have a visible slider
  And numeric setting <setting> should have a visible text field
  And numeric setting <setting> text field should show value <value>

Examples:
  | setting                   | value |
  | Mass                      | 1.0   |
  | Elasticity                | 0.8   |
  | Kspring                   | 12.0  |
  | Kdamp                     | 0.4   |
  | Gravity                   | 10.0  |
  | Center Attraction         | 0.0   |
  | Center Of Mass Attraction | 0.0   |
  | Wall Repulsion            | 0.0   |
  | Viscosity                 | 0.0   |
  | Stick                     | 0.0   |
  | Speed                     | 1.0   |
  | Time Step                 | 0.016 |
  | Precision                 | 0.001 |

Scenario: right panel numeric controls fit without overlap
  When the coder renders the right inspector
  Then every numeric setting label should fit inside the right inspector
  And every numeric setting slider should fit inside the right inspector
  And every numeric setting text field should fit inside the right inspector
  And numeric setting controls should not overlap other numeric setting controls
  And numeric setting controls should not overlap right inspector section headings

Scenario Outline: slider text fields use a compact row layout
  When the coder renders numeric setting <setting>
  Then numeric setting <setting> label should be left of its slider
  And numeric setting <setting> text field should be right of its slider
  And numeric setting <setting> slider and text field should be on the same row
  And numeric setting <setting> text field should fit value <longest_value>

Examples:
  | setting                   | longest_value |
  | Mass                      | 1000.0        |
  | Elasticity                | 1.0           |
  | Kspring                   | 1000.0        |
  | Kdamp                     | 1000.0        |
  | Gravity                   | 1000.0        |
  | Center Attraction         | 1000.0        |
  | Center Of Mass Attraction | 1000.0        |
  | Wall Repulsion            | 100000.0      |
  | Viscosity                 | 1000.0        |
  | Stick                     | 1000.0        |
  | Speed                     | 10.0          |
  | Time Step                 | 0.0001        |
  | Precision                 | 0.000001      |

Scenario Outline: stickiness is adjusted with a slider
  When the coder changes numeric setting Stick with the slider to <new_value>
  Then parameter stickiness should be <new_value>
  And numeric setting Stick text field should show value <new_value>

Examples:
  | new_value |
  | 5.0       |

Scenario Outline: slider text fields accept direct keyboard entry
  Given numeric setting <setting> has value <old_value>
  When the coder focuses numeric setting <setting> text field
  Then numeric setting <setting> text field cursor should blink
  When the coder enters text value <new_value>
  Then numeric setting <setting> should have value <new_value>
  And numeric setting <setting> slider should show value <new_value>

Examples:
  | setting   | old_value | new_value |
  | Stick     | 0.0       | 5.0       |
  | Gravity   | 10.0      | 3.5       |
  | Time Step | 0.016     | 0.02      |

Scenario Outline: slider and text field stay synchronized
  Given numeric setting <setting> has value <old_value>
  When the coder changes numeric setting <setting> with the slider to <new_value>
  Then numeric setting <setting> text field should show value <new_value>
  When the coder enters text value <final_value>
  Then numeric setting <setting> slider should show value <final_value>

Examples:
  | setting | old_value | new_value | final_value |
  | Stick   | 0.0       | 2.5       | 7.5         |
