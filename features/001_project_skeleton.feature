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
