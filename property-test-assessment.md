# Property Test Assessment

This assessment covers production Go code in `cmd/` and `internal/`. Test files, generated acceptance files, and generated headless output are excluded. Every production function was assessed either directly or as part of a narrow function family when the functions share the same behavior class.

Property tests should be kept behind `//go:build property` and run explicitly, for example:

```sh
GOCACHE=$PWD/.gocache go test -tags 'appunit property' ./internal/sim
```

They should not be part of normal coverage, CRAP, mutate4go, or Gherkin mutation runs unless a task explicitly requests that.

## Legend

- `property`: useful property-based coverage should be added.
- `example`: ordinary example/table/unit tests are the right tool.
- `mutation`: existing unit/acceptance tests plus mutation testing are sufficient.
- `none`: no useful property exists, or the function is a trivial wrapper/glue path.

## Current Property Test

- `internal/sim/wall_spring_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyMassCrossingWallSpringStaysOnStartingSide`
  - `TestPropertyVec2AlgebraIdentities`
  - `TestPropertyBoundsCenterMatchesConfiguredExtents`
  - `TestPropertyCloneAndLoadFromDoNotAliasSlicesOrMaps`
  - `TestPropertyAddMassAndSpringLookupConsistency`
  - `TestPropertySpringForcesAreEqualAndOpposite`
  - `TestPropertyFixedMassesHaveZeroAcceleration`
  - `TestPropertyScreenWallCollisionReturnsMassInside`
  - `TestPropertyWallSpringLengthConstraintRestoresRestLength`
  - `TestPropertyClosestPointAndFractionStayOnSegment`
  - `TestPropertyOffCanvasCleanupKeepsValidSpringEndpoints`
  - `TestPropertyParametersRoundTripWithoutAliasing`
  - `TestPropertyAdvanceDurationUsesPositiveBoundedSteps`
  - `TestPropertyStepWithNoForcesMovesLinearlyAndKeepsFixedMasses`
  - `TestPropertyGravityForceScalesWithMass`
  - `TestPropertyViscosityForceOpposesVelocity`
  - `TestPropertyCenterOfMassTranslatesWithMasses`
  - `TestPropertyForceCenterTracksSingleSelectedMass`
  - `TestPropertyMassRadiusIsBoundedAndMonotonic`
  - `TestPropertyWallSpringContactFractionFindsSegmentCrossing`
  - `TestPropertyResolveWallSpringVelocitySeparatesOrKeepsSeparating`
  - `TestPropertyResetClearsWorldAndInsertFromAppendsObjects`
  - `TestPropertyAddMassAtAndAddSpringBetweenGenerateConsistentIDs`
  - `TestPropertyAdaptiveStepHelpersStayPositiveAndBounded`
  - `TestPropertyCenterAndWallForcesPointTowardTheirTargets`
  - `TestPropertyEnabledForceMatchesParameterState`
  - `TestPropertyStuckMassStaysOnWallUntilReleaseForceWins`
  - `TestPropertyStepDurationFollowsConfiguredTimestep`
  - `TestPropertyMassCollisionConservesMovableMomentum`
  - `TestPropertyFiniteStepOutputsRemainFinite`
  - `TestPropertyForceEvaluationSkipsInvalidSpringsAndScalesAcceleration`
  - `TestPropertyWallSpringLengthConstraintCollisionKeepsEndpointOnBarrierSide`
  - `TestPropertyMovingWallSpringFixedEndpointCollisionSeparatesContact`
  - `TestPropertyWallSpringCollisionConservesMomentumWithVaryingMasses`
- `internal/edit/editor_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyEditorSnapIDsAndDistance`
  - `TestPropertyBoxSelectIsInvariantUnderBoxCornerOrder`
  - `TestPropertyBoxSelectChoosesSinglePartialSpringOnlyWhenNothingElseIsEnclosed`
  - `TestPropertyMoveAndThrowSelectedAffectOnlySelectedMovableMasses`
  - `TestPropertyDeleteSelectedRemovesAttachedSpringsAndReindexes`
  - `TestPropertyDuplicateSelectedCreatesUniqueValidIDs`
  - `TestPropertySelectionGeometryIsOrderAndTranslationInvariant`
- `internal/format/format_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyFromSimulationAndXSPRoundTripPreserveNormalizedWorld`
- `internal/gherkin/gherkin_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyParseJSONTableAndStepRoundTrips`
- `internal/mutationstamp/stamp_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertySplitHashAndStampFormattingRoundTrip`
- `internal/acceptancemutation/mutator_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyMutationBuildersAndSummariesAreDeterministic`
  - `TestPropertyScenarioManifestRoundTripsAndSkipPlansAreStable`
