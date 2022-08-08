package models

import (
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

type Counter int64

func (c Counter) Add(x int64) Counter {
	return Counter(int64(c) + x)
}

func (c Counter) Type() string {
	return constants.MetricCounter
}

func (c Counter) String() string {
	return fmt.Sprintf("%d", int64(c))
}
