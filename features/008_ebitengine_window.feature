# mutation-stamp: sha256=e2f97e4f27f12f93ec51c5c210163397f84615d4fb0adfc22fc0647dd3923f5f
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:02:18-05:00","feature_name":"Ebitengine window","feature_path":"features/008_ebitengine_window.feature","background_hash":"78f7c3df071fc68a9cda0ed4034fec4614f52d2912a37bd8ef9300a5d6f7cb3e","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"the application opens without scene data","scenario_hash":"240b8b542b86752f8c469a85b87a781060f726e7e3de2dcc020badde9791e7b5","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:02:18-05:00"},{"index":1,"name":"the application window is resizable","scenario_hash":"614f419b7f39bc1ad48802b3290acac26cf5d64e42d361643ab0cabf40cf093f","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:02:18-05:00"},{"index":2,"name":"simulation pause state controls stepping","scenario_hash":"597f251575cf488cd679beacdbc78ee7edaa69424d98ce332066a54e187545dc","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:02:18-05:00"},{"index":3,"name":"closing the window exits cleanly","scenario_hash":"8013a5fdcbe006fa47d1066d101aba97c9760d628315a8bf694f9aa17f07b284","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:02:18-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Ebitengine window

Background:
  Given the Ebitengine window task is accepted

Scenario: the application opens without scene data
  When the coder starts the desktop application
  Then the application window should open successfully
  And the world should be empty

Scenario Outline: the application window is resizable
  When the coder resizes the application window to <window_size>
  Then the application should continue running

Examples:
  | window_size |
  | small       |
  | large       |

Scenario Outline: simulation pause state controls stepping
  Given the application simulation pause state is <paused>
  When the coder updates the application loop
  Then simulation stepping should be <stepping>
  And input handling should remain active
  And rendering should remain active

Examples:
  | paused | stepping |
  | true   | stopped  |
  | false  | active   |

Scenario: closing the window exits cleanly
  When the coder closes the application window
  Then the application should exit without error
