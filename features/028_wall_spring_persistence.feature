# mutation-stamp: sha256=6e9495124f0615857a7d6e50226cad45a6c74f781b8f8358fd14508d35d396f1
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T06:20:44-05:00","feature_name":"Wall spring persistence","feature_path":"features/028_wall_spring_persistence.feature","background_hash":"89cf4779ae04022daf21a3d55c396ddd402783642cfaf49907ee5a94913997e9","implementation_hash":"b42605902311c09b4f78e2866b729f3ec71f09f0d874d72363170db0a13bd856","scenarios":[{"index":0,"name":"wall attribute persists through XSP files","scenario_hash":"4a88836ca80a58f0b7478a6e64651eaeace2591b399bce35b4d43c2e2e6460da","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:20:44-05:00"},{"index":1,"name":"temperature attribute persists through XSP files","scenario_hash":"20f33e2b76c12b1f0dbd1dee99191197bef443fa7bdf67df4d2a8990f724505b","mutation_count":8,"result":{"Total":8,"Killed":8,"Survived":0,"Errors":0},"tested_at":"2026-05-22T06:20:44-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Wall spring persistence

Background:
  Given the wall spring barriers task is accepted

Scenario Outline: wall attribute persists through XSP files
  Given XSP input contains spring <spring_id> with Wall value <input_wall>
  When the coder loads and saves the XSP input
  Then loaded spring <spring_id> should have Wall value <loaded_wall>
  And saved spring <spring_id> should include Wall value <saved_wall>

Examples:
  | spring_id | input_wall | loaded_wall | saved_wall |
  | 1         | true       | true        | true       |
  | 1         | absent     | false       | false      |

Scenario Outline: temperature attribute persists through XSP files
  Given XSP input contains spring <spring_id> with Temperature value <input_temperature>
  When the coder loads and saves the XSP input
  Then loaded spring <spring_id> should have Temperature value <loaded_temperature>
  And saved spring <spring_id> should include Temperature value <saved_temperature>

Examples:
  | spring_id | input_temperature | loaded_temperature | saved_temperature |
  | 1         | 7.5               | 7.5                | 7.5               |
  | 1         | absent            | 0                  | 0                 |
