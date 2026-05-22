# mutation-stamp: sha256=213f8b86b4d183a89ae5edf79f93e5f2cd5e1c5621e555347398169130b3d31d
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:35:04-05:00","feature_name":"Save and load dialogs","feature_path":"features/027_save_load_dialogs.feature","background_hash":"70e11bf3ef60420d3e73b4822bbf5bbbf797c30abb407e57c353370f29f80c8a","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"save asks for a filename before writing","scenario_hash":"da824eaf740d07de592f0b479643734e8863f57df53837ece9298cb1951d11a0","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"},{"index":1,"name":"save writes named files under saves","scenario_hash":"11e4152727248e9f0b4e1e87d816e641899f08477fd2d4b43820e11c7acad8c1","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"},{"index":2,"name":"saved files load back from the picker","scenario_hash":"5b5ef7ecaf744f45ed09e544ec1c50976f9f87e0fc8b2b56bc05561ea60cc939","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"},{"index":3,"name":"load picker groups saved files before packaged files","scenario_hash":"9eb324363f3283be6bb56ee67ab6416f0492b4b8d704fdab0e84c8462f95a1d3","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"},{"index":4,"name":"load picker reads saved files from saves","scenario_hash":"5b09de0d69e0cc1fc581e734f66b5be1c8bf8d5f2877540990a05013174a8c00","mutation_count":3,"result":{"Total":3,"Killed":3,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"},{"index":5,"name":"load picker refreshes saved files each time it opens","scenario_hash":"69e4453869c5dd26e2e8d07667aa77ae51346a081834fbbd8e4242d552a05409","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:35:04-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Save and load dialogs

Background:
  Given the save and load dialogs task is accepted

Scenario Outline: save asks for a filename before writing
  When the coder clicks inside rendered bounds of visible control Save
  Then save filename dialog should open
  And save filename field should contain <field_text>
  And save filename cursor should be positioned <cursor_position>

Examples:
  | field_text | cursor_position      |
  | .xsp       | before xsp extension |

Scenario Outline: save writes named files under saves
  Given the current world contains <world_state>
  When the coder enters save filename prefix <filename_prefix>
  And the coder submits the save filename dialog
  Then saved XSP file should exist at <saved_path>
  And saved XSP file <saved_path> should contain <world_state>

Examples:
  | world_state   | filename_prefix | saved_path             |
  | simple masses | lab_scene       | saves/lab_scene.xsp    |

Scenario Outline: saved files load back from the picker
  Given the current world contains <world_state>
  When the coder enters save filename prefix <filename_prefix>
  And the coder submits the save filename dialog
  And the coder opens the load picker
  And the coder chooses load picker entry <save_file>
  Then loaded world should include <world_state>
  And current file path should be <saved_path>

Examples:
  | world_state   | filename_prefix | save_file      | saved_path             |
  | simple masses | simple hex      | simple hex.xsp | saves/simple hex.xsp   |

Scenario Outline: load picker groups saved files before packaged files
  Given saved XSP file <save_file> exists in saves
  And demo XSP file <demo_file> exists in demos
  And original XSP file <original_file> exists in demos/original
  When the coder clicks inside rendered bounds of visible control Load
  Then load picker should show <save_file> before <separator>
  And load picker should show <separator> before <demo_file>
  And load picker should show <demo_file> before <original_file>

Examples:
  | save_file     | separator | demo_file    | original_file |
  | lab_scene.xsp | separator | pendulum.xsp | pend.xsp      |

Scenario Outline: load picker reads saved files from saves
  Given saved XSP file <save_file> exists in saves with <world_state>
  When the coder opens the load picker
  And the coder chooses load picker entry <save_file>
  Then loaded world should include <world_state>
  And current file path should be <saved_path>

Examples:
  | save_file     | world_state   | saved_path           |
  | lab_scene.xsp | simple masses | saves/lab_scene.xsp  |

Scenario Outline: load picker refreshes saved files each time it opens
  Given saved XSP file <old_save_file> exists in saves
  When the coder opens the load picker
  And saved XSP file <new_save_file> exists in saves with <world_state>
  And the coder opens the load picker
  Then load picker should show <new_save_file> before <separator>
  When the coder chooses load picker entry <new_save_file>
  Then loaded world should include <world_state>
  And current file path should be <saved_path>

Examples:
  | old_save_file | new_save_file  | world_state   | separator | saved_path              |
  | lab_scene.xsp | simple hex.xsp | simple masses | separator | saves/simple hex.xsp    |