- `internal/acceptancegen/generator_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyGeneratedTestNameIsDeterministicAndValid`
- `internal/appcore/startup_scene_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyStartupSceneHelpersAreDeterministicAndApplyBounds`
- `internal/app/app_properties_test.go` is behind the `property` build tag. Completed property tests now cover:
  - `TestPropertyCanvasBoundsClampAndSnapStayConsistent`
  - `TestPropertyCanvasCoordinatesRoundTrip`
  - `TestPropertyMassHitTestingMatchesDrawRadius`
  - `TestPropertyDialogRectsStayInsideScreen`
  - `TestPropertyVisibleControlLookupsAndNumericHelpersAreStable`
  - `TestPropertyUpdateByIDOnlyUpdatesMatches`
  - `TestPropertyRenderSpringEndpointLookupMatchesValidEndpoints`
  - `TestPropertyValueDialogSliderAndRectsAreBounded`
  - `TestPropertyDemoPickerGeometryAndClampAreBounded`
  - `TestPropertyNumericSettingControlsRoundTrip`
  - `TestPropertyClipboardPasteKeepsIDsAndReferencesValid`
  - `TestPropertyRenderGeometryHelpersAreDeterministic`
  - `TestPropertyVisibleControlReportsAreInternallyConsistent`

## Highest-Value Additions

1. `internal/sim`: swept collisions, wall springs, conservation laws, fixed masses, and cleanup thresholds.
2. `internal/edit/selection.go`: rectangle/segment geometry and selection invariants.
3. `internal/format`: XSP load/save round trips.
4. `internal/gherkin` and `internal/acceptancemutation`: parser and mutation-manifest round trips/determinism.
5. `internal/app`: pure geometry, coordinate transforms, numeric setting bounds, and layout helpers only.

## internal/sim

Physics code has the strongest case for property testing because the valid input space is large and failures often appear only under unusual geometry, mass, velocity, or timestep combinations.

### `simulation.go`

- `Vec2.Add`, `Sub`, `Scale`, `Normalize`, `dot`: `property` - **done**. Algebraic identities, commutativity for `Add`, inverse relation of `Add`/`Sub`, distributivity of `Scale`, dot-product symmetry, and normalized vector length are covered.
- `Bounds.MinX`, `MaxX`, `MinY`, `MaxY`, `Center`, `configuredBoundary`: `property` - **done**. Center must lie halfway between min/max; min/max must use configured values when present and defaults otherwise.
- `NewWorld`, `NewDemoSimulation`: `example`. Constructor defaults are finite and expected; broad randomization is not useful.
- `Simulation.Clone`, `LoadFrom`: `property` - **done**. Clone/load must preserve values while avoiding aliasing.
- `Simulation.Reset`, `InsertFrom`: `property` - **done**. Reset and insert should preserve/reset source objects and produce usable IDs.
- `SetTemperatureSeed`, `temperatureRandom`: `example`. Determinism can be table tested; randomized property tests add little.
- `AddMass`, `AddSpring`: `property` - **done**. Successful adds preserve lookup consistency; duplicate IDs and invalid spring endpoints are rejected.
- `AddMassAt`, `AddSpringBetween`: `property` - **done**. Generated IDs are unique and above existing IDs.
- `MassByID`, `SpringByID`, `byID`, `massIndexByID`, `validSpringMassIndexes`: `property` - **done**. Lookup result must match exactly one item with the ID, and miss for absent IDs.
- `AdvanceDuration`, `Step`, `configuredTimeStep`, `positiveAdvanceStep`: `property` - **done** for positive bounded advance duration and zero-force linear stepping.
- `Advance`, `advanceStepDuration`, `configuredPrecision`, `positiveParameterOrDefault`, `adaptiveStepDuration`: `property` - **done**. Positive durations should advance in bounded positive chunks, never loop forever, and never exceed requested remaining duration.
- `stepRK4`, `derivatives`, `offsetMasses`, `weightedDerivative`: `property` - **done**. Zero-force/free-linear, fixed-mass behavior, and finite-output edge cases are covered.
- `massPositions`, `activeMasses`: `example`. Simple slice projection/filtering.
- `length`, `sqrt`: `none`. Thin math wrappers.

### `forces.go`

- `EvaluateForces`, `addSpringForces`, `springEndpointMasses`, `springEndpointMassesByID`, `springEndpointMassesByIndex`, `validSpringMassIndexes`: `property` - **done**. Equal-and-opposite spring forces, endpoint lookup consistency, invalid endpoint skips, and wall spring force skips are covered.
- `addEnvironmentalForces`, `computeAccelerations`: `property` - **done**. Fixed masses have zero acceleration; movable positive-mass acceleration equals force divided by mass.
- `gravityForce`, `viscosityForce`: `property` - **done**. Gravity scales with mass; viscosity opposes velocity.
- `centerForce`, `wallForce`, `wallChecks`: `property` - **done**. Center force points toward the configured center; enabled wall force points inward.
- `centerOfMass`, `forceCenter`, `screenCenter`, `SetForceCenter`, `CenterMassID`, `IsCenterMass`: `property` - **done** for center translation and selected force-center tracking.
- `enabledForce`: `property` - **done**. Enabled force lookup should match parameter force state.

### `walls.go`

- `applyWallCollision`, `wallCollisionActive`, `bounceOrStick`, `signedRebound`, `collisionWalls`: `property` - **done** for outside moving masses returning to the boundary side and no longer moving outward.
- `keepStuck`, `wallReleasedBy`, `stuckWall`: `property` - **done**. Stuck masses should not jitter through walls, and release force should clear stuck state only when appropriate.

