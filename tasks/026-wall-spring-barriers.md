# Wall Spring Barriers

## Wall Attribute

Add a boolean spring attribute named `Wall`.

When `Wall` is false, the spring behaves normally:

- Applies spring force from `Kspring`.
- Applies damping force from `Kdamp`.
- Uses `RestLen`.
- Can stretch, compress, and oscillate.

When `Wall` is true, the spring becomes a barrier:

- It is impenetrable to masses.
- It is inflexible.
- It does not apply normal spring or damping forces.
- `Kspring` and `Kdamp` are ignored for physics while `Wall` is enabled.
- `RestLen` is the wall's fixed length while `Wall` is enabled.
- If `RestLen` is zero or absent when `Wall` is enabled, the current endpoint distance becomes the wall's fixed length.
- The barrier is the line segment between the spring's two endpoint masses.
- The barrier moves as those endpoint masses move.

## Collision Behavior

Masses collide with wall springs.

A wall spring should be treated as having zero mass for collision purposes.

Collision detection must use relative motion between the wall spring segment and each mass. A wall spring whose segment sweeps across a stationary mass must collide with that mass instead of passing over it.

A mass that is an endpoint of one wall spring must still collide with every other wall spring for which it is not an endpoint. This includes endpoint motion caused by wall spring fixed-length correction, not only endpoint motion represented by velocity.

When a mass collides with a wall spring:

- The colliding mass is prevented from crossing the wall segment.
- Its velocity is reflected or otherwise resolved so it no longer penetrates the wall.
- The collision impulse is transmitted to the endpoint masses based on the contact point's position along the wall segment.
- If the contact point is at fraction `t` from the first endpoint to the second endpoint, the first endpoint receives `(1 - t)` of the wall's response impulse and the second endpoint receives `t`.
- Fixed endpoint masses should not move, consistent with existing fixed-mass behavior.
- A fixed endpoint's impulse share is absorbed by that fixed endpoint and should not be redistributed to the other endpoint.

The wall itself does not have independent state or position beyond its two endpoint masses.

## Temperature

Add a numeric spring attribute named `Temperature`.

Behavior:

- `Temperature` ranges from `0` to `10`.
- New springs default to `Temperature = 0`.
- Existing files without the attribute load with `Temperature = 0`.
- `Temperature` only affects wall springs.
- A non-wall spring may store a `Temperature` value, but it must not apply any temperature force while `Wall` is false.
- When a mass collides with a wall spring, the wall spring applies a random kick vector to the colliding mass based on `Temperature`.
- `Temperature = 0` applies no temperature kick.
- `Temperature = 10` applies a kick strong enough for a mass of `1` to travel the full screen height against gravity `10`.
- Intermediate temperatures scale the kick strength between those endpoints.
- Random temperature kick direction should come from the simulation's random source so tests can make it deterministic with a seed.

## Fixed Length Behavior

Wall spring endpoints are constrained to remain at the wall's fixed length.

The wall spring collision response and fixed-length response are separate:

- Collision response is applied perpendicular to the wall segment and is distributed by contact fraction.
- Fixed-length response is applied along the wall segment.
- The fixed-length response prevents endpoint movement that would stretch or compress the wall.
- The wall may translate or rotate when endpoint masses are not fixed.
- Fixed endpoint masses absorb any fixed-length response that would move them.

## Persistence

The `Wall` attribute must be saved and loaded in `.xsp` files.

Existing files without the attribute should load with:

```text
Wall = false
```

The save format should remain backward-compatible where practical.

## Right Controls

In the right-hand control panel, under `----- Selected Spring(s) -----`, add a `Wall` toggle.

Behavior:

- If one selected spring is a wall, the toggle shows active.
- If multiple selected springs are selected, use the existing multi-selection toggle convention.
- Clicking the toggle changes every selected spring's `Wall` value.
- If many springs are selected and the toggle is turned on, every selected spring becomes a wall.

## Spring Right-Click Menu

The spring right-click menu should include:

```text
Kspring
Kdamp
RestLen
Wall
Temperature
```

`Wall` is a toggle item, not a slider dialog.

Selecting it toggles the `Wall` attribute for that spring.

`Temperature` opens a value dialog with a slider like the other spring numeric items.

The `Temperature` dialog range is:

```text
0 to 10
```

## Rendering

Wall springs should be visually distinguishable from normal springs, for example with a heavier or different-colored line.

Normal springs keep the existing appearance.
