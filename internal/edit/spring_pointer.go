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

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:50:26-05:00","module_hash":"1da9e217fc90facdc74d1c03cfcb867850d673fe3a3aa33e582926ce715cd1c8","functions":[{"id":"func/Editor.BeginSpring","name":"Editor.BeginSpring","line":23,"end_line":37,"hash":"c0f41927dae47571ff01cf7d48afb3f5f0b470bc624f5adcedbda6e356054574"},{"id":"func/Editor.DragSpring","name":"Editor.DragSpring","line":39,"end_line":43,"hash":"3c3506338561c497d8bf9c7ecd27671e666201b3fb5bd9c9acb179d0d862eb20"},{"id":"func/Editor.ReleaseSpring","name":"Editor.ReleaseSpring","line":45,"end_line":63,"hash":"205f8a955beeb9589dda287bd2fc7945367fda881141ec0a20410fb7668bf8bb"},{"id":"func/Editor.PendingSpring","name":"Editor.PendingSpring","line":65,"end_line":70,"hash":"4b199b8aba09387c985995eb5306c72259455ba276b35b3a28903e96c02ee751"},{"id":"func/newPendingSpring","name":"newPendingSpring","line":72,"end_line":83,"hash":"f577d1b3ea2ec8cabc754e85733116b60176ca780221cecdc032e3fdfc16e3f0"},{"id":"func/Editor.massNear","name":"Editor.massNear","line":85,"end_line":95,"hash":"f9624be65bc3b4702b2bdfbb78886e22e9b88431a9c3f00ae9b8bf41f03eda59"}]}
// mutate4go-manifest-end
