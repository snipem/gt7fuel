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
	ConnectionActive      bool
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
	s.ConnectionActive = false
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
	LapStart     time.Time
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
func (s *Stats) GetAverageFuelConsumptionPerLap() (avgFuelConsumption float32, err error) {
	var totalFuelConsumption float32
	lapsAccountable := GetAccountableFuelConsumption(s.Laps)

	for _, f := range lapsAccountable {
		totalFuelConsumption += f
	}

	if len(lapsAccountable) == 0 {
		return -1, fmt.Errorf("no accountable laps found")
	}

	return totalFuelConsumption / float32(len(lapsAccountable)), err
}

func (s *Stats) GetFuelConsumptionPerMinute() (float32, error) {
	averageLapTime, err := s.GetAverageLapTime()
	if err != nil {
		return -1, fmt.Errorf("error getting average lap time: %v", err)
	}

	avgFuelConsumption, err := s.GetAverageFuelConsumptionPerLap()
	if err != nil {
		return -1, fmt.Errorf("error getting average fuel consumption: %v", err)
	}

	return avgFuelConsumption / float32(averageLapTime.Minutes()), nil
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

	isValid := s.getValidState()

	durationSinceStart, err := s.GetDurationSinceStart()
	if err != nil {
		timeSinceStart = NoStartDetected
		isValid = false
	} else {
		timeSinceStart = GetSportFormat(durationSinceStart)
	}

	minTemp := math.Min(float64(s.LastData.TyreTempFL), math.Min(float64(s.LastData.TyreTempFR), math.Min(float64(s.LastData.TyreTempRR), float64(s.LastData.TyreTempRL))))

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

	avgFuelConsumption, err := s.GetAverageFuelConsumptionPerLap()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Avg Fuel Consumption unknown: %v", err))
		isValid = false
	}

	nextPitStop, err := s.GetNextNecessaryPitStopAtEndOfLap()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Next Pit Stop unknown: %v", err))
		isValid = false
	}

	currentLapProgressAdjusted, err := s.GetProgressAdjustedCurrentLap()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Current Lap Progress unknown: %v", err))
		isValid = false
	}

	message := Message{
		Speed:                      fmt.Sprintf("%.0f", s.LastData.CarSpeed),
		PackageID:                  s.LastData.PackageID,
		FuelLeft:                   fmt.Sprintf("%.2f", s.LastData.CurrentFuel),
		FuelConsumptionLastLap:     fmt.Sprintf("%.2f", fuelConsumptionLastLap),
		FuelConsumptionAvg:         fmt.Sprintf("%.2f", avgFuelConsumption),
		FuelConsumptionPerMinute:   fmt.Sprintf("%.2f", fuelConsumptionPerMinute),
		TimeSinceStart:             timeSinceStart,
		FuelNeededToFinishRace:     RoundUpAlways(fuelNeededToFinishRaceInTotal),
		LapsLeftInRace:             lapsLeftInRace,
		EndOfRaceType:              s.getEndOfRaceType(),
		FuelDiv:                    fmt.Sprintf("%.0f", fuelDiv),
		RaceTimeInMinutes:          int32(s.getRaceDuration().Minutes()),
		ValidState:                 isValid,
		LowestTireTemp:             float32(minTemp),
		ErrorMessage:               errorMessage,
		NextPitStop:                int16(nextPitStop),
		CurrentLapProgressAdjusted: fmt.Sprintf("%.0f", currentLapProgressAdjusted),
	}
	return message

}

func (s *Stats) getValidState() bool {
	validState := true

	fuelConsumptionLastLap, err := s.getFuelConsumptionLastLap()
	if err != nil {
		validState = false
	}

	durationSinceStart, err := s.GetDurationSinceStart()
	if err != nil {
		validState = false
	}

	if durationSinceStart > 1000*time.Hour || fuelConsumptionLastLap < 0 {
		validState = false
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

		bestLap, err := s.getBestLap()
		if err != nil {
			return -1, fmt.Errorf("BestLap is 0, impossible to calculate laps left based on lap time")
		}

		durationSinceStart, err := s.GetDurationSinceStart()
		if err != nil {
			return -1, fmt.Errorf("error getting duration since start: %v", err)
		}

		lapsLeftInRace, err := GetLapsLeftInRace(durationSinceStart, s.getRaceDuration(), bestLap)
		if err != nil {
			return -1, fmt.Errorf("error getting laps left: %v", err)
		}
		return lapsLeftInRace, nil
	}

}

func (s *Stats) getBestLap() (time.Duration, error) {

	if s.LastData.BestLap == 0 {
		return 0, fmt.Errorf("BestLap is 0, impossible to calculate best lap")
	}

	bestLap := GetDurationFromGT7Time(s.LastData.BestLap)
	return bestLap, nil
}

