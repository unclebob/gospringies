# mutation-stamp: sha256=82d859d889b6fd839989c8214461ea03b4c135fa4f5b8498040f7b4026317196
Feature: Pipeline smoke

Scenario: smoke
  Given acceptance smoke is ready
  Then acceptance smoke should pass
