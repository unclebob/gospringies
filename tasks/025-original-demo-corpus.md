# Task 025: Original Demo Corpus

## Goal

Import the original XSpringies demo `.xsp` corpus so the application has the same broad example set as packaged XSpringies.

## Scope

- Locate a preserved `xspringies-1.12` source/package source that includes the original demo files.
- Import the original `.xsp` demo files into `demos/original/`.
- Preserve original demo file names and text content except for unavoidable newline normalization.
- Keep the existing starter demos in `demos/`.
- Add a provenance note documenting the source URL, retrieval date, and license/package context.
- Ensure every imported `.xsp` file loads with the project XSP parser.
- Ensure the imported corpus includes the known packaged demo names listed by FreshPorts.

## Acceptance Notes

- Do not hand-transcribe the corpus.
- The imported files should be usable by Load and Insert commands.
- If an upstream demo uses a documented XSP command not yet supported, report the unsupported command and leave the failing file out until Task 020 support lands.

## Done When

- `demos/original/` contains the imported original demo corpus.
- A provenance document is committed with the imported files.
- Acceptance or unit coverage verifies every imported `.xsp` loads.
