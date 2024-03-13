package lib

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"log"
)

func LogTick(ld *gt7.GTData, gt7stats *Stats) bool {
	//if ld.CurrentLap == 0 {
	//	// Race reset
	//	return false
	//}

	//if ld.CurrentLap > 0 && len(gt7stats.Laps) == 0 {
	//	gt7stats.Laps = append(gt7stats.Laps, Lap{
	//		FuelStart: ld.CurrentFuel,
	//		Number:    ld.CurrentLap,
	//	})
	//}

	//if len(gt7stats.Laps) == 0 {
	//	return false
	//}

	if gt7stats.LastLoggedData.CurrentLap == 0 && ld.CurrentLap == 1 {
		// First crossing of the line
		gt7stats.Reset()

		gt7stats.OngoingLap = Lap{
			FuelStart: ld.CurrentFuel,
			Number:    ld.CurrentLap,
		}

		log.Printf("RACE START 🏁 %s \n", gt7stats.raceStartTime.Format("2006-01-02 15:04:05"))
	}

	if gt7stats.OngoingLap.Number != ld.CurrentLap {
		// Change of laps detected
		finishLap(ld, gt7stats)
	}

	// FIXME Use deep copy here
	gt7stats.LastLoggedData.FuelCapacity = ld.FuelCapacity
	gt7stats.LastLoggedData.CurrentLap = ld.CurrentLap
	return true
}

func finishLap(ld *gt7.GTData, gt7stats *Stats) {
	gt7stats.OngoingLap.FuelEnd = ld.CurrentFuel
	gt7stats.OngoingLap.Duration = GetDurationFromGT7Time(ld.LastLap)

	// Do not log last laps fuel consumption in the first lap
	if ld.CurrentLap != 1 {
		fuelConsumptionLastLap := gt7stats.OngoingLap.FuelStart - gt7stats.OngoingLap.FuelEnd
		gt7stats.OngoingLap.FuelConsumed = fuelConsumptionLastLap
	}

	log.Printf("Add new Lap. Last Lap was: %s\n", gt7stats.OngoingLap)

	gt7stats.Laps = append(gt7stats.Laps, gt7stats.OngoingLap)
	gt7stats.OngoingLap = Lap{
		FuelStart: ld.CurrentFuel,
		Number:    ld.CurrentLap,
	}
}
