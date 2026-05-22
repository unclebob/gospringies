# mutation-stamp: sha256=b9600f557cd0b68b10725f37c01ad583801e37338a825936834d6e142d0d9eb0
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:43:20-05:00","feature_name":"Pipeline smoke","feature_path":"features/pipeline_smoke.feature","background_hash":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"smoke","scenario_hash":"f5d7bc7679c25337e09ccd5f57564458196d0c7539095ac3e7a5dd3b79160653","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:43:20-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Pipeline smoke

Scenario: smoke
  Given acceptance smoke is ready
  Then acceptance smoke should pass
