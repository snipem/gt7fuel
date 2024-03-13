package lib

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"log"
	"time"
)

func LogRace(c *gt7.GT7Communication, gt7stats *Stats) {
	for {
		logTick(&c.LastData, gt7stats)
		time.Sleep(100 * time.Millisecond)
	}
}

func logTick(ld *gt7.GTData, gt7stats *Stats) bool {
	if ld.CurrentLap == 0 {
		// Race reset
		return false
	}

	if len(gt7stats.Laps) == 0 {
		gt7stats.Laps = append(gt7stats.Laps, Lap{
			FuelStart: ld.CurrentFuel,
			Number:    ld.CurrentLap,
		})
	}

	if gt7stats.Laps[len(gt7stats.Laps)-1].Number != ld.CurrentLap {
		// Change of laps detected

		if ld.CurrentLap == 1 {
			// First crossing of the line
			gt7stats.Reset()
			log.Printf("RACE START 🏁 %s \n", gt7stats.raceStartTime.Format("2006-01-02 15:04:05"))
		}

		gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd = ld.CurrentFuel
		gt7stats.Laps[len(gt7stats.Laps)-1].Duration = GetDurationFromGT7Time(ld.LastLap)

		// Do not log last laps fuel consumption in the first lap
		if ld.CurrentLap != 1 {
			fuelConsumptionLastLap := gt7stats.Laps[len(gt7stats.Laps)-1].FuelStart - gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd
			gt7stats.Laps[len(gt7stats.Laps)-1].FuelConsumed = fuelConsumptionLastLap
		}

		log.Printf("Add new Lap. Last Lap was: %s\n", gt7stats.Laps[len(gt7stats.Laps)-1])

		newLap := Lap{
			FuelStart: ld.CurrentFuel,
			Number:    ld.CurrentLap,
		}
		gt7stats.Laps = append(gt7stats.Laps, newLap)

	}
	// FIXME Use deep copy here
	gt7stats.LastLoggedData.FuelCapacity = ld.FuelCapacity
	return true
}
