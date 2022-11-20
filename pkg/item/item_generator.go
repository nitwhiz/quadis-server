package item

import "github.com/nitwhiz/quadis-server/pkg/rng"

type Generator struct {
	*rng.Bag[*Item]
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		rng.NewBag[*Item](seed, func() []*Item {
			var b []*Item

			for _, i := range All {
				b = append(b, i)
			}

			return b
		}),
	}
}
