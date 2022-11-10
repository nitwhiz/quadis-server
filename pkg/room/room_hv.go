package room

import (
	"time"
)

type HypervisorStartType = byte

const HypervisorStartTypeLazy = HypervisorStartType(0)
const HypervisorStartTypeInstant = HypervisorStartType(10)

type HypervisorHandlerFunc = func()

type Hypervisor struct {
	startType HypervisorStartType
	interval  time.Duration
	room      *Room
	handler   HypervisorHandlerFunc
}

type HypervisorConfig struct {
	Interval  time.Duration
	StartType HypervisorStartType
}

func (r *Room) NewHypervisor(config *HypervisorConfig, handlerFunc HypervisorHandlerFunc) *Hypervisor {
	return &Hypervisor{
		startType: config.StartType,
		interval:  config.Interval,
		room:      r,
		handler:   handlerFunc,
	}
}

func (h *Hypervisor) runner() {
	h.room.wg.Add(1)
	defer h.room.wg.Done()

	if h.startType == HypervisorStartTypeInstant {
		go h.handler()
	}

	for {
		select {
		case <-h.room.ctx.Done():
			return
		case <-time.After(h.interval):
			go h.handler()
			break
		}
	}
}

func (h *Hypervisor) Start() {
	go h.runner()
}
