package lib

import (
	"fmt"
	"os/exec"
	"runtime"
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

// GetLapsLeftInRace calculates the number of laps left in a race based on the current time in the race, total duration of the race, and the best lap time.
func GetLapsLeftInRace(timeInRace time.Duration, totalDurationOfRace time.Duration, bestLapTime time.Duration) int16 {
	// Calculate the time left in the race with an extra lap
	timeLeftWithExtraLap := getTimeLeftInRaceWithExtraLap(timeInRace, totalDurationOfRace, bestLapTime)

	if bestLapTime == 0 {
		return 0
	}

	// Calculate the number of laps left based on the time left with the best lap time
	lapsLeft := timeLeftWithExtraLap / bestLapTime

	// If lapsLeft is negative, return 0, otherwise return lapsLeft as int16
	if lapsLeft < 0 {
		return 0
	}
	return int16(lapsLeft)
}

// calculateFuelNeededToFinishRace calculates the fuel needed to finish the race.
//
// timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration, lastlaptime time.Duration, fuelconsumedlastlap float32 float32
func calculateFuelNeededToFinishRace(timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration, lastlaptime time.Duration, fuelconsumedlastlap float32) float32 {
	fuelConsumptionQuota := fuelconsumedlastlap / float32(lastlaptime.Milliseconds())
	timeLeftInRace := getTimeLeftInRaceWithExtraLap(timeInRace, totalDurationOfRace, bestlaptime)
	return float32(timeLeftInRace.Milliseconds()) * fuelConsumptionQuota

}

func getTimeLeftInRaceWithExtraLap(timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration) time.Duration {
	totalDurationPlusExtraLap := totalDurationOfRace + bestlaptime
	timeLeftInRace := totalDurationPlusExtraLap - timeInRace
	return timeLeftInRace
}

func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
