package lib

import "sort"

func Median(data []float32) float32 {
	dataCopy := make([]float64, len(data))
	for i, v := range data {
		dataCopy[i] = float64(v)
	}

	sort.Float64s(dataCopy)

	var median float32
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = float32((dataCopy[l/2-1] + dataCopy[l/2]) / 2)
	} else {
		median = float32(dataCopy[l/2])
	}

	return median
}

func RoundUpAlways(d float32) int32 {
	simpleRoundRup := int32(d)
	diff := d - float32(simpleRoundRup)
	if diff > 0 {
		return simpleRoundRup + 1
	}
	return simpleRoundRup
}
