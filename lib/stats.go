package lib

import (
	"fmt"
	"github.com/jmhodges/clock"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"math"
	"strings"
	"time"
)

type Stats struct {
	LastLoggedData gt7.GTData
	Laps           []Lap
	OngoingLap     Lap
	LastData       *gt7.GTData
	// ManualSetRaceDuration is the race duration manually set by the user if it is not
	// transmitted over telemetry
	ManualSetRaceDuration time.Duration
	raceStartTime         time.Time
	clock                 clock.Clock
}

func (s *Stats) getFuelConsumptionLastLap() (float32, error) {
	if len(s.Laps) < 1 {
		return -1, fmt.Errorf("not enough laps to return fuel consumption of last lap, nr of laps: %d", len(s.Laps))
	}
	return s.Laps[len(s.Laps)-1].FuelConsumed, nil
}

func NewStats() *Stats {
	s := Stats{}
	s.LastLoggedData = gt7.GTData{}
	s.LastData = &gt7.GTData{}
	// set a proper clock
	s.setClock(clock.New())
	return &s
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
	s.LastData = &gt7.GTData{}

	// Set empty ongoing lap
	s.raceStartTime = s.clock.Now()
}
func (s *Stats) GetAverageFuelConsumption() float32 {
	var totalFuelConsumption float32
	lapsAccountable := GetAccountableFuelConsumption(s.Laps)

	for _, f := range lapsAccountable {
		totalFuelConsumption += f
	}
	return totalFuelConsumption / float32(len(lapsAccountable))
}

func (s *Stats) GetFuelConsumptionPerMinute() (float32, error) {
	averageLapTime, err := s.GetAverageLapTime()
	if err != nil {
		return -1, fmt.Errorf("error getting average lap time: %v", err)
	}
	return s.GetAverageFuelConsumption() / float32(averageLapTime.Minutes()), nil
}

func (s *Stats) GetAverageLapTime() (time.Duration, error) {
	var totalDuration time.Duration
	accountableLaps := getAccountableLaps(s.Laps)
	for _, lap := range accountableLaps {
		totalDuration += lap.Duration
	}

	if len(accountableLaps) == 0 {
		return time.Duration(0), fmt.Errorf("no accounatble laps found")
	}

	return totalDuration / time.Duration(len(accountableLaps)), nil
}

const NoStartDetected = "Noch kein Start erfasst"

func (s *Stats) GetMessage() interface{} {

	timeSinceStart := ""
	errorMessages := []string{}

	if s.raceStartTime.IsZero() {
		timeSinceStart = NoStartDetected
	} else {
		timeSinceStart = GetSportFormat(s.GetTimeSinceStart())
	}

	minTemp := math.Min(float64(s.LastData.TyreTempFL), math.Min(float64(s.LastData.TyreTempFR), math.Min(float64(s.LastData.TyreTempRR), float64(s.LastData.TyreTempRL))))
	isValid := s.getValidState()

	lapsLeftInRace, err := s.getLapsLeftInRace()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Laps left in race unknown: %v", err))
		isValid = false
	}

	fuelNeededToFinishRaceInTotal, err := s.GetFuelNeededToFinishRaceInTotal()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Fuel needed to finish race unknown: %v", err))
		isValid = false
	}

	fuelDiv, err := s.GetFuelDiv()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Fuel Div unknown: %v", err))
		isValid = false
	}

	errorMessage := strings.Join(errorMessages, "\n")

	fuelConsumptionLastLap, err := s.getFuelConsumptionLastLap()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Fuel Consumption last lap unknown: %v", err))
		isValid = false
	}
	fuelConsumptionPerMinute, err := s.GetFuelConsumptionPerMinute()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Fuel Consumption per minute unknown: %v", err))
		isValid = false
	}

	message := Message{
		Speed:                    fmt.Sprintf("%.0f", s.LastData.CarSpeed),
		PackageID:                s.LastData.PackageID,
		FuelLeft:                 fmt.Sprintf("%.2f", s.LastData.CurrentFuel),
		FuelConsumptionLastLap:   fmt.Sprintf("%.2f", fuelConsumptionLastLap),
		FuelConsumptionAvg:       fmt.Sprintf("%.2f", s.GetAverageFuelConsumption()),
		FuelConsumptionPerMinute: fmt.Sprintf("%.2f", fuelConsumptionPerMinute),
		TimeSinceStart:           timeSinceStart,
		FuelNeededToFinishRace:   RoundUpAlways(fuelNeededToFinishRaceInTotal),
		LapsLeftInRace:           lapsLeftInRace,
		EndOfRaceType:            s.getEndOfRaceType(),
		FuelDiv:                  fmt.Sprintf("%.0f", fuelDiv),
		RaceTimeInMinutes:        int32(s.getRaceDuration().Minutes()),
		ValidState:               isValid,
		LowestTireTemp:           float32(minTemp),
		ErrorMessage:             errorMessage,
	}
	return message

}

