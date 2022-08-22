package handlers

func NewRepStore() *RepStore {
	rp := new(RepStore)
	rp.New()

	return rp
}