### `collisions.go`

- `applyMassCollisions`, `firstCollisionPartnerIndex`, `applyMassCollision`, `collisionGeometryFor`, `collisionVelocitiesSeparating`, `axisVelocitiesSeparating`, `avoidVerticalDivision`, `applyCollisionVelocity`, `collisionRatio`, `effectiveCollisionMass`, `MassRadius`: `property` - **done**. Mass collision resolution should keep fixed masses immobile, separate or leave separating velocities, respect effective masses, and conserve linear momentum for movable masses within tolerance.
- `applyWallSpringLengthConstraints`, `applyWallSpringLengthConstraint`, `applyWallSpringLengthCorrection`, `moveSingleFixedWallSpringEndpoint`, `shareWallSpringLengthCorrection`, `moveWallSpringEndpoint`: `property` - **done**. Wall-spring length constraint error reduction is covered; the post-correction collision properties cover barrier-side preservation after constraint movement.
- `applyWallSpringLengthConstraintCollisions`, `applyWallSpringEndpointConstraintCollisions`: `property` - **done**. Endpoint movement caused by stiff-wall length correction is collision-checked against other wall springs and leaves endpoints on the valid side of barriers.
- `applyWallSpringCollisions`, `wallSpringBoundaryStartPenetrating`, `wallSpringPreviousPosition`, `wallSpringEndpointIndexes`, `shouldApplyWallSpringCollision`, `springEndpointIndexes`, `applyWallSpringCollision`: `property` - **done**. Moving mass no-crossing, boundary-start contact, previous-position lookup, endpoint lookup, endpoint tunneling, and variable-mass momentum transfer are covered.
- `applyMovingWallSpringFixedEndpointCollisions`, `movingWallSpringEndpointIndexes`, `applyMovingWallSpringAgainstFixedEndpoints`, `applyMovingWallSpringFixedEndpointCollision`, `skipMovingWallSpringFixedEndpointCollision`, `fixedEndpointContactOutside`, `fixedEndpointContactResolved`, `movingWallSpringFixedEndpointContact`, `previousFixedEndpointNormal`, `currentFixedEndpointContact`, `resolvedFixedEndpointContactVelocity`, `shareMovingWallSpringContactImpulse`, `contactShareInverseMass`: `property` - **done**. Moving wall springs should not pass through fixed endpoint masses; fixed endpoints stay fixed; impulse sharing depends on contact fraction and endpoint masses while resolving the contact velocity.
- `closestFractionOnSegment`, `wallSpringContactFraction`, `wallSpringCrossingRejected`, `wallSpringIntersectionFraction`, `sameSign`, `sideSign`, `collisionStartSide`, `closestPointOnSegment`: `property` - **done** for closest/contact points staying on the segment, parametric fractions staying in `[0,1]`, and reversed endpoint equivalence.
- `wallSpringContactVelocity`, `resolveWallSpringVelocity`, `wallSpringVelocitySeparating`, `shareWallSpringImpulse`: `property` - **done**. Resolved velocity separation and momentum transfer to movable wall endpoints, including unequal endpoint masses, are covered.
- `applyWallSpringTemperatureKick`, `fullScreenGravityKick`: `example`. Temperature kick uses random direction and bounded magnitude; deterministic seed examples are clearer than broad properties.
- Degenerate/zero-length guards: `example`. These are boundary table tests rather than broad properties.

### `off_canvas_cleanup.go`

- `cleanupOffCanvasObjects`, `assignSpringEndpointIDs`, `validMassIndex`, `massBeyondCleanupBoundary`, `removeSpringsAttachedTo`, `reindexSprings`: `property` - **done**. Any mass beyond the cleanup boundary must be removed; springs attached to removed masses must be removed; surviving spring indexes/IDs must point to valid masses; objects inside the boundary are preserved.

### `parameters.go`

- `Parameters.Clone`, `Has`, `Value`, `Set`, `EnableForce`, `SelectForce`, `EnableWall`, `Force`, `WallEnabled`: `property` - **done**. Clone must not alias; set/get round trips; enabling/selecting preserves configured values where expected.
- `StepDuration`: `property` - **done**. Step duration parsing should follow the configured timestep value.

## internal/edit

Editor code splits into geometry helpers, which are good property targets, and command/state methods, which are better covered by focused examples.

### `editor.go`

- `NewEditor`: `example`.
- `Click`, `CreateSpring`, `DragMass`: `example`. User operation sequencing is clearer with table tests.
- `snap`: `property` - **done**. Snapped coordinates are grid multiples when snap is enabled and unchanged when disabled.
- `nextMassID`, `nextSpringID`, `nextID`: `property` - **done**. Returned ID is greater than every existing ID and is `1` for empty input.
- `distance`: `property` - **done**. Non-negative, symmetric, zero only for equal points.
- `parameterFloat`, `parameterBool`: `example`. Default parsing behavior is small and explicit.

### `selection.go`

