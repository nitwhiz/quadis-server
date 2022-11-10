package dirty

import "sync"

type Dirtiness struct {
	value bool
	mu    *sync.Mutex
}

func New() *Dirtiness {
	return &Dirtiness{
		value: false,
		mu:    &sync.Mutex{},
	}
}

// Trip sets the dirtiness state to true
func (d *Dirtiness) Trip() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.value = true
}

// Clear returns the dirtiness state and clears it
func (d *Dirtiness) Clear() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.value {
		d.value = false

		return true
	}

	return false
}
