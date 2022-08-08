package models

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

type Gauge float64

func (g Gauge) Type() string {
	return constants.MetricGauge
}

func (g Gauge) String() string {
	fg := float64(g)
	return fmt.Sprintf("%g", fg)
}
