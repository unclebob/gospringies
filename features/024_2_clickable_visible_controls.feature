# mutation-stamp: sha256=543643013aa167a069a7b246e0734073fda34ee16e9d07811f3204b2878a0304
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:19:29-05:00","feature_name":"Clickable visible controls","feature_path":"features/024_2_clickable_visible_controls.feature","background_hash":"9272c4e929aeafa7a70b383e9eaa914f25e07b46a29e9c30c972617f75260997","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"clicking command controls runs commands","scenario_hash":"f0520e443963103d3016abd7c00d5661df2a307229c853fde3db2203294d8f24","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"},{"index":1,"name":"clicking Load opens the demo picker","scenario_hash":"ddfb8e1c04272cad9fb71ccaa005865634a8b320250c7823343fbac4d838fb12","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"},{"index":2,"name":"clicking path-based file controls opens keyboard path entry","scenario_hash":"9d1d2c654f80b0949e1c1f0228bef478861205a0ada6e9ed8fc6c1aef02cc67b","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"},{"index":3,"name":"clicked controls match keyboard shortcut behavior","scenario_hash":"0ffce9123a5cb47ab6dc18385667fd376f27fb0f058a13e14dc575dea23cdf74","mutation_count":12,"result":{"Total":12,"Killed":12,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"},{"index":4,"name":"clicking outside visible controls does nothing","scenario_hash":"0ead59667a48a3bd3703dcc7f61e13c7c0d16b85e1b928ccfd72052349ab4de1","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"},{"index":5,"name":"clicking run and pause controls changes simulation state","scenario_hash":"e574ab901bd08498144558dfbaa44da395af034e9549c8b10898975e6648debb","mutation_count":6,"result":{"Total":6,"Killed":6,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:19:29-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Clickable visible controls

Background:
  Given the clickable visible controls task is accepted

Scenario Outline: clicking command controls runs commands
  When the coder clicks inside rendered bounds of visible control <control>
  Then command <command> should run

Examples:
  | control | command |
  | Pause   | pause   |
  | Run     | run     |
  | Reset   | reset   |
  | Quit    | quit    |

Scenario: clicking Load opens the demo picker
  When the coder clicks inside rendered bounds of visible control Load
  Then the demo picker should open

Scenario Outline: clicking path-based file controls opens keyboard path entry
  When the coder clicks inside rendered bounds of visible control <control>
  Then keyboard path entry should open for <command>

Examples:
  | control | command |
  | Insert  | Insert  |
  | Save    | Save    |

Scenario Outline: clicked controls match keyboard shortcut behavior
  Given visible control <control> maps to shortcut <shortcut>
  When the coder clicks inside rendered bounds of visible control <control>
  Then the result should match pressing shortcut <shortcut>

Examples:
  | control | shortcut |
  | Pause   | Space    |
  | Reset   | R        |
  | Load    | Ctrl+O   |
  | Insert  | Ctrl+I   |
  | Save    | Ctrl+S   |
  | Quit    | Q        |

Scenario: clicking outside visible controls does nothing
  Given the application state is recorded
  When the coder clicks outside all visible controls
  Then the application state should remain unchanged

Scenario Outline: clicking run and pause controls changes simulation state
  Given simulation state is <old_state>
  When the coder clicks inside rendered bounds of visible control <control>
  Then simulation state should be <new_state>

Examples:
  | old_state | control | new_state |
  | running   | Pause   | paused    |
  | paused    | Run     | running   |
