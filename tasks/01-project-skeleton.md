# Task 01: Project Skeleton

## Goal

Create the initial Go module layout for the application.

## Scope

- Initialize or confirm the Go module.
- Add a command entry point for the desktop app.
- Add internal packages for domain simulation, file format, and UI/app boundary.
- Keep Ebitengine imports out of domain packages.

## Acceptance Notes

- The repository has a clear Go package layout.
- `go test ./...` runs successfully.
- The app command can be built even if it only opens a placeholder or exits cleanly.

## Done When

- Module metadata is committed.
- Empty package tests or smoke tests prove the layout compiles.
