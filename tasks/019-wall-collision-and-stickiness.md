# Task 019: Wall Collision And Stickiness

## Goal

Match XSpringies wall collision, one-way wall, elasticity, and stickiness behavior.

## Scope

- Enabled walls are located at current window boundaries.
- A mass moving from inside the screen toward an enabled wall bounces.
- Enabled walls use swept-path collision detection: if the segment from a mass's previous position to its current timestep position crosses an enabled wall, the mass collides even when its final position is already beyond the wall.
- The wall collision response uses the side from the previous position, so the mass is resolved back toward the side it came from.
- The wall-normal velocity component reverses on bounce.
- Bounced wall-normal velocity is scaled by mass Elasticity.
- Walls are one-way: a mass moving from off-screen toward the screen passes through.
- Stickiness reduces wall-normal velocity on collision.
- If stickiness removes all wall-normal velocity, the mass remains stuck to the wall.
- A stuck mass is released when sufficient opposing force pulls it off the wall.

## Acceptance Notes

- Disabled walls do not collide or repel.
- Fixed masses remain fixed even at boundaries.

## Done When

- Wall bounce, one-way behavior, elasticity scaling, sticky capture, sticky release, and disabled walls are covered.