- `SelectMass`, `AddMassSelection`, `SelectSpring`, `SelectNearest`, `selectExisting`, `SelectAll`, `ClearSelection`, `MassSelected`, `SpringSelected`: `example`. State transitions need precise scenarios, not broad generated inputs.
- `BoxSelect`, `selectFullyEnclosedSprings`, `selectSinglePartiallyEnclosedSpring`, `singlePartiallyEnclosedSpringID`: `property` - **done**. Selection should be invariant under box corner ordering and should select exactly objects inside/intersecting the box under the documented rules.
- `MoveSelected`, `ThrowSelected`: `property` - **done**. Only selected movable masses should change; deltas/velocities should be applied uniformly.
- `DeleteSelected`, `DuplicateSelected`, `deleteSelectedMasses`, `deleteSelectedSprings`, `duplicateMasses`, `duplicateSprings`, `reindexSprings`, `replacementID`, `keepSpring`: `property` - **done**. After deletion/duplication, all springs reference existing masses and duplicated IDs are unique.
- `toggleMassSelection`, `clearSelection`, `massExists`, `springExists`, `objectExists`, `worldIndexByMassID`, `nearestMassID`: `example`.
- `withinBox`, `segmentIntersectsBox`, `segmentFullyWithinBox`, `segmentsIntersect`, `oppositeSides`, `hasCollinearEndpoint`, `collinearEndpointOnSegment`, `orientation`, `onSegment`, `between`, `ordered`: `property` - **done**. These are pure geometry functions; test endpoint order symmetry, box order symmetry, collinear edge cases, and translation invariance.

### `spring_pointer.go`

- `BeginSpring`, `DragSpring`, `ReleaseSpring`, `PendingSpring`, `massNear`: `example`. Gesture lifecycle is discrete and stateful.

### `parameter_editing.go`

- `ChangeControl`, `SetRestLength`, `changeMassFloat`, `changeMassFixed`, `changeSpringFloat`, `changeSpringWall`, `changeFloat`, `currentSpringLength`, `hasSelectedMass`, `hasSelectedSpring`: `example`. These are parsing and selected-object command paths; mutation-tested examples are appropriate.

## internal/format

- `FromSimulation`: `property` - **done**. Document conversion should preserve all masses, springs, parameters, bounds, force center, and index/ID endpoint intent.
- XSP load/save functions in `xsp.go`: `property` - **done**. `LoadXSP(SaveXSP(world))` should preserve a normalized simulation, and `SaveXSP(LoadXSP(text))` should be stable after normalization.
- Small parsing/formatting helpers in `xsp.go`: `example`. Invalid input and compatibility cases should stay table-driven.

## internal/gherkin

- `Parse`, `ReadFile`, `WriteJSON`, `ReadJSON`: `property` - **done**. Parse/write/read should preserve a feature through JSON round trip.
- `parseTableRow`: `property` - **done**. Cell count and trimming should be stable, including escaped or empty cells if supported.
- `parseStep`, `isStep`: `property` - **done**. Every generated recognized step prefix should parse to the same keyword/body split.
- `lineParser.parseLine`, `ignoreBlankOrComment`, `parseFeatureLine`, `parseBackgroundLine`, `parseScenarioOutlineLine`, `parseScenarioLine`, `parseScenarioWithPrefix`, `parseExamplesLine`, `parseExampleRowLine`, `parseStepLine`, `startFeature`, `startBackground`, `startScenario`, `startExamples`, `addExampleRow`, `exampleRow`, `addStep`: `example`. Parser state transitions and error wording are clearer as examples.

## internal/mutationstamp

- `Split`, `Hash`, `formatStamp`, `stampHash`: `property` - **done**. Splitting stamped/unstamped content must be stable; `stampHash(formatStamp(hash)) == hash`; hash changes when implementation content changes.
- `Valid`, `Stamp`, `Remove`: `example`. File I/O behavior should be tested with temp files and fixed fixtures.

## internal/acceptancemutation

### `mutator.go`

- `BuildMutations`, `buildMutation`, `mutateValue`, `mutateList`, `mutateKeyword`, `mutateNumber`, `signedIntDelta`, `signedFloatDelta`, `mutateDate`, `mutateDuration`, `deterministicRand`, `dither`: `property` - **done**. Mutation generation should be deterministic, should not return the original value when a mutation is possible, and should preserve parseable shape for numbers/dates/durations.
- Equivalent-mutation filters (`isEquivalent...`, `scenarioOnlyEquivalentCheck`, `controlsParameterSetupKey`, `mutationKeyIn`): `example`. These are policy tables and should remain explicit.
- `filterMutations`, `mutationWorkerCount`, `Summarize`, `cloneFeature`: `property` - **done**. Filtering preserves order; worker count is bounded; summaries equal counts by status; clone must not alias.
- Worker/process functions (`RunMutations`, `RunMutationsWithOptions`, `runMutationJobs`, `withMutationContext`, `startMutationWorkers`, `collectMutationResults`, `runMutationWorker`, `nextMutationJob`, `enqueueMutationJobs`, `closeCompletedWhenDone`, `mutationProgressTracker.record`, `add`, `shouldReport`, `runMutation`, `mutationCommandContext`, `mutationPaths`, `writeMutationTest`, `mutationStatus`, `mutationCommandTimedOut`): `example`. Concurrency and subprocess behavior should be fixture tested with deterministic commands.

