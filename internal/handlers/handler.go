package handlers

import (
	"github.com/go-chi/chi/v5"
)

func NewRepStore() chi.Router {
	rp := new(RepStore)
	rp.New()

	return rp.Router
}
