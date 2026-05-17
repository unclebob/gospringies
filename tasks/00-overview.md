# XSpringies Implementation Task Slices

These tasks break the application into small, reviewable chunks for a Go implementation using Ebitengine for desktop windowing, input, rendering, and the real-time loop.

Keep the physics and file-format behavior in plain Go packages that do not depend on Ebitengine. Use Ebitengine only at the app boundary. Each implementation slice should include focused unit tests for domain behavior before production code, and acceptance coverage where externally visible behavior changes.

Before a task is forwarded to coder, the task must have a concise Gherkin acceptance specification committed under `features/`.

Suggested order:

1. `01-project-skeleton.md`
2. `02-acceptance-pipeline.md`
3. `03-domain-model.md`
4. `04-system-parameters.md`
5. `05-force-evaluation.md`
6. `06-simulation-step.md`
7. `07-xsp-load-save.md`
8. `08-ebitengine-window.md`
9. `08a-screen-and-controls.md`
10. `09-render-world.md`
11. `10-mouse-editing.md`
12. `11-selection-and-editing.md`
13. `12-controls-and-hotkeys.md`
14. `13-demo-files.md`
15. `14-packaging-and-docs.md`
16. `015-edit-mode-details.md`
17. `016-spring-mode-mouse-semantics.md`
18. `017-state-save-restore.md`
19. `018-selected-object-parameter-editing.md`
20. `019-wall-collision-and-stickiness.md`
21. `020-xsp-complete-file-format.md`
22. `021-force-center-and-force-parameters.md`
23. `022-adaptive-rk4-numerics.md`
24. `023-nonblank-startup-editor.md`
