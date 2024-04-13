package lib

import (
	"fmt"
	"time"
)

func RoundUpAlways(d float32) int32 {
	simpleRoundRup := int32(d)
	diff := d - float32(simpleRoundRup)
	if diff > 0 {
		return simpleRoundRup + 1
	}
	return simpleRoundRup
}
func GetSportFormat(duration time.Duration) string {

	sign := ""

	if duration < 0 {
		duration = -duration
		sign = "-"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	milliseconds := duration.Milliseconds() % 1000

	if hours > 0 {
		return fmt.Sprintf("%s%02d:%02d:%02d.%03d", sign, hours, minutes, seconds, milliseconds)
	}

	return fmt.Sprintf("%s%02d:%02d.%03d", sign, minutes, seconds, milliseconds)

}
func GetDurationFromGT7Time(gt7time int32) time.Duration {
	seconds := gt7time / 1000
	milliseconds := gt7time % 1000

	return time.Duration(seconds)*time.Second + time.Duration(milliseconds*int32(time.Millisecond))

}
func GetAccountableFuelConsumption(laps []Lap) []float32 {

	var fuelConsumptionAccountable []float32

	for _, lap := range getAccountableLaps(laps) {
		fuelConsumptionAccountable = append(fuelConsumptionAccountable, lap.GetFuelConsumed())
	}
	return fuelConsumptionAccountable
}

func getAccountableLaps(laps []Lap) []Lap {
	var lapsAccountable []Lap
	for _, lap := range laps {
		if lap.GetFuelConsumed() >= 0 && lap.Number > 0 {
			lapsAccountable = append(lapsAccountable, lap)
		}
	}
	return lapsAccountable
}

// GetLapsLeftInRaceBasedOnTotalRaceDuration calculates the number of laps left in a race based on the current time in the race, total duration of the race, and the best lap time.
func GetLapsLeftInRaceBasedOnTotalRaceDuration(timeInRace time.Duration, totalDurationOfRace time.Duration, bestLapTime time.Duration) (int16, error) {
	// Calculate the time left in the race with an extra lap
	timeLeftInRace := getTimeLeftInRace(timeInRace, totalDurationOfRace, bestLapTime)

	if bestLapTime == 0 {
		return -1, fmt.Errorf("best Lap is 0 impossible to calculate Laps left in Race")
	}

	// Calculate the number of laps left based on the time left with the best lap time
	lapsLeft := timeLeftInRace / bestLapTime

	// If lapsLeft is negative, return 0, otherwise return lapsLeft as int16
	if lapsLeft < 0 {
		return 0, nil
	}
	return int16(lapsLeft), nil
}

// calculateFuelNeededToFinishRace calculates the fuel needed to finish the race.
//
// timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration, lastlaptime time.Duration, fuelconsumedlastlap float32 float32
func calculateFuelNeededToFinishRace(timeInRace time.Duration, totalDurationOfRace time.Duration, laptime time.Duration, fuelconsumedlastlap float32) float32 {

	fuelConsumptionQuota := fuelconsumedlastlap / float32(laptime.Milliseconds())
	timeLeftInRace := getTimeLeftInRace(timeInRace, totalDurationOfRace, laptime)
	return float32(timeLeftInRace.Milliseconds()) * fuelConsumptionQuota

}

func getTimeLeftInRace(timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration) time.Duration {
	timeLeftInRace := totalDurationOfRace - timeInRace
	return timeLeftInRace
}
