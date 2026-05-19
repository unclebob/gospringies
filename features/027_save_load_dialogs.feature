# mutation-stamp: sha256=12e4ba906c97acdeba13b72286b7cff5e19a848ef1ee8702339efdf178ce15dc
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