### `scenario_manifest.go`

- `ParseScenarioManifest`, `scenarioManifestBlock`, `RemoveScenarioManifest`, `ScenarioSkipPlanFor`, `scenarioManifestEntryFor`, `mutationIDsForScenario`, `BuildScenarioManifest`, `MergeScenarioManifest`, and related hash/key helpers: `property` - **done**. Manifest parse/remove/merge should be idempotent; unchanged scenario keys under soft mode should skip; changed keys should rerun; mutation IDs should be deterministic and unique.
- File I/O wrappers: `example`.

## internal/acceptancegen

- `generatedTestName`: `property` - **done**. Generated names should be deterministic, valid Go identifiers, and collision-resistant for common path variants.
- `GenerateGoTest`, `GenerateTaggedGoTest`, `generateGoTest`, `generatedTestSource`, `writeFormattedGo`, `readFeatureForGeneration`: `example`. Template output and file errors are fixture based.

## internal/appcore

- `DefaultStartupScenePath`, `NewDefaultStartupWorld`, `LoadDefaultStartupWorld`, `DefaultStartupSceneCandidates`, `ApplyBounds`: `property` - **done**. Startup scene helpers should be deterministic and apply configured bounds.

## internal/app

Most app functions are UI state orchestration. Property tests are appropriate only for pure geometry, numeric bounds, coordinate conversion, ID allocation, and lookup invariants. Rendering, Ebiten polling, dialog keyboard flow, and user-command dispatch should remain example/appunit tested.

### `app.go`

- `Run`, `Update`, `pollDemoPickerScroll`, `pollMouseControls`, `pollKeyboardControls`, `pollEscapeShortcut`, `pollControlShortcuts`, `handlePressedShortcut`, `pressedControlShortcut`, `pressedControlShortcutFrom`, `firstPressedShortcut`, `controlKeyPressed`, `shiftKeyPressed`, `controlKeyPressed`, `throwKeyPressed`, `anyKeyPressed`: `example`. Ebiten polling and shortcut dispatch should stay appunit/headless fixture driven.
- `Draw`, `drawOpenOverlays`, `drawEditorChrome`, `drawGridPoints`, `drawSprings`, `drawPendingSpring`, `drawSpringLine`, `drawSelectionDrag`, `drawMasses`, `drawWalls`, `drawSelection`, `drawSelectionLine`: `example`. Pixel/rendering behavior belongs in headless or appunit render reports.
- `editorChromeAntiAlias`, `editorChromeRects`, `gridPointPixelSize`, `gridPointAntiAlias`, `springLineAntiAlias`, `massDrawAntiAlias`: `none`. Constants/trivial draw configuration.
- `gridPointRects`, `gridPoints`, `validGridSnapSize`, `firstGridCoordinateAtOrAfter`: `property` - **done**. Grid points should be bounded by the canvas, monotonic by grid size, and use the first coordinate at or after the visible minimum.
- `springDrawColor`, `massDrawColor`, `drawColorFor`: `example`. Small selection/color policy tables.
- `pendingSpringLine`, `selectionRectangleLines`, `wallDrawLines`, `selectedMasses`, `explicitSelectedMasses`, `allMassesImplicitlySelected`, `selectedSpringLines`, `selectedMassOutline`, `selectionOutline`: `property` - **done**. Generated lines/outlines should be deterministic, bounded around selected objects, and stable under endpoint or selection ordering where applicable.

### `app_appunit.go`

- `Run`, `Update`, `shiftKeyPressed`, `controlKeyPressed`, `throwKeyPressed`: `none`. Appunit build stubs.

### `app_shared.go`

- `NewGame`, `DefaultWindowConfig`, `advanceSimulationFrame`, `markDirty`, `clearDirty`, `setSelected`, `reattachEditor`, `loadWorldIntoSession`, `scrollDemoPicker`, `demoList`, `editing`, `Layout`, `World`, `SetPaused`, `Paused`, `InputActive`, `RenderingActive`, `RenderFrame`, `Close`, `Closed`: `example`. Constructor/session state and accessors are fixed appunit scenarios, except scroll clamping is covered through `demo_picker.go` candidates.

### `canvas_bounds.go`

- `canvasWorldBounds`, `canvasWorldBoundsForHeight`, `applyCanvasWallBounds`, `clampToCanvas`, `positionInCanvas`, `snapToCanvas`: `property` - **done**. Clamp results must be inside bounds, snap must preserve grid multiples, and positions already inside remain unchanged.

### `canvas_coordinates.go`

- `selectNearest`: `example`. Selection policy is a user interaction scenario.
- `massAt`, `massDrawCircle`, `screenToWorld`, `worldToScreen`, `canvasCoordinate`, `flipCanvasY`: `property` - **done**. Screen/world round trips should be within tolerance; y-up/y-down modes should invert consistently; hit testing should agree with mass radius.

