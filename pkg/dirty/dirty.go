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
	defer d.mu.Unlock()
	d.mu.Lock()

	d.value = true
}

// Clear returns the dirtiness state and clears it
func (d *Dirtiness) Clear() bool {
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.value {
		d.value = false

		return true
	}

	return false
}
