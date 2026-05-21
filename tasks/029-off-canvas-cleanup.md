# Off-Canvas Cleanup

Delete simulation objects that have left the canvas by more than one screen height.

## Cleanup Boundary

Use the current canvas height as the cleanup margin on every side of the canvas.

For a canvas with width `W` and height `H`, a mass is outside the cleanup boundary when its position is:

```text
x < -H
x > W + H
y < -H
y > H + H
```

The boundary is strict. A mass exactly at `x = -H`, `x = W + H`, `y = -H`, or `y = H + H` remains in the world.

## Masses

When cleanup runs, delete every mass outside the cleanup boundary.

Masses inside the cleanup boundary remain unchanged.

## Springs

When cleanup deletes a mass, delete every spring attached to that mass.

Springs whose endpoint masses remain inside the cleanup boundary remain in the world.

## Timing

Cleanup should run during normal simulation advancement so long-running scenes do not accumulate unreachable off-canvas objects.
