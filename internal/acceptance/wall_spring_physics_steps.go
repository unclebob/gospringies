package acceptance

func init() {
	for step, handler := range map[string]stepHandler{
		"the wall spring barriers task is accepted":                                                                                 acceptStep,
		"spring <spring_id> connects mass <mass_a> to mass <mass_b>":                                                                addBarrierSpring,
		"spring <spring_id> has Wall value <wall>":                                                                                  setBarrierSpringWall,
		"spring <spring_id> has Wall value false":                                                                                   setBarrierSpringWallFalse,
		"spring <spring_id> has Kspring <kspring> Kdamp <kdamp> RestLen <rest_len>":                                                 setBarrierSpringParameters,
		"the coder evaluates spring <spring_id> forces":                                                                             evaluateBarrierSpringForces,
		"spring <spring_id> should apply spring force state <spring_force_state>":                                                   assertBarrierSpringForceState,
		"spring <spring_id> should apply damping force state <damping_force_state>":                                                 assertBarrierSpringDampingState,
		"wall spring <spring_id> endpoints start <initial_length> apart with RestLen <rest_len>":                                    createWallSpringLengthConstraint,
		"the coder advances wall spring length constraint":                                                                          advanceWallSpringLengthConstraint,
		"wall spring <spring_id> endpoint distance should be <expected_length>":                                                     assertWallSpringEndpointDistance,
		"wall spring <spring_id> endpoint correction should be <correction_direction>":                                              assertWallSpringEndpointCorrection,
		"wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2>":                                           createWallSpringByCoordinates,
		"moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>":                                     createBarrierMovingMass,
		"fast moving mass <mass_id> starts at <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>":                                createFastBarrierMovingMass,
		"the coder advances through wall spring collision":                                                                          advanceThroughWallSpringCollision,
		"the coder advances through wall spring collision by <duration>":                                                            advanceThroughWallSpringCollisionByDuration,
		"mass <mass_id> should remain on the starting side of wall spring <spring_id>":                                              assertMassOnStartingWallSpringSide,
		"mass <mass_id> velocity should be resolved away from wall spring <spring_id>":                                              assertMassVelocityResolvedAwayFromWallSpring,
		"moving wall spring <spring_id> spans from <wall_x1>, <wall_y1> to <wall_x2>, <wall_y2> with velocity <wall_vx>, <wall_vy>": createMovingWallSpringByCoordinates,
		"stationary mass <mass_id> starts at <mass_x>, <mass_y>":                                                                    createBarrierStationaryMass,
		"the coder advances through moving wall spring collision":                                                                   advanceThroughWallSpringCollision,
		"mass <mass_id> should remain on the starting side of moving wall spring <spring_id>":                                       assertMassOnStartingWallSpringSide,
		"moving wall spring <spring_id> velocity should be resolved away from mass <mass_id>":                                       assertMovingWallSpringVelocityResolvedAwayFromMass,
		"wall spring <barrier_spring> spans from <barrier_x1>, <barrier_y1> to <barrier_x2>, <barrier_y2>":                          createBarrierWallSpringByCoordinates,
		"constrained wall spring <moving_spring> endpoint <endpoint_a> starts at <endpoint_a_x>, <endpoint_a_y>":                    createConstrainedWallSpringEndpointA,
		"constrained wall spring <moving_spring> endpoint <endpoint_b> starts at <endpoint_b_x>, <endpoint_b_y>":                    createConstrainedWallSpringEndpointB,
		"constrained wall spring <moving_spring> has RestLen <rest_len>":                                                            createConstrainedWallSpring,
		"the coder advances wall spring length constraints and collisions":                                                          advanceWallSpringLengthConstraintsAndCollisions,
		"wall spring endpoint <endpoint_a> should remain on the starting side of wall spring <barrier_spring>":                      assertWallSpringEndpointAOnStartingBarrierSide,
		"wall spring endpoint <endpoint_b> should remain on the starting side of wall spring <barrier_spring>":                      assertWallSpringEndpointBOnStartingBarrierSide,
		"wall spring <spring_id> spans from mass <endpoint_a> to mass <endpoint_b>":                                                 createWallSpringByEndpointIDs,
		"wall spring endpoint <endpoint_a> fixed state is <fixed_a>":                                                                setWallSpringEndpointFixed,
		"wall spring endpoint <endpoint_b> fixed state is <fixed_b>":                                                                setWallSpringEndpointBFixed,
		"moving mass <mass_id> collides with wall spring <spring_id> at contact fraction <contact_fraction>":                        createMassCollidingWithWallSpring,
		"the coder resolves the wall spring collision":                                                                              resolveWallSpringCollision,
		"wall spring endpoint <endpoint_a> should receive impulse share <impulse_share_a>":                                          assertWallSpringEndpointImpulseShare,
		"wall spring endpoint <endpoint_b> should receive impulse share <impulse_share_b>":                                          assertWallSpringEndpointBImpulseShare,
		"moving mass <mass_id> with elasticity <elasticity> collides with wall spring <spring_id> at normal speed <normal_speed>":   createElasticMassCollidingWithWallSpring,
		"mass <mass_id> normal rebound speed should be <expected_rebound_speed>":                                                    assertMassNormalReboundSpeed,
		"wall spring <spring_id> should receive collision impulse for rebound speed <expected_rebound_speed>":                       assertWallSpringReceivesReboundImpulse,
		"wall spring <spring_id> has Temperature <temperature>":                                                                     createWallSpringWithTemperature,
		"temperature random seed is <seed>":                                                                                         setTemperatureRandomSeed,
		"mass <mass_id> should receive temperature kick <kick_behavior>":                                                            assertMassTemperatureKick,
		"spring <spring_id> has Temperature <temperature>":                                                                          setSpringTemperature,
		"moving mass <mass_id> collides with spring <spring_id>":                                                                    createMassCollidingWithSpring,
		"the coder resolves spring collision":                                                                                       resolveSpringCollision,
	} {
		stepHandlers[step] = handler
	}
	for step, handler := range map[string]stepHandler{
		"fixed mass <fixed_mass> at <fixed_x>, <fixed_y> is an endpoint of wall spring <fixed_spring>":                                              createFixedWallEndpointMass,
		"moving wall spring <moving_spring> spans from <moving_x1>, <moving_y1> to <moving_x2>, <moving_y2> with velocity <moving_vx>, <moving_vy>": createMovingWallSpringTowardFixedEndpoint,
		"the simulation advances through fixed endpoint collision":                                                                                  advanceThroughFixedEndpointCollision,
		"moving wall spring <moving_spring> should remain on the starting side of fixed endpoint mass <fixed_mass>":                                 assertMovingWallSpringOnFixedEndpointStartingSide,
		"moving wall spring <moving_spring> contact point velocity should be resolved away from fixed endpoint mass <fixed_mass>":                   assertMovingWallSpringVelocityAwayFromFixedEndpoint,
	} {
		stepHandlers[step] = handler
	}
	for step, handler := range map[string]stepHandler{
		"a stationary floating wall with endpoint masses <endpoint_a_mass> and <endpoint_b_mass>":                                                                                                                        createUnequalMassFloatingWall,
		"moving mass <mass_id> with mass <moving_mass> is aimed at the floating wall from <mass_x>, <mass_y> with velocity <mass_vx>, <mass_vy>":                                                                         createMassAimedAtFloatingWall,
		"the simulation advances until the mass collides with the floating wall":                                                                                                                                         advanceUntilFloatingWallCollision,
		"the total momentum of the moving mass and floating wall endpoints is unchanged":                                                                                                                                 assertFloatingWallMomentumUnchanged,
		"floating wall spring <spring_id> moves from <previous_wall_x1>, <previous_wall_y1> and <previous_wall_x2>, <previous_wall_y2> to <current_wall_x1>, <current_wall_y1> and <current_wall_x2>, <current_wall_y2>": createSweptFloatingWallSpring,
		"moving mass <mass_id> moves from <previous_mass_x>, <previous_mass_y> to <current_mass_x>, <current_mass_y> with velocity <mass_vx>, <mass_vy>":                                                                 createSweptWallMovingMass,
		"the coder resolves swept floating wall spring collision":                                                                                                                                                        advanceSweptFloatingWallSpringCollision,
		"mass <mass_id> should remain on the previous side of floating wall spring <spring_id>":                                                                                                                          assertMassOnStartingWallSpringSide,
		"mass <mass_id> velocity should be resolved away from floating wall spring <spring_id>":                                                                                                                          assertMassVelocityResolvedAwayFromWallSpring,
		"moving floating wall spring <spring_id> has endpoint masses <endpoint_a_mass> and <endpoint_b_mass>":                                                                                                            createFiniteMassFloatingWallSpring,
		"moving floating wall spring <spring_id> endpoint velocities are <endpoint_a_vx>, <endpoint_a_vy> and <endpoint_b_vx>, <endpoint_b_vy>":                                                                          setFiniteMassFloatingWallSpringVelocities,
		"moving mass <mass_id> with mass <moving_mass> and elasticity <elasticity> collides with floating wall spring <spring_id> at contact fraction <contact_fraction> with velocity <mass_vx>, <mass_vy>":             createFiniteMassFloatingWallCollidingMass,
		"the coder resolves the finite-mass floating wall spring collision":                                                                                                                                              advanceFiniteMassFloatingWallSpringCollision,
		"the total kinetic energy of mass <mass_id> and floating wall spring <spring_id> should be <energy_behavior>":                                                                                                    assertFiniteMassFloatingWallSpringEnergy,
		"the total momentum of mass <mass_id> and floating wall spring <spring_id> should be unchanged":                                                                                                                  assertFiniteMassFloatingWallSpringMomentum,
	} {
		stepHandlers[step] = handler
	}
}
