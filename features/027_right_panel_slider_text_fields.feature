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
