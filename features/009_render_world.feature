# mutation-stamp: sha256=eba79e9d07897494c536dce95468b248ce0d11ee64c1cdf3809307db98f68d12
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:04:33-05:00","feature_name":"Render world","feature_path":"features/009_render_world.feature","background_hash":"d24bcfa6208d8f4c75b63fe521709cbc8934ac25caa1f9c908abf9520f5b3e46","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"worlds render without crashing","scenario_hash":"328cf2671a7280647ee7b02e86e0a701792f6a740d706963dffeca148fd4f70b","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:04:33-05:00"},{"index":1,"name":"renderable objects are visible","scenario_hash":"841a269f5235aee9e84919102a9e173234497a7ba6d6cfb258b7cacddabf282f","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:04:33-05:00"},{"index":2,"name":"show springs controls spring visibility","scenario_hash":"836c1265d91886a44ee50b438d3a3d0c1b715d1b363ce6cf8caf8d80055fb87f","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:04:33-05:00"},{"index":3,"name":"fixed masses are visually distinguishable","scenario_hash":"9f205dd1c8e5c4182dfc100a1fc3e30828d096431511d53e1367ff558d83bfb0","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:04:33-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Render world

Background:
  Given the render world task is accepted

Scenario Outline: worlds render without crashing
  Given the application has <world_state>
  When the coder renders the world
  Then rendering should complete successfully

Examples:
  | world_state      |
  | an empty world   |
  | a non-empty world |

Scenario Outline: renderable objects are visible
  Given the world contains <object>
  When the coder renders the world
  Then <object> should have a visible representation

Examples:
  | object        |
  | movable mass  |
  | fixed mass    |
  | spring        |
  | enabled wall  |
  | selection     |

Scenario Outline: show springs controls spring visibility
  Given the world contains a spring
  And show springs is <show_springs>
  When the coder renders the world
  Then spring lines should be <spring_visibility>
  And masses should remain visible

Examples:
  | show_springs | spring_visibility |
  | true         | visible           |
  | false        | hidden            |

Scenario: fixed masses are visually distinguishable
  Given the world contains a fixed mass and a movable mass
  When the coder renders the world
  Then the fixed mass should be visually distinguishable from the movable mass
