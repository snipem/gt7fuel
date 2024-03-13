package lib

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"log"
	"time"
)

func LogRace(c *gt7.GT7Communication, gt7stats *Stats) {
	for {

		if c.LastData.CurrentLap == 0 {
			// Race reset

			time.Sleep(100 * time.Millisecond)
			continue
		}

		if len(gt7stats.Laps) == 0 {
			gt7stats.Laps = append(gt7stats.Laps, Lap{
				FuelStart: c.LastData.CurrentFuel,
				Number:    c.LastData.CurrentLap,
			})
		}

		if gt7stats.Laps[len(gt7stats.Laps)-1].Number != c.LastData.CurrentLap {
			// Change of laps detected

			if c.LastData.CurrentLap == 1 {
				// First crossing of the line
				gt7stats.Reset()
				log.Printf("RACE START 🏁 %s \n", gt7stats.raceStartTime.Format("2006-01-02 15:04:05"))
			}

			gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd = c.LastData.CurrentFuel
			gt7stats.Laps[len(gt7stats.Laps)-1].Duration = GetDurationFromGT7Time(c.LastData.LastLap)

			// Do not log last laps fuel consumption in the first lap
			if c.LastData.CurrentLap != 1 {
				fuelConsumptionLastLap := gt7stats.Laps[len(gt7stats.Laps)-1].FuelStart - gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd
				gt7stats.Laps[len(gt7stats.Laps)-1].FuelConsumed = fuelConsumptionLastLap
			}

			log.Printf("Add new Lap. Last Lap was: %s\n", gt7stats.Laps[len(gt7stats.Laps)-1])

			newLap := Lap{
				FuelStart: c.LastData.CurrentFuel,
				Number:    c.LastData.CurrentLap,
			}
			gt7stats.Laps = append(gt7stats.Laps, newLap)

		}
		// FIXME Use deep copy here
		gt7stats.LastLoggedData.FuelCapacity = c.LastData.FuelCapacity
		time.Sleep(100 * time.Millisecond)
	}
}
