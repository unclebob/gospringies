package acceptance

func booleanState(name string, states map[string]bool) (bool, bool) {
	value, ok := states[name]
	return value, ok
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-16T12:14:15-05:00","module_hash":"f5edfc5dfa335307227794b39728a423c468038036a17f54971f433f5081004b","functions":[{"id":"func/booleanState","name":"booleanState","line":3,"end_line":6,"hash":"468ed185232971ed7329f55a28c4cc2b9f00f52a219e6b3b63421fd51043a6d5"}]}
// mutate4go-manifest-end
