Feature: XSP complete file format

Background:
  Given the XSP complete file format task is accepted

Scenario Outline: documented commands load and save
  Given XSP input contains command <command>
  When the coder loads and saves the XSP input
  Then saved XSP output should include command <command>

Examples:
  | command |
  | cmas    |
  | elas    |
  | kspr    |
  | kdmp    |
  | fixm    |
  | shws    |
  | cent    |
  | frce    |
  | visc    |
  | stck    |
  | step    |
  | prec    |
  | adpt    |
  | gsnp    |
  | wall    |
  | mass    |
  | spng    |

Scenario Outline: file format validation rejects documented invalid structure
  Given XSP input has problem <problem>
  When the coder loads the XSP input
  Then loading should fail with reason <reason>

Examples:
  | problem              | reason                |
  | blank line           | blank lines not allowed |
  | missing final newline | missing final newline |
  | non-positive id      | ids must be positive  |

Scenario Outline: file operations normalize file paths
  Given filename input is <filename>
  And environment variable SPRINGDIR is <springdir>
  When the coder resolves an XSP filename
  Then resolved filename should be <resolved_filename>

Examples:
  | filename | springdir | resolved_filename        |
  | demo     | unset     | demo.xsp                 |
  | demo.xsp | unset     | demo.xsp                 |
  | demo     | examples  | examples/demo.xsp        |

Scenario Outline: insert loads only objects
  Given current parameters are <current_parameters>
  When the coder inserts XSP file <input_file>
  Then inserted masses and springs should be added
  And parameters should remain <current_parameters>

Examples:
  | current_parameters | input_file |
  | custom             | complete   |
