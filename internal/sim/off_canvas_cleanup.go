package sim

func (s *Simulation) cleanupOffCanvasObjects() {
	s.assignSpringEndpointIDs()
	deleted := map[int]bool{}
	keptMasses := make([]Mass, 0, len(s.Masses))
	for _, mass := range s.Masses {
		if s.massBeyondCleanupBoundary(mass) {
			deleted[mass.ID] = true
			continue
		}
		keptMasses = append(keptMasses, mass)
	}
	if len(deleted) == 0 {
		return
	}
	s.Masses = keptMasses
	s.removeSpringsAttachedTo(deleted)
	s.reindexSprings()
}

func (s *Simulation) assignSpringEndpointIDs() {
	for i := range s.Springs {
		if s.Springs[i].MassA == 0 && s.validMassIndex(s.Springs[i].A) {
			s.Springs[i].MassA = s.Masses[s.Springs[i].A].ID
		}
		if s.Springs[i].MassB == 0 && s.validMassIndex(s.Springs[i].B) {
			s.Springs[i].MassB = s.Masses[s.Springs[i].B].ID
		}
	}
}

func (s *Simulation) validMassIndex(index int) bool {
	return index >= 0 && index < len(s.Masses)
}

func (s *Simulation) massBeyondCleanupBoundary(mass Mass) bool {
	margin := s.Bounds.Height
	return mass.Position.X < s.Bounds.MinX()-margin ||
		mass.Position.X > s.Bounds.MaxX()+margin ||
		mass.Position.Y < s.Bounds.MinY()-margin ||
		mass.Position.Y > s.Bounds.MaxY()+margin
}

func (s *Simulation) removeSpringsAttachedTo(deleted map[int]bool) {
	keptSprings := make([]Spring, 0, len(s.Springs))
	for _, spring := range s.Springs {
		if deleted[spring.MassA] || deleted[spring.MassB] {
			continue
		}
		keptSprings = append(keptSprings, spring)
	}
	s.Springs = keptSprings
}

func (s *Simulation) reindexSprings() {
	for i := range s.Springs {
		a, okA := s.massIndexByID(s.Springs[i].MassA)
		b, okB := s.massIndexByID(s.Springs[i].MassB)
		if okA {
			s.Springs[i].A = a
		}
		if okB {
			s.Springs[i].B = b
		}
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:48:26-05:00","module_hash":"6a1d063e03759ccbd0368a3291c95a80e378201a1b859dd220411b40b19d8293","functions":[{"id":"func/Simulation.cleanupOffCanvasObjects","name":"Simulation.cleanupOffCanvasObjects","line":3,"end_line":20,"hash":"98a5d9b26bff1aa4ff09afd62fd3489aca46113b364cc8f64ff60ff32e29a042"},{"id":"func/Simulation.assignSpringEndpointIDs","name":"Simulation.assignSpringEndpointIDs","line":22,"end_line":31,"hash":"09bcddabd1707a470cc6546b31e79c0a01173d61f612c0fc870f174a4a109618"},{"id":"func/Simulation.validMassIndex","name":"Simulation.validMassIndex","line":33,"end_line":35,"hash":"d71d2fffb34a1beeed87fe7cacf2e1541b8f5711663b38fb9118b7e33789bb48"},{"id":"func/Simulation.massBeyondCleanupBoundary","name":"Simulation.massBeyondCleanupBoundary","line":37,"end_line":43,"hash":"5f8e814a0370f7b96399f11513aef2e0a98f094d20232e3f8dea158b01ce3d23"},{"id":"func/Simulation.removeSpringsAttachedTo","name":"Simulation.removeSpringsAttachedTo","line":45,"end_line":54,"hash":"9b98dfee347131eb232c817e8a3188310f549e11a265fcaa08023329968c9154"},{"id":"func/Simulation.reindexSprings","name":"Simulation.reindexSprings","line":56,"end_line":67,"hash":"d94aab6d885d571045fb84d7a8401410726691efe812151408b4da2188b30014"}]}
// mutate4go-manifest-end
