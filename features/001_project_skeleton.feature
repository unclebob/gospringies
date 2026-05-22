# mutation-stamp: sha256=ec5f2ecb6ced33f4852228e6c7840254c27d81809880bb1824ac73dff31073b2
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T11:55:55-05:00","feature_name":"Project skeleton","feature_path":"features/001_project_skeleton.feature","background_hash":"6e3a02b81889a15674c35e7f070019fa0995cc0234b6633c9248653e6450b23a","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"domain packages stay independent from the desktop graphics library","scenario_hash":"3414d4d190253be48ad8b8e253371b6bca917ab0fdf02ce25e43e498ec58d846","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:55:55-05:00"},{"index":1,"name":"the application command is buildable","scenario_hash":"09add9a8a84daa4a6aacb2a62b3ea363df0376447648157a220f83d603b09359","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:55:55-05:00"},{"index":2,"name":"the initial project test suite passes","scenario_hash":"c304210865c6c46137c94b4d58d9a5884b2a89ac402e126a0c8709a267989431","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T11:55:55-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Project skeleton

Background:
  Given the project skeleton task is accepted

Scenario Outline: domain packages stay independent from the desktop graphics library
  When the coder creates the initial Go package layout
  Then the <package> package should not import <graphics_library>

Examples:
  | package         | graphics_library |
  | simulation      | Ebitengine       |
  | file format     | Ebitengine       |

Scenario: the application command is buildable
  When the coder creates the desktop application command
  Then the application command should build successfully

Scenario: the initial project test suite passes
  When the coder creates the initial Go module
  Then the Go test suite should pass
