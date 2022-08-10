package models

import (
	"fmt"
)

type Counter int64

func (c Counter) Add(x int64) Counter {
	return Counter(int64(c) + x) //sync.Mutex{} //
}

func (c Counter) Type() string {
	return "counter"
}

func (c Counter) TypeFromString() string {
	return "counter"
}

func (c Counter) String() string {
	return fmt.Sprintf("%d", int64(c))
}
