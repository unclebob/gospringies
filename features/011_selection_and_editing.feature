# mutation-stamp: sha256=698e91d4e64f7951f01483a20f5051b6dae4895db7cf883a01aad897904728ad
# acceptance-mutation-manifest-begin
# {"version":1,"tested_at":"2026-05-22T12:06:17-05:00","feature_name":"Selection and editing","feature_path":"features/011_selection_and_editing.feature","background_hash":"b9bc7f5f302b99cb82c483b16bb12f5816128dbf469c8c6df7828b2af77c0b98","implementation_hash":"0a770bae08f130ca996aec47a8d033925e33cb9481c31df3fd9eeaca32e1424c","scenarios":[{"index":0,"name":"objects can be selected individually","scenario_hash":"7496d08a7c5c694f22238c6261574a65c6dac8bec81daf4c2b5ac91dcdccf85d","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:06:17-05:00"},{"index":1,"name":"select all selects every object","scenario_hash":"ef215948df35edc02afb6f5410910aefaf26a0498eae6eac7881a55913655da8","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:06:17-05:00"},{"index":2,"name":"deleting a selected object removes it","scenario_hash":"0c6f803fd2d5c21bb90fad18b2527f5dbd9175457fef0dc92aabe13f8172a3b0","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:06:17-05:00"},{"index":3,"name":"deleting a mass deletes attached springs","scenario_hash":"8df43e61f47211af1384378cc28f18286dfa63427619a0bbfc9cf36aa07d60f7","mutation_count":0,"result":{"Total":0,"Killed":0,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:06:17-05:00"},{"index":4,"name":"duplicating selected objects creates independent objects","scenario_hash":"350253be430d113bf12b85dff11c4f8b13f3f5ffacc8a49f36e8ed45834cd34d","mutation_count":2,"result":{"Total":2,"Killed":2,"Survived":0,"Errors":0},"tested_at":"2026-05-22T12:06:17-05:00"}]}
# acceptance-mutation-manifest-end
Feature: Selection and editing

Background:
  Given the selection and editing task is accepted

Scenario Outline: objects can be selected individually
  Given the world contains a <object_type> with id <id>
  When the coder selects <object_type> <id>
  Then <object_type> <id> should be selected

Examples:
  | object_type | id |
  | mass        | 1  |
  | spring      | 2  |

Scenario: select all selects every object
  Given the world contains masses and springs
  When the coder selects all objects
  Then every mass should be selected
  And every spring should be selected

Scenario Outline: deleting a selected object removes it
  Given the world contains a <object_type> with id <id>
  And <object_type> <id> is selected
  When the coder deletes selected objects
  Then <object_type> <id> should not exist

Examples:
  | object_type | id |
  | mass        | 1  |
  | spring      | 2  |

Scenario: deleting a mass deletes attached springs
  Given mass 1 is connected to mass 2 by spring 3
  And mass 1 is selected
  When the coder deletes selected objects
  Then mass 1 should not exist
  And spring 3 should not exist
  And mass 2 should still exist

Scenario Outline: duplicating selected objects creates independent objects
  Given selected <object_set> exists
  When the coder duplicates selected objects
  Then duplicated objects should have unique ids
  And duplicated objects should be independent from the original objects

Examples:
  | object_set             |
  | one mass               |
  | two masses and a spring |
