package main

import metrics "runtime/metrics"

func main() {

	v := metrics.All()
	samples := make([]metrics.Sample, len(v))
	for i := range samples {
		samples[i].Name = v[i].Name
	}
	metrics.Read(samples)
}
