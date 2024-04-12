package lib

import (
	"fmt"
	"github.com/jmhodges/clock"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib/experimental"
	"math"
	"strings"
	"time"
)

type Stats struct {
	LastLoggedData gt7.GTData
	Laps           []Lap
	OngoingLap     Lap
	LastData       *gt7.GTData
	LastTireData   *experimental.TireData
	// ManualSetRaceDuration is the race duration manually set by the user if it is not
	// transmitted over telemetry
	ManualSetRaceDuration time.Duration
	raceStartTime         time.Time
	clock                 clock.Clock
	ConnectionActive      bool
}

func (s *Stats) GetFuelConsumptionLastLap() (float32, error) {
	return getFuelConsumptionLastLap(s.Laps)
}

func getFuelConsumptionLastLap(laps []Lap) (float32, error) {
	if len(laps) < 1 {
		return -1, fmt.Errorf("not enough laps to return fuel consumption of last lap, nr of laps: %d", len(laps))
	}

	lastLap := laps[len(laps)-1]
	// if last lap was a pit lap, take the lap before
	if lastLap.GetFuelConsumed() < 0 {
		return lastLap.PreviousLap.GetFuelConsumed(), nil
	}

	return lastLap.GetFuelConsumed(), nil

}

func NewStats() *Stats {
	s := Stats{}
	s.LastLoggedData = gt7.GTData{}
	s.LastData = &gt7.GTData{}
	s.LastTireData = &experimental.TireData{}
	s.ConnectionActive = false
	// set a proper clock
	s.setClock(clock.New())
	return &s
}

type Lap struct {
	FuelStart    float32
	FuelEnd      float32
	TireConsumed float32
	Number       int16
	Duration     time.Duration
	LapStart     time.Time
	PreviousLap  *Lap
	TiresEnd     experimental.TireData
	TiresStart   experimental.TireData
}

func (l Lap) String() string {
	return fmt.Sprintf("Lap %d: FuelStart=%.2f, FuelEnd=%.2f, FuelConsumed=%.2f, Duration=%s",
		l.Number, l.FuelStart, l.FuelEnd, l.GetFuelConsumed(), l.Duration)
}

func (l *Lap) GetTotalRaceDurationAtStartOfLap() time.Duration {
	return l.GetTotalRaceDurationAtEndOfLap() - l.Duration
}
func (l *Lap) GetTotalRaceDurationAtEndOfLap() time.Duration {
	if l.PreviousLap != nil {
		return l.Duration + l.PreviousLap.GetTotalRaceDurationAtEndOfLap()
	} else {
		return l.Duration
	}
}

func (l Lap) GetFuelConsumed() float32 {
	return l.FuelStart - l.FuelEnd
}

func (l Lap) TireConsumptionOnAllTires() experimental.TireDelta {
	td := experimental.TireDelta{FrontLeft: l.TiresStart.FrontLeft - l.TiresEnd.FrontLeft, FrontRight: l.TiresStart.FrontRight - l.TiresEnd.FrontRight, RearLeft: l.TiresStart.RearLeft - l.TiresEnd.RearLeft, RearRight: l.TiresStart.RearRight - l.TiresEnd.RearRight}
	return td
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

func (s *Stats) GetMessage() Message {

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

	fuelConsumptionLastLap, err := s.GetFuelConsumptionLastLap()
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

	raceduration, err := s.getRaceDuration()
	if err != nil {
		errorMessages = append(errorMessages, fmt.Sprintf("Raceduration unknown: %v", err))
		isValid = false
	}

	formattedLaps := getHtmlTableForLaps(s.Laps)

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
		RaceTimeInMinutes:          int32(raceduration.Minutes()),
		ValidState:                 isValid,
		LowestTireTemp:             float32(minTemp),
		ErrorMessage:               errorMessage,
		NextPitStop:                int16(nextPitStop),
		CurrentLapProgressAdjusted: fmt.Sprintf("%.1f", currentLapProgressAdjusted),
		FormattedLaps:              formattedLaps,
		Tires:                      fmt.Sprintf("Vorne: %d%%, %d%% Hinten: %d%%, %d%%", s.LastTireData.FrontLeft, s.LastTireData.FrontRight, s.LastTireData.RearLeft, s.LastTireData.RearRight),
	}
	return message

}

func getHtmlTableForLaps(laps []Lap) string {

	html := "<table>\n"

	// Header
	html += "\t<tr>\n" +
		fmt.Sprintf("\t\t<th>#</th>\n") +
		fmt.Sprintf("\t\t<th>Duration</th>\n") +
		fmt.Sprintf("\t\t<th>Time</th>\n") +
		fmt.Sprintf("\t\t<th>Fuel</th>\n") +
		fmt.Sprintf("\t\t<th>Tires</th>\n") +
		"\t</tr>\n"

	for i := len(laps) - 1; i >= 0; i-- {

		lap := laps[i]

		html += "\t<tr>\n" +
			fmt.Sprintf("\t\t<td>%d</td>\n", lap.Number) +
			fmt.Sprintf("\t\t<td>%s</td>\n", GetSportFormat(lap.GetTotalRaceDurationAtEndOfLap())) +
			fmt.Sprintf("\t\t<td>%s</td>\n", GetSportFormat(lap.Duration)) +
			fmt.Sprintf("\t\t<td>%.1f%%</td>\n", lap.GetFuelConsumed()) +
			fmt.Sprintf("\t\t<td>%s</td>\n", lap.TiresEnd.Html()) +
			"\t</tr>\n"
	}
	html += "</table>\n"
	return html
}

