# mutation-stamp: sha256=9582a93c94b27a5bb1783d944c68691f1c842f3590c3f38cbeb08787527242dc
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
