Feature: Packaging and docs

Background:
  Given the packaging and docs task is accepted

Scenario Outline: documented commands are available
  When a developer reads the project documentation
  Then command <command> should be documented

Examples:
  | command          |
  | unit tests       |
  | acceptance tests |
  | mutation tests   |
  | build            |
  | run              |

Scenario Outline: documented commands work from a clean checkout
  Given a clean checkout
  When a developer runs documented command <command>
  Then command <command> should pass

Examples:
  | command          |
  | unit tests       |
  | acceptance tests |
  | build            |
  | run              |

Scenario Outline: documentation covers desktop prerequisites and user workflows
  When a developer reads the project documentation
  Then the documentation should explain <topic>

Examples:
  | topic                         |
  | Ebitengine desktop prerequisites |
  | creating a simulation         |
  | loading a simulation          |
  | saving a simulation           |
  | running a simulation          |

Scenario: verification results are included in the handoff
  When the coder completes the packaging and docs task
  Then the handoff should include the local verification commands that were run
  And the handoff should include the result of each verification command
