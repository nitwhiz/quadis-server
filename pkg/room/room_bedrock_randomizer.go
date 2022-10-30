package room

func (d *BedrockDistribution) defaultRandomizer() {
	d.targetMap = map[string]string{}

	d.randomGameIdBag.NextBag()

	bSize := d.randomGameIdBag.GetSize()

	if bSize <= 1 {
		return
	}

	for i := 0; i < bSize; i++ {
		randomId := d.randomGameIdBag.NextElement()

		d.targetMap[randomId] = randomId
	}

	// just to be extra sure: new bag
	d.randomGameIdBag.NextBag()

	for id := range d.targetMap {
		target := d.randomGameIdBag.NextElement()

		if d.random.Probably(.1) {
			continue
		}

		d.targetMap[id] = target
	}
}

func (d *BedrockDistribution) deathMatchRandomizer() {
	defer d.room.gamesMutex.RUnlock()
	d.room.gamesMutex.RLock()

	d.targetMap = map[string]string{}

	d.randomGameIdBag.NextBag()

	bSize := d.randomGameIdBag.GetSize()

	if bSize <= 1 {
		return
	}

	for i := 0; i < bSize; i++ {
		randomId := d.randomGameIdBag.NextElement()

		d.targetMap[randomId] = randomId
	}

	var prevGameId string
	var lastTargetId string

	for gameId, targetId := range d.targetMap {
		if prevGameId == "" {
			prevGameId = gameId
			lastTargetId = targetId
			continue
		}

		d.targetMap[prevGameId] = targetId

		prevGameId = gameId
	}

	d.targetMap[prevGameId] = lastTargetId
}
