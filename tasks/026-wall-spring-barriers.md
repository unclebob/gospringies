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
- `Kspring`, `Kdamp`, and `RestLen` are ignored for physics while `Wall` is enabled.
- The barrier is the line segment between the spring's two endpoint masses.
- The barrier moves as those endpoint masses move.

## Collision Behavior

Masses collide with wall springs.

A wall spring should be treated as having zero mass for collision purposes.

When a mass collides with a wall spring:

- The colliding mass is prevented from crossing the wall segment.
- Its velocity is reflected or otherwise resolved so it no longer penetrates the wall.
- The collision impulse is transmitted to the endpoint masses based on the contact point's position along the wall segment.
- If the contact point is at fraction `t` from the first endpoint to the second endpoint, the first endpoint receives `(1 - t)` of the wall's response impulse and the second endpoint receives `t`.
- Fixed endpoint masses should not move, consistent with existing fixed-mass behavior.
- A fixed endpoint's impulse share is absorbed by that fixed endpoint and should not be redistributed to the other endpoint.

The wall itself does not have independent state or position beyond its two endpoint masses.

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
- Clicking the toggle changes the selected springs' `Wall` value.

## Spring Right-Click Menu

The spring right-click menu should include:

```text
Kspring
Kdamp
RestLen
Wall
```

`Wall` is a toggle item, not a slider dialog.

Selecting it toggles the `Wall` attribute for that spring.

## Rendering

Wall springs should be visually distinguishable from normal springs, for example with a heavier or different-colored line.

Normal springs keep the existing appearance.
