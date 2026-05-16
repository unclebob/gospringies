package acceptance

func booleanState(name string, states map[string]bool) (bool, bool) {
	value, ok := states[name]
	return value, ok
}
