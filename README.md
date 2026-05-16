# Springs

Springs is an XSpringies-style 2D spring-mass editor and simulation written in Go. The desktop app uses Ebitengine for the window, input, rendering, and real-time loop. Simulation, editing, and XSP file-format behavior live in plain Go packages so they can be tested without graphics.

## Local Commands

Run focused unit tests:

```sh
GOCACHE=/tmp/springs-gocache go test -timeout 120s ./internal/sim ./internal/format ./internal/edit ./internal/gherkin
```

Run the generated acceptance tests for a feature:

```sh
./scripts/acceptance.sh features/013_demo_files.feature
```

Run Gherkin mutation tests for the current feature:

```sh
./scripts/acceptance-mutate.sh
```

Build the desktop application:

```sh
GOCACHE=/tmp/springs-gocache go build -o /tmp/springs-app ./cmd/springs
```

Run a non-graphical command smoke check:

```sh
GOCACHE=/tmp/springs-gocache go run ./cmd/springs-check
```

Run the desktop application:

```sh
GOCACHE=/tmp/springs-gocache go run ./cmd/springs
```

## Desktop prerequisites

The app uses Ebitengine. A desktop run needs a graphical session and the native OpenGL/Metal stack available to Ebitengine. On macOS, run the app from a logged-in desktop session. On Linux, install the native development libraries Ebitengine documents for desktop builds, including X11/OpenGL audio-related packages as required by the target distribution.

## User Workflows

Creating a simulation starts on the editor screen. Use the visible mode controls to add masses, create springs between masses, select objects, and edit parameters. Fixed masses anchor structures; movable masses respond to active forces and springs.

Loading a simulation uses an XSP file. The repository includes demo files in `demos/`: `pendulum.xsp`, `spring-chain.xsp`, and `small-mesh.xsp`.

Saving a simulation writes the current world as deterministic XSP text. Fixed masses are represented with negative mass values in the file format while remaining explicit fixed-mass state in the domain model.

Running a simulation uses the pause control to start or stop time advancement. Rendering and input remain active while paused, so objects and parameters can be inspected and edited before resuming.

## Handoff Verification

For a task handoff, include each local verification command that was run and the result of each command. At minimum, report unit tests, acceptance tests, build, `git diff --check`, CRAP, DRY, and mutation verification when they are part of the task checklist.
