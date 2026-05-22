package acceptance

func init() {
	for step, handler := range map[string]stepHandler{
		"selected spring <spring_id> has Wall value <old_wall>":                      createSelectedSpringWithWall,
		"selected springs <spring_ids> have Wall values <old_walls>":                 createSelectedSpringsWithWalls,
		"the coder changes spring control Wall to <new_wall>":                        changeSpringWallControl,
		"spring <spring_id> should have Wall value <new_wall>":                       assertSpringWallValue,
		"selected springs <spring_ids> should have Wall values <new_walls>":          assertSelectedSpringsWallValues,
		"spring <spring_id> has Wall value <old_wall>":                               createMenuSpringWithWall,
		"spring <spring_id> right-click menu includes item <menu_item>":              assertSpringMenuIncludesItem,
		"the coder selects spring menu item Wall for spring <spring_id>":             selectSpringMenuWallItem,
		"spring <spring_id> has Temperature value <old_temperature>":                 createMenuSpringWithTemperature,
		"the coder selects spring menu item Temperature for spring <spring_id>":      selectSpringMenuTemperatureItem,
		"spring Temperature dialog should open with range <minimum> to <maximum>":    assertSpringTemperatureDialogRange,
		"the coder changes the spring Temperature dialog value to <new_temperature>": changeSpringTemperatureDialogValue,
		"spring <spring_id> should have Temperature value <new_temperature>":         assertSpringTemperatureValue,
		"the coder renders spring <spring_id>":                                       renderWallSpring,
		"spring <spring_id> should use spring rendering style <rendering_style>":     assertWallSpringRenderingStyle,
	} {
		stepHandlers[step] = handler
	}
}
