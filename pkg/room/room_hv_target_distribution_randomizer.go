package room

func (t *TargetsDistribution) defaultRandomizer() {
	t.targetMap = map[string]string{}

	t.randomGameIdBag.NextBag()

	bSize := t.randomGameIdBag.GetSize()

	if bSize <= 1 {
		return
	}

	for i := 0; i < bSize; i++ {
		randomId := t.randomGameIdBag.NextElement()

		t.targetMap[randomId] = randomId
	}

	// just to be extra sure: new bag
	t.randomGameIdBag.NextBag()

	for id := range t.targetMap {
		target := t.randomGameIdBag.NextElement()

		if t.random.Probably(.1) {
			continue
		}

		t.targetMap[id] = target
	}
}

func (t *TargetsDistribution) deathMatchRandomizer() {
	t.targetMap = map[string]string{}

	t.randomGameIdBag.NextBag()

	bSize := t.randomGameIdBag.GetSize()

	if bSize <= 1 {
		return
	}

	for i := 0; i < bSize; i++ {
		randomId := t.randomGameIdBag.NextElement()

		t.targetMap[randomId] = randomId
	}

	var prevGameId string
	var lastTargetId string

	for gameId, targetId := range t.targetMap {
		if prevGameId == "" {
			prevGameId = gameId
			lastTargetId = targetId
			continue
		}

		t.targetMap[prevGameId] = targetId

		prevGameId = gameId
	}

	t.targetMap[prevGameId] = lastTargetId
}
