package lib

import (
	"fmt"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"time"
)

type Stats struct {
	LastLoggedData gt7.GTData
	Laps           []Lap
	LastData       *gt7.GTData
	// ManualSetRaceDuration is the race duration manually set by the user if it is not
	// transmitted over telemetry
	ManualSetRaceDuration  time.Duration
	raceStartTime          time.Time
	fuelConsumptionLastLap float32
}

type Lap struct {
	FuelStart    float32
	FuelEnd      float32
	FuelConsumed float32
	Number       int16
	Duration     time.Duration
}

func (l Lap) String() string {
	return fmt.Sprintf("Lap %d: FuelStart=%.2f, FuelEnd=%.2f, FuelConsumed=%.2f, Duration=%s",
		l.Number, l.FuelStart, l.FuelEnd, l.FuelConsumed, l.Duration)
}

func (s *Stats) Reset() {
	s.LastLoggedData = gt7.GTData{}
	// Set empty ongoing lap
	s.Laps = []Lap{{FuelStart: 0, FuelEnd: 0, FuelConsumed: 0, Number: 0, Duration: 0}}
	s.raceStartTime = time.Now()
	s.fuelConsumptionLastLap = 0
}
func (s *Stats) GetAverageFuelConsumption() float32 {
	var totalFuelConsumption float32
	lapsAccountable := GetAccountableFuelConsumption(s.Laps)

	for _, f := range lapsAccountable {
		totalFuelConsumption += f
	}
	return totalFuelConsumption / float32(len(lapsAccountable))
}

func (s *Stats) GetFuelConsumptionPerMinute() float32 {
	return s.GetAverageFuelConsumption() / float32(s.GetAverageLapTime().Minutes())
}

func (s *Stats) GetAverageLapTime() time.Duration {
	var totalDuration time.Duration
	accountableLaps := getAccountableLaps(s.Laps)
	for _, lap := range accountableLaps {
		totalDuration += lap.Duration
	}

	if len(accountableLaps) == 0 {
		return 0
	}

	return totalDuration / time.Duration(len(accountableLaps))
}

func (s *Stats) GetMessage() interface{} {

	timeSinceStart := ""

	if s.raceStartTime.IsZero() {
		timeSinceStart = "Noch kein Start erfasst"
	} else {
		timeSinceStart = GetSportFormat(s.GetTimeSinceStart())
	}

	message := Message{
		Speed:                    fmt.Sprintf("%.0f", s.LastData.CarSpeed),
		PackageID:                s.LastData.PackageID,
		FuelLeft:                 fmt.Sprintf("%.2f", s.LastData.CurrentFuel),
		FuelConsumptionLastLap:   fmt.Sprintf("%.2f", s.fuelConsumptionLastLap),
		FuelConsumptionAvg:       fmt.Sprintf("%.2f", s.GetAverageFuelConsumption()),
		FuelConsumptionPerMinute: fmt.Sprintf("%.2f", s.GetFuelConsumptionPerMinute()),
		TimeSinceStart:           timeSinceStart,
		FuelNeededToFinishRace:   RoundUpAlways(s.GetFuelNeededToFinishRaceInTotal()),
		LapsLeftInRace:           s.getLapsLeftInRace(),
		EndOfRaceType:            s.getEndOfRaceType(),
		FuelDiv:                  fmt.Sprintf("%.0f", s.GetFuelDiv()),
		RaceTimeInMinutes:        int32(s.getRaceDuration().Minutes()),
		ValidState:               s.getValidState(),
	}
	return message

}

func (s *Stats) getValidState() bool {
	validState := true

	if s.GetTimeSinceStart() > 1000*time.Hour {
		validState = false
	} else if s.fuelConsumptionLastLap < 0 {
		validState = false
	}
	return validState
}

func (s *Stats) getEndOfRaceType() string {

	endOfRaceType := ""
	if s.LastData.TotalLaps > 0 {
		endOfRaceType = "By Laps"
	} else {
		endOfRaceType = "By Time"
	}
	return endOfRaceType
}

func (s *Stats) getLapsLeftInRace() int16 {
	if s.LastData.TotalLaps > 0 {
		return s.LastData.TotalLaps - s.LastData.CurrentLap + 1 // because the current lap is ongoing
	} else {
		bestLap := GetDurationFromGT7Time(s.LastData.BestLap)
		return GetLapsLeftInRace(s.GetTimeSinceStart(), s.getRaceDuration(), bestLap)
	}

}
func (s *Stats) getRaceDuration() time.Duration {
	var raceDuration time.Duration
	if s.LastData.TotalLaps > 0 {
		raceDuration = GetDurationFromGT7Time(s.LastData.BestLap) * time.Duration(s.LastData.TotalLaps)
	} else {
		return s.ManualSetRaceDuration
	}
	return raceDuration

}

func (s *Stats) SetManualSetRaceDuration(duration time.Duration) {
	s.ManualSetRaceDuration = duration
}

func (s *Stats) SetRaceStartTime(startTime time.Time) {
	s.raceStartTime = startTime

}

func (s *Stats) GetTimeSinceStart() time.Duration {
	return time.Now().Sub(s.raceStartTime)
}

func (s *Stats) GetFuelNeededToFinishRaceInTotal() float32 {

	// it is best to use the last lap, since this will compensate for missed packages etc.
	fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
		s.GetTimeSinceStart(),
		s.getRaceDuration(),
		GetDurationFromGT7Time(s.LastData.BestLap),
		GetDurationFromGT7Time(s.LastData.LastLap),
		s.fuelConsumptionLastLap)

	return fuelNeededToFinishRaceInTotal

}

func (s *Stats) SetFuelConsumptionLastLap(fuelConsumption float32) {
	s.fuelConsumptionLastLap = fuelConsumption
}

func (s *Stats) GetFuelDiv() any {
	fuelDiv := s.GetFuelNeededToFinishRaceInTotal() - s.LastData.CurrentFuel
	return fuelDiv

}
