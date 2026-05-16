package sim

import "testing"

func TestStepMovesFreeMassTowardSpringRestLength(t *testing.T) {
	s := NewSimulation()
	left := s.AddMass(Vec2{X: 0, Y: 0}, 1, true)
	right := s.AddMass(Vec2{X: 120, Y: 0}, 1, false)
	s.AddSpring(left, right, 100, 10)

	s.Step(0.1)

	if got := s.Masses[right].Position.X; got >= 120 {
		t.Fatalf("free mass x = %f, expected it to move left", got)
	}
	if got := s.Masses[left].Position.X; got != 0 {
		t.Fatalf("fixed mass x = %f, expected it to stay fixed", got)
	}
}

func TestAdvanceIsDeterministic(t *testing.T) {
	first := NewDemoSimulation()
	second := NewDemoSimulation()

	first.Advance(10, 0.016)
	second.Advance(10, 0.016)

	if first.Masses[1].Position.X != second.Masses[1].Position.X {
		t.Fatalf("advance not deterministic: %f != %f", first.Masses[1].Position.X, second.Masses[1].Position.X)
	}
}
