package app

func runIfPressed(pressed func() bool, run func()) {
	if pressed() {
		run()
	}
}