func (s *Stats) getValidState() bool {
	validState := false

	fuelConsumptionLastLap, err := s.getFuelConsumptionLastLap()
	if err != nil {
		validState = false
	}
	if s.GetTimeSinceStart() < 1000*time.Hour && fuelConsumptionLastLap >= 0 {
		validState = true
	}

	return validState
}

const ByLaps = "By Laps"
const ByTime = "By Time"

func (s *Stats) getEndOfRaceType() string {

	endOfRaceType := ""
	if s.LastData.TotalLaps > 0 {
		endOfRaceType = ByLaps
	} else {
		endOfRaceType = ByTime
	}
	return endOfRaceType
}

func (s *Stats) getLapsLeftInRace() (int16, error) {

	if s.LastData.TotalLaps > 0 {
		return s.LastData.TotalLaps - s.LastData.CurrentLap + 1, nil // because the current lap is ongoing
	} else {

		if s.LastData.BestLap == 0 {
			return -1, fmt.Errorf("BestLap is 0, impossible to calculate laps left based on lap time")
		}

		bestLap := GetDurationFromGT7Time(s.LastData.BestLap)
		lapsLeftInRace, err := GetLapsLeftInRace(s.GetTimeSinceStart(), s.getRaceDuration(), bestLap)
		if err != nil {
			return -1, fmt.Errorf("error getting laps left: %v", err)
		}
		return lapsLeftInRace, nil
	}

}
func (s *Stats) getRaceDuration() time.Duration {
	if s.LastData.TotalLaps > 0 {
		return GetDurationFromGT7Time(s.LastData.BestLap) * time.Duration(s.LastData.TotalLaps)
	} else {
		return s.ManualSetRaceDuration
	}

}

func (s *Stats) SetManualSetRaceDuration(duration time.Duration) {
	s.ManualSetRaceDuration = duration
}

func (s *Stats) SetRaceStartTime(startTime time.Time) {
	s.raceStartTime = startTime

}

func (s *Stats) GetTimeSinceStart() time.Duration {
	return s.clock.Now().Sub(s.raceStartTime)
}

func (s *Stats) setClock(clock clock.Clock) {
	s.clock = clock
}

func (s *Stats) GetFuelNeededToFinishRaceInTotal() (float32, error) {

	if s.LastData.BestLap == 0 || s.LastData.LastLap == 0 {
		return -1, fmt.Errorf("BestLap or LastLap is 0, impossible to calculate fuel needed to finish race")
	}

	// it is best to use the last lap, since this will compensate for missed packages etc.
	fuelConsumptionLastLap, err := s.getFuelConsumptionLastLap()
	if err != nil {
		return -1, fmt.Errorf("error getting fuel consumption last lap: %v", err)
	}

	fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
		s.GetTimeSinceStart(),
		s.getRaceDuration(),
		GetDurationFromGT7Time(s.LastData.BestLap),
		GetDurationFromGT7Time(s.LastData.LastLap),
		fuelConsumptionLastLap)

	return fuelNeededToFinishRaceInTotal, nil

}

func (s *Stats) GetFuelDiv() (float32, error) {
	fuelNeededToFinishRaceInTotal, err := s.GetFuelNeededToFinishRaceInTotal()
	if err != nil {
		return -1, fmt.Errorf("error getting fuel needed to finish race: %v", err)
	}
	fuelDiv := fuelNeededToFinishRaceInTotal - s.LastData.CurrentFuel
	return fuelDiv, nil

}
