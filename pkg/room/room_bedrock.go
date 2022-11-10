package room

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
	"log"
	"time"
)

type BedrockDistribution struct {
	targets *TargetsDistribution
	room    *Room
	Channel chan *game.Bedrock
}

func (r *Room) StartBedrockDistribution() {
	if r.bedrockDistribution != nil {
		log.Fatalln("trying to init another bedrock distribution")
	}

	bd := BedrockDistribution{
		targets: r.targets,
		room:    r,
		Channel: make(chan *game.Bedrock, 64),
	}

	go bd.startDistribution()

	r.bedrockDistribution = &bd
}

func (d *BedrockDistribution) startDistribution() {
	d.room.wg.Add(1)
	defer d.room.wg.Done()

	for {
		select {
		case <-d.room.ctx.Done():
			return
		case b := <-d.Channel:
			targetGameId := d.targets.GetTargetGameId(b.SourceId)

			if targetGameId != "" && targetGameId != b.SourceId {
				d.room.gamesMutex.RLock()

				if targetGame, ok := d.room.games[targetGameId]; ok {
					targetGame.AddBedrock(b.Amount)
				}

				d.room.gamesMutex.RUnlock()
			}

			// avoid lock up
			time.Sleep(time.Microsecond * 250)
			break
		}
	}
}
