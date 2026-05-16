package edit

import (
	"fmt"

	"springs/internal/sim"
)

const (
	SpringButtonLeft   = "left"
	SpringButtonMiddle = "middle"
	SpringButtonRight  = "right"
	springHitRadius    = 8.0
)

type PendingSpring struct {
	StartMassID int
	Cursor      sim.Vec2
	Active      bool
	Temporary   bool
}

func (e *Editor) BeginSpring(position sim.Vec2, button string) error {
	if e.Mode != ModeAddSpring {
		return fmt.Errorf("unsupported spring mode %q", e.Mode)
	}
	id, ok := e.massNear(position)
	if !ok {
		return fmt.Errorf("no spring endpoint near pointer")
	}
	pending, err := newPendingSpring(id, position, button)
	if err != nil {
		return err
	}
	e.pendingSpring = &pending
	return nil
}

func (e *Editor) DragSpring(cursor sim.Vec2) {
	if e.pendingSpring != nil {
		e.pendingSpring.Cursor = cursor
	}
}

func (e *Editor) ReleaseSpring(position sim.Vec2) (int, bool, error) {
	pending := e.pendingSpring
	e.pendingSpring = nil
	if pending == nil {
		return 0, false, fmt.Errorf("no pending spring")
	}
	if pending.Temporary {
		return 0, false, nil
	}
	endID, ok := e.massNear(position)
	if !ok {
		return 0, false, nil
	}
	id, err := e.CreateSpring(pending.StartMassID, endID)
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (e *Editor) PendingSpring() (PendingSpring, bool) {
	if e.pendingSpring == nil {
		return PendingSpring{}, false
	}
	return *e.pendingSpring, true
}

func newPendingSpring(startMassID int, cursor sim.Vec2, button string) (PendingSpring, error) {
	switch button {
	case SpringButtonLeft:
		return PendingSpring{StartMassID: startMassID, Cursor: cursor, Active: true}, nil
	case SpringButtonMiddle:
		return PendingSpring{StartMassID: startMassID, Cursor: cursor, Active: true, Temporary: true}, nil
	case SpringButtonRight:
		return PendingSpring{StartMassID: startMassID, Cursor: cursor}, nil
	default:
		return PendingSpring{}, fmt.Errorf("unsupported spring button %q", button)
	}
}

func (e *Editor) massNear(position sim.Vec2) (int, bool) {
	id, ok := e.nearestMassID(position)
	if !ok {
		return 0, false
	}
	mass, _ := e.World.MassByID(id)
	if distance(mass.Position, position) > springHitRadius {
		return 0, false
	}
	return id, true
}
