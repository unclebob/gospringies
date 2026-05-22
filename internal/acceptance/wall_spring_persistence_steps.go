package acceptance

func init() {
	for step, handler := range map[string]stepHandler{
		"XSP input contains spring <spring_id> with Wall value <input_wall>":               createWallSpringXSPInput,
		"loaded spring <spring_id> should have Wall value <loaded_wall>":                   assertLoadedWallSpringXSP,
		"saved spring <spring_id> should include Wall value <saved_wall>":                  assertSavedWallSpringXSP,
		"XSP input contains spring <spring_id> with Temperature value <input_temperature>": createTemperatureSpringXSPInput,
		"loaded spring <spring_id> should have Temperature value <loaded_temperature>":     assertLoadedSpringTemperatureXSP,
		"saved spring <spring_id> should include Temperature value <saved_temperature>":    assertSavedSpringTemperatureXSP,
	} {
		stepHandlers[step] = handler
	}
}
