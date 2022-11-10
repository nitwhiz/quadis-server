package room

import (
	"log"
	"time"
)

const curfewActivityThreshold = time.Minute * 15

func (r *Room) StartCurfewBouncer() {
	r.NewHypervisor(&HypervisorConfig{
		StartType: HypervisorStartTypeLazy,
		Interval:  time.Second * 10,
	}, func() {
		lastRoomActivity := r.GetLastActivity()

		if lastRoomActivity.Add(curfewActivityThreshold).Before(time.Now()) {
			rId := r.GetId()

			log.Printf("stopping room %s ...\n", rId)

			r.Shutdown()

			log.Printf("room %s stopped\n", rId)
		}
	}).Start()
}
