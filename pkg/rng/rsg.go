package rng

type RSG struct {
	*RBG[string]
}

func NewRSG(seed int64, bagGen BagGenerator[string]) *RSG {
	return &RSG{
		NewRBG[string](seed, bagGen),
	}
}
