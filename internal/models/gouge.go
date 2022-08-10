package models

import (
	"fmt"
)

type Gauge float64

func (g Gauge) Type() string {
	return "gauge"
}

func (g Gauge) String() string {
	fg := float64(g)
	return fmt.Sprintf("%g", fg)
}
