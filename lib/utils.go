package lib

import (
	"fmt"
	"sort"
	"time"
)

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
func GetSportFormat(duration time.Duration) string {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	milliseconds := duration.Milliseconds() % 1000

	// If hours are present, accumulate them into minutes
	minutes += hours * 60
	return fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, milliseconds)

}
func GetDurationFromGT7Time(gt7time int32) time.Duration {
	seconds := gt7time / 1000
	milliseconds := gt7time % 1000

	return time.Duration(seconds)*time.Second + time.Duration(milliseconds*int32(time.Millisecond))

}
func GetAccountableFuelConsumption(laps []Lap) []float32 {

	var fuelConsumptionAccountable []float32

	for _, lap := range getAccountableLaps(laps) {
		fuelConsumptionAccountable = append(fuelConsumptionAccountable, lap.FuelConsumed)
	}
	return fuelConsumptionAccountable
}

func getAccountableLaps(laps []Lap) []Lap {
	var lapsAccountable []Lap
	for _, lap := range laps {
		if lap.FuelConsumed > 0 && lap.Number > 0 {
			lapsAccountable = append(lapsAccountable, lap)
		}
	}
	return lapsAccountable
}