func (s *Stats) getTotalLapsInRace() (int16, error) {

	if s.LastData.TotalLaps > 0 {
		return s.LastData.TotalLaps, nil
	}

	bestLap, err := s.getBestLap()
	if err != nil {
		return -1, fmt.Errorf("BestLap is 0, impossible to calculate total laps based on lap time")
	}

	// we assume the race hos not started yet to get the total number of laps

	lapsLeftInRace, err := GetLapsLeftInRace(time.Duration(0), s.getRaceDuration(), bestLap)
	if err != nil {
		return -1, fmt.Errorf("error getting laps left: %v", err)
	}
	return lapsLeftInRace, nil
}
func (s *Stats) getRaceDuration() time.Duration {
	if s.LastData.TotalLaps > 0 {
		return GetDurationFromGT7Time(s.LastData.BestLap) * time.Duration(s.LastData.TotalLaps)
	} else {
		return s.ManualSetRaceDuration
	}
}

func (s *Stats) GetNextNecessaryPitStopAtEndOfLap() (int, error) {

	avgFuelConsumptionPerLap, err := s.GetAverageFuelConsumptionPerLap()
	if err != nil {
		return -1, fmt.Errorf("error getting average fuel consumption per lap: %v", err)
	}
	progressAdjustedCurrentLap, err := s.GetProgressAdjustedCurrentLap()
	if err != nil {
		return -1, fmt.Errorf("error getting next neccessary pit stop: %v", err)
	}
	nextPitStop := getNextPitStop(s.LastData.CurrentFuel, avgFuelConsumptionPerLap, progressAdjustedCurrentLap)
	return nextPitStop, nil
}

func getNextPitStop(currentFuel float32, avgFuelConsumptionPerLap float32, progressAdjustedCurrentLap float32) int {
	if avgFuelConsumptionPerLap == 0 {
		return -1
	}

	currentFuel64 := float64(currentFuel)
	avgFuelConsumptionPerLap64 := float64(avgFuelConsumptionPerLap)
	progressAdjustedCurrentLap64 := float64(progressAdjustedCurrentLap)

	progressInCurrentLap := math.Ceil(progressAdjustedCurrentLap64) - progressAdjustedCurrentLap64

	fuelNeededToFinishCurrentLap := avgFuelConsumptionPerLap64 * progressInCurrentLap

	currentLapFloor := math.Floor(progressAdjustedCurrentLap64)

	// It is not possible to attempt another lap
	if currentFuel64 <= fuelNeededToFinishCurrentLap+avgFuelConsumptionPerLap64 {
		return int(currentLapFloor)
	}
	nextPitStop := currentLapFloor + (currentFuel64 / avgFuelConsumptionPerLap64)

	return int(math.Floor(nextPitStop))

	//print(fuelNeededToFinishCurrentLap)

	//currentLapFloor := math.Floor(float64(progressAdjustedCurrentLap))
	//
	//// 3.5 > 3 --> current lap
	//if currentFuel > avgFuelConsumptionPerLap {
	//	fuelLeftForNumberOfLaps := currentFuel / avgFuelConsumptionPerLap
	//	return int(currentLapFloor + math.Floor(float64(fuelLeftForNumberOfLaps)))
	//} else {
	//	return int(currentLapFloor)
	//}
	//
	////ceilOfCurrentLap := math.Floor(float64(progressAdjustedCurrentLap))
	////ceilOfFullLapsLeft := math.Floor(float64(fuelLeftForNumberOfLaps - 1))
	////kk
	//return 0
}

func (s *Stats) GetProgressAdjustedCurrentLap() (float32, error) {

	if s.OngoingLap.LapStart.IsZero() {
		return float32(-1), fmt.Errorf("LapStart is Zero, impossible to calculate Lap progress")
	}

	durationInCurrentLap := s.clock.Now().Sub(s.OngoingLap.LapStart)

	bestLap, err := s.getBestLap()
	if err != nil {
		return -1, fmt.Errorf("impossible to calculate progress adjusted current lap: %v", err)
	}
	return getDurationInLap(durationInCurrentLap, bestLap, s.LastData.CurrentLap)

}

func getDurationInLap(durationInCurrentLap time.Duration, bestLap time.Duration, currentLap int16) (float32, error) {
	relativeProgressCurrentLap := float32(durationInCurrentLap) / float32(bestLap)

	if relativeProgressCurrentLap > 1 {
		relativeProgressCurrentLap = 0.99
	}

	return float32(currentLap) + relativeProgressCurrentLap, nil
}

func (s *Stats) SetManualSetRaceDuration(duration time.Duration) {
	s.ManualSetRaceDuration = duration
}

func (s *Stats) SetRaceStartTime(startTime time.Time) {
	s.raceStartTime = startTime

}

func (s *Stats) GetDurationSinceStart() (time.Duration, error) {
	if s.raceStartTime.IsZero() {
		return time.Duration(0), fmt.Errorf("race start time is not detected, cannot get time since start")
	}
	return s.clock.Now().Sub(s.raceStartTime), nil
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

	durationSinceStart, err := s.GetDurationSinceStart()
	if err != nil {
		return -1, fmt.Errorf("error getting duration since start: %v", err)
	}
	fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
		durationSinceStart,
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