### `clickable_controls.go`

- `ClickAt`, `clickVisibleControl`, `clickAwayFromVisibleControls`, `ClickVisibleControl`, `activateVisibleControl`, `editMenuControlAt`, `activateInspectorControl`, `setSliderAt`, `continueNumericStepHold`, `stepEditorControl`, `toggleFixedMass`, `toggleSelectedSpringWall`, `toggleForce`, `stepForceValue`, `setForceValue`, `toggleWall`, `toggleGridSnap`, `toggleParameter`, `stepParameter`, `VisibleControlActive`, `DragMass`, `dragSelectedMasses`, `dragSingleMass`, `finishMassDragStep`, `moveSelectedMasses`, `applyDraggingOffsets`: `example`. These are command paths and state transitions best covered by appunit scenarios.
- `VisibleControlBounds`, `sliderFractionAt`, `parameterForEditorControl`, `forceConfig`, `nonNilStringMap`, `forceValueFloat`, `selectedMassIDs`, `parameterFloat`, `gridSnapSize`, `formatControlFloat`, `roundControlFloat`, `clampFloat`, `visibleControlAt`, `controlAt`, `visibleControlWithLabel`, `visibleControlWithName`, `visibleControlWithField`, package-level `visibleControlAt`, `visibleControlWithLabel`, `visibleControlWithName`, `visibleControlWithField`, `snapToGrid`: `property` - **done**. Lookup by label/name should return visible controls only; slider fractions and clamps should stay bounded; snap results should be grid-aligned.
- `PathEntryCommand`, `DemoPickerOpen`: `none`. Simple accessors.

### `commands.go`

- `RunCommand`, `resetWorld`, `restoreWorldState`, `selectAllObjects`, `deleteSelection`, `cutSelection`, `pasteAtCursor`, `duplicateSelection`, `clearSelection`, `syncSelectionState`, `openDemoPicker`, `SaveXSP`, `LoadXSP`, `LoadXSPFromFile`, `InsertXSP`, `SaveState`, `RestoreState`, `SetParameter`, `ReplaceWorld`, `setAppBounds`, `appBounds`: `example`. These are command/session scenarios already aligned with appunit and acceptance tests.
- `pathEntryLabel`: `example`. Small command-label policy table.

### `context_menu.go` and Draw Files

- `clickContextMenu`, `contextMenuLabels`, `selectContextMenuItem`, `contextMenuRect`, `contextMenuRowRect`: `example`. Menu row selection and close behavior are fixed UI scenarios.
- `drawContextMenu`, `drawMassContextMenu`, `drawSpringContextMenu`: `example`. Rendering belongs in appunit/headless checks.

### `demo_picker.go` and `demo_picker_draw.go`

- `demoPickerRect`, `visibleDemoPaths`, `LoadPickerEntries`, `loadPickerEntryMatches`, `demoPickerVisibleRows`, `demoRowRect`, `demoPathAt`, `buildDemoList`, `globXSP`, `groupedLoadPickerEntries`, `clampInt`: `property` - **done**. Visible rows should be bounded by picker size, scrolling should not produce invalid indices, grouping should preserve section order, and clamp should stay within bounds.
- `demoPickerTitlePoint`, `demoPickerRowTextPoint`, `demoPickerRowFill`: `property` - **done**. Title/text points should stay inside their source rectangles and row fill should alternate deterministically.
- `ChooseLoadPickerEntry`, `clickDemoPicker`, `loadDemoAt`, `loadDemoPath`, `recordDemoLoadResult`, `LastFileError`: `example`. File loading and UI click behavior are fixture scenarios.
- `drawDemoPicker`, `demoPickerPanelAntiAlias`, `demoPickerRowAntiAlias`: `example` or `none`. Rendering and constants are not property targets.

### Dialogs and Keyboard

- `centeredDialogRect`, `dialogTextRect`, `dialogOKRect`, `saveFilenameDialogRect`, `saveFilenameTextRect`, `saveFilenameDialogOKRect`, `valueDialogRect`, `valueDialogTextRect`: `property` - **done**. Rectangles should be inside the window and stable for repeated calls.
- `valueDialogSliderTrack`, `valueDialogDecrementRect`, `valueDialogIncrementRect`, `valueDialogOKRect`: `property` - **done**. Rectangles should be inside the dialog, non-overlapping where required, and stable for repeated calls.
- `pollTextDialogKeyboard`, `handleTextDialogControlKeys`, `handleBackspaceKey`, `handleSubmitKey`, `handleEscapeKey`, `handleJustPressedKey`, `valueDialogSubmitPressed`, `pollNumericTextFieldKeyboard`, `handleNumericTextFieldControlKeys`, `handleNumericTextFieldBackspace`, `handleNumericTextFieldCancel`, `handleNumericTextFieldSubmit`, `handleNumericTextFieldBlur`, `pollSaveFilenameDialogKeyboard`, `handleSaveFilenameDialogBackspace`, `handleSaveFilenameDialogSubmit`, `handleSaveFilenameDialogCancel`, `pollValueDialogKeyboard`, `handleValueDialogBackspace`, `handleValueDialogSubmit`, `handleValueDialogCancel`: `example`. Keyboard event ordering and Ebiten polling should stay fixed appunit scenarios.
- Appunit keyboard stubs in `save_filename_dialog_appunit.go` and `value_dialog_appunit.go`: `none`.