func (s *Stats) getValidState() bool {
	validState := true

	durationSinceStart, err := s.GetDurationSinceStart()
	if err != nil {
		validState = false
	}

	if durationSinceStart > 1000*time.Hour {
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

		bestLap, err := s.getReferenceLapDuration()
		if err != nil {
			return -1, fmt.Errorf("ReferenceLap is 0, impossible to calculate laps left based on lap time")
		}

		durationSinceStart, err := s.GetDurationSinceStart()
		if err != nil {
			return -1, fmt.Errorf("error getting duration since start: %v", err)
		}

		raceDuration, err := s.getRaceDuration()
		if err != nil {
			return -1, fmt.Errorf("error getting duration since start: %v", err)
		}

		lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(durationSinceStart, raceDuration, bestLap)
		if err != nil {
			return -1, fmt.Errorf("error getting laps left: %v", err)
		}
		return lapsLeftInRace, nil
	}

}

func (s *Stats) getBestLapDuration() (time.Duration, error) {

	if s.LastData.BestLap < 0 {
		return 0, fmt.Errorf("BestLap is %d, impossible to calculate best lap", s.LastData.BestLap)
	}

	bestLap := GetDurationFromGT7Time(s.LastData.BestLap)
	return bestLap, nil
}

func (s *Stats) getLastLapDuration() (time.Duration, error) {

	if s.LastData.LastLap < 0 {
		return 0, fmt.Errorf("LastLap is %d, impossible to calculate last lap", s.LastData.LastLap)
	}

	lastLap := GetDurationFromGT7Time(s.LastData.LastLap)
	return lastLap, nil
}

func (s *Stats) getReferenceLapDuration() (time.Duration, error) {
	referenceLap, err := s.getBestLapDuration()
	if err != nil {
		referenceLap, err = s.getLastLapDuration()
		if err != nil {
			return -1, fmt.Errorf("error getting reference lap, both BestLap and LastLap are <0: %v", err)
		}
	}
	return referenceLap, nil
}

func (s *Stats) getTotalLapsInRace() (int16, error) {

	if s.LastData.TotalLaps > 0 {
		return s.LastData.TotalLaps, nil
	}

	bestLap, err := s.getReferenceLapDuration()
	if err != nil {
		return -1, fmt.Errorf("BestLap is 0, impossible to calculate total laps based on lap time")
	}

	raceDuration, err := s.getRaceDuration()
	if err != nil {
		return -1, fmt.Errorf("error getting duration since start: %v", err)
	}

	// we assume the race hos not started yet to get the total number of laps

	lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(time.Duration(0), raceDuration, bestLap)
	if err != nil {
		return -1, fmt.Errorf("error getting laps left: %v", err)
	}
	return lapsLeftInRace, nil
}
func (s *Stats) getRaceDuration() (time.Duration, error) {
	referenceLap, err := s.getReferenceLapDuration()
	if err != nil {
		return 0, fmt.Errorf("error getting reference lap: %v", err)
	}
	if s.LastData.TotalLaps > 0 {
		return referenceLap * time.Duration(s.LastData.TotalLaps), nil
	} else {
		// Add a lap to the manual set race duration because there might be an additional lap
		// This is the only place we account for this matter
		return s.ManualSetRaceDuration + referenceLap, nil
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

	bestLap, err := s.getReferenceLapDuration()
	if err != nil {
		return -1, fmt.Errorf("impossible to calculate progress adjusted current lap: %v", err)
	}
	return getDurationInLap(durationInCurrentLap, bestLap, s.LastData.CurrentLap)

}

func (s *Stats) GetProgressAdjustedLapsLeftInRace() (float32, error) {

	totalLapsInRace, err := s.getTotalLapsInRace()
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}

	progressAdjustedCurrentLap, err := s.GetProgressAdjustedCurrentLap()
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}

	return float32(totalLapsInRace) - progressAdjustedCurrentLap, nil
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

// GetFuelNeededToFinishRaceInTotal calculates how much fuel is needed to finish the race in total
// Implicit values needed: fuel consumption last lap, reference lap, race duration, duration since start
func (s *Stats) GetFuelNeededToFinishRaceInTotal() (float32, error) {

	// it is best to use the last lap, since this will compensate for missed packages etc.
	fuelConsumptionLastLap, err := s.GetFuelConsumptionLastLap()
	if err != nil {
		return -1, fmt.Errorf("error getting fuel consumption last lap: %v", err)
	}

	referenceLap, err := s.getReferenceLapDuration()
	if err != nil {
		return -1, fmt.Errorf("error getting reference lap: %v", err)
	}

	raceDuration, err := s.getRaceDuration()
	if err != nil {
		return -1, fmt.Errorf("error getting race duration: %v", err)
	}

	durationSinceStart, err := s.GetDurationSinceStart()
	if err != nil {
		return -1, fmt.Errorf("error getting duration since start: %v", err)
	}
	fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
		durationSinceStart,
		raceDuration,
		referenceLap,
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
