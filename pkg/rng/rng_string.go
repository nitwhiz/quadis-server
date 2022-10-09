package rng

type String struct {
	*Bag[string]
}

func NewString(seed int64, bagGen BagGenerator[string]) *String {
	return &String{
		NewBag[string](seed, bagGen),
	}
}