### `edit_clipboard.go`

- `copySelection`, `copySelectedMasses`, `copySelectedSprings`, `pasteSelectionAt`, `pasteClipboardMasses`, `pasteClipboardSprings`, `origin`, `nextMassID`, `nextSpringID`, `nextID`: `property` - **done**. Pasted IDs must be unique, springs must refer to pasted or existing masses correctly, and origin/next-ID calculations should be deterministic.

### `editor_screen.go`

- `EditorScreen`, `RegionPurpose`, `HasCommandControl`, `SetSelected`, `SelectSpring`, `SelectSprings`, `SetDirty`, `HandleShortcut`, `LastCommand`, `simulationState`, `selectionState`, `fileState`, `stateLabel`, `contains`: `example`. Screen reports and command shortcuts are fixed appunit scenarios; helper predicates are trivial.

### Gesture and Pointer Flow

- `handleWindowClose`, `handleRightPointer`, `handlePointer`, `handlePressedPointer`, `continuePointerPress`, `continueControlPress`, `releasePointer`, `beginPointerPress`, `clickOpenOverlay`, `overlayClick.run`, `openOverlayClicks`, `controlPointerPress`: `example`. Pointer orchestration is a state machine best covered by explicit gestures.
- `finishWorldPointer`, `finishMassDrag`, `throwDraggedMasses`, `throwSelectedDraggingMasses`, `throwSingleDraggingMass`, `beginMassDrag`, `captureDraggingOffsets`, `captureSelectedDraggingOffsets`, `pinDraggingMasses`: `example`. Drag/throw behavior is scenario based.
- `beginCanvasGesture`, `beginSelectGesture`, `finishSelectGesture`, `selectionClick`: `example`. Gesture thresholds and selection flow are fixed scenarios.
- `createMassAt`, `beginSpringAt`, `finishSpringAt`, `updateSpringChainEnd`, `beginControlPlacementAt`, `continueSpringChainAt`, `springChainEndpointAt`, `finishSpringChainStep`, `createSpringBetween`, `clearPendingSpring`, `massPosition`: `example`. Spring creation chains are discrete UI workflows.

### Context Menus and Object Updates

- `openContextAt`, `openMassContextMenu`, `simVec`, `clickMassContextMenu`, `massContextMenuItems`, `fixedToggleLabel`, `massContextMenuRect`, `massContextMenuRowRect`, `setMassFixed`, `setMassValue`, `updateMass`: `example`. Menu labels, row geometry, and selected-object changes are table/appunit scenarios.
- `openSpringContextMenu`, `clickSpringContextMenu`, `springContextMenuItems`, `toggleSpringWall`, `SpringContextMenuLabelsForSpring`, `SelectSpringContextMenuItem`, `springContextMenuRect`, `springContextMenuRowRect`: `example`. Same context-menu policy as mass menus.
- `updateByID`, `updateByIDAndMarkDirty`: `property` - **done**. Exactly matching items are updated, non-matches are untouched, and dirty marking occurs only on updates.

### `numeric_settings.go`

- `numericSettingControls`, `numericSettingRects`, `numericControlName`, `numericSettingForceToggleControl`, `numericSettingParameterToggleControl`, `numericSettingToggleControl`, `wallToggleControlsForSetting`, `numericSettingForSlider`, `numericSettingForStepButton`, `numericSettingForTextField`, `numericSettingForControl`, `numericSettingByName`, `numericSettingValue`, `numericSettingValueText`, `committedNumericSettingValueText`, `rawNumericSettingValue`, `formatNumericSettingText`, `numericSettingSliderFraction`, `setNumericSettingFromSlider`, `stepNumericSetting`, `formatNumericSettingSliderValue`, `isNumericInputCharacter`, `NumericSettingReport`, `NumericSettingText`, `NumericSettingSliderValue`, `inspectorRect`, `numericSettingReports`, `validateNumericSetting`: `property` - **done**. Slider fractions must remain in `[0,1]`; formatting should be idempotent after parse/format; lookup by generated name should round trip.
- `setNumericSettingValue`, `setParameterNumericSetting`, `focusNumericSettingTextField`, `appendNumericSettingInput`, `deleteNumericSettingCharacter`, `commitNumericSettingInput`, `cancelNumericSettingInput`, `focusedNumericSetting`, `appendNumericSettingCharacter`, `tickNumericTextField`, `numericTextCursorVisible`, `numericTextHighlighted`, `SetNumericSettingValue`, `ChangeNumericSettingWithSlider`, `FocusNumericSettingTextField`, `EnterNumericSettingText`, `TypeNumericSettingText`, `CommitNumericSettingText`: `example`. Text focus/edit/commit behavior is a UI state scenario.

### `render_world.go`

