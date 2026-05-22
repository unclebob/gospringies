# mutation-stamp: sha256=badd8a0b0bc67db21333e09a05216988dac48bddc6b35e01bf450a164c270d8d
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:08:42-05:00","feature_name":"Packaging and docs","feature_path":"features/014_packaging_and_docs.feature","background_hash":"a27e0809572e55debafb449f048b43b2b7016926b6cde1c506da0ce0cbda3053","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"documented commands are available","scenario_hash":"291c3d023510c5d594ab6eb6b736f10bbd4e9d0069d2619a6fec53080342abd1","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:42-05:00"},{"index":1,"name":"documented commands work from a clean checkout","scenario_hash":"83f5c848c09d0d75a0458c8df5aa7a2330b93319ee3f2a166201e8dd7719a789","mutation_count":4,"result":{"Total":4,"Killed":4,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:42-05:00"},{"index":2,"name":"documentation covers desktop prerequisites and user workflows","scenario_hash":"be403066ab2a04988cb0212c7967721437dc6f69d89655e18178ab0c39b2e2da","mutation_count":5,"result":{"Total":5,"Killed":5,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:42-05:00"},{"index":3,"name":"verification results are included in the handoff","scenario_hash":"00e053fbdecba3fa4604aa2852e21b001d46c03b87467bc4f8e94be3fa9c9489","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:08:42-05:00"}]}
# acceptance-mutation-manifest-end
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
  | topic                            |
  | Ebitengine desktop prerequisites |
  | creating a simulation            |
  | loading a simulation             |
  | saving a simulation              |
  | running a simulation             |

Scenario: verification results are included in the handoff
  When the coder completes the packaging and docs task
  Then the handoff should include the local verification commands that were run
  And the handoff should include the result of each verification command