- `RenderWorld`, `renderRepresentations`, `springRepresentation`, `wallRepresentation`, `selectionRepresentation`, `centerRepresentation`, `HasVisibleRepresentation`, `massRepresentations`, `showSprings`, `hasEnabledWall`: `example`. Representation labels are deterministic render-report scenarios.
- `validSpring`, `springEndpoints`, `springIDEndpoints`, `springIndexEndpoints`, `validSpringIndex`: `property` - **done**. ID and index endpoint modes should agree for equivalent worlds; invalid endpoints should not render.

### `save_filename_dialog.go`

- `openSaveFilenameDialog`, `SaveFilenameDialogOpen`, `SaveFilenameText`, `SaveFilenameCursor`, `EnterSaveFilenamePrefix`, `insertSaveFilenameText`, `deleteSaveFilenameCharacter`, `clickSaveFilenameDialog`, `SubmitSaveFilenameDialog`, `CurrentFilePath`: `example`. Dialog lifecycle and file writing are fixed UI/I/O scenarios.
- `saveFilenamePath`: `example`. Filename validation and path resolution should be table driven.

### `startup_scene.go`

- `DefaultStartupScenePath`, `newDefaultStartupWorld`, `loadDefaultStartupWorld`, `defaultStartupSceneCandidates`: `property` - **done** through `internal/appcore`. The app wrappers should remain example-covered.

### `value_dialog.go`

- `openMassValueDialog`, `openSpringConstantDialogAt`, `openSpringValueDialog`, `springValueDialogSpec`, `tickValueDialog`, `valueDialogCursorVisible`, `SpringTemperatureDialogRange`, `ApplyValueDialogText`, `clickValueDialog`, `appendValueDialogInput`, `deleteValueDialogCharacter`, `applyValueDialog`, `continueValueDialogStepHold`, `setSpringConstant`, `setSpringDamping`, `setSpringRestLength`, `setSpringTemperature`, `setSpringFloat`, `applySpringFloat`, `updateSpring`, `springAt`, `springAtPosition`: `example`. Dialog and selected-spring mutations are stateful scenarios.
- `setValueDialogFromSlider`, `stepValueDialog`, `valueDialogFraction`, `distanceToSegment`: `property` - **done**. Slider results should stay bounded; stepping should clamp to dialog range; segment distance should be non-negative, symmetric under endpoint reversal, and zero on the segment.
- `drawValueDialog`, `drawValueDialogCursor`: `example`. Rendering belongs in appunit/headless checks.

### `visible_controls.go` and `visible_controls_draw.go`

- `isSliderControl`, `sliderTrack`, `sliderFraction`, `sliderLabel`, `activeControl`, `activeRunControl`, `activeForceControl`, `activeParameterControl`, `activeSelectedSpringControl`, `activeWallControl`, `visibleControls`, package-level `visibleControls`, `menuControls`, `editMenuControls`, `toolbarControls`, `commandControls`, `runPauseToggleLabel`, `inspectorControls`, `inspectorSections`, `sectionHeaderLabel`, `inspectorLeft`, `forceEnabled`, `parameterEnabled`, `wallEnabled`, `gridSnapEnabled`, `statusFields`, `objectCountsStatusLabel`, `currentFileStatusLabel`, `selectedObjectCount`, `DrawFrameReport`, `analyzeDrawnFrame`, `visibleActiveControls`, `visibleControlLabels`, `visibleInspectorSections`, `visibleInspectorSectionRects`, `visibleStatusFields`, `visibleRegionControlCounts`, `visibleLabelsFit`, `controlLabelsFit`, `statusLabelsFit`, `labelsFitItems`, `labelFits`, `visibleRegionRects`, `visibleRegionPixels`, `regionControlPixels`, `regionStatusPixels`, `rectPixels`, `visibleWorldPixels`: `property` - **done**. Visible controls should have stable labels/regions, active-state maps should agree with the source controls, labels should fit declared rectangles, and pixel counts should be non-negative and bounded.
- `drawVisibleControls`, `drawControl`, `drawSlider`, `drawNumericTextField`, `drawLabeledRect`, `drawCenteredLabeledRect`: `example`. Drawing is covered through render reports/headless checks.

## cmd

- `cmd/acceptance-generator`, `cmd/gherkin-parser`, `cmd/gherkin-mutator`, `cmd/springs`, `cmd/springs-check`: `example`. CLI parsing, exit codes, and output formats should stay fixture tested. Property tests are useful only indirectly through the underlying `internal/gherkin`, `internal/acceptancegen`, and `internal/acceptancemutation` packages.

## Recommendation

Do not try to property-test every function directly. Assess every function, but add property tests only where the function has a stable invariant over a broad input domain. The first tranche should be:

1. Keep the existing wall-spring no-crossing property test.
2. Add simulator conservation/no-tunneling properties for wall-spring collisions with unequal endpoint masses.
3. Add edit geometry properties for segment/box intersection.
4. Add XSP load/save round-trip properties.
5. Add mutation manifest parse/remove/merge properties.

That gives coverage over the bug-prone math and persistence surfaces without making CI slow or turning property tests into noisy duplicates of example tests.
