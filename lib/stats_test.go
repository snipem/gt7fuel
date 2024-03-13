package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_getAverageFuelConsumption(t *testing.T) {

	gt7stats := NewStats()
	gt7stats.Laps = []Lap{
		{FuelConsumed: -20, Number: 0},
		{FuelConsumed: 20, Number: 1},
		{FuelConsumed: 40, Number: 2},
		{FuelConsumed: -200, Number: 3},
	}
	avg := gt7stats.GetAverageFuelConsumption()

	assert.Equal(t, float32(30), avg)
}

func Test_getAverageFuelConsumptionPerMinute(t *testing.T) {

	gt7stats := NewStats()
	gt7stats.Laps = []Lap{
		{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
		{FuelConsumed: 20, Number: 1, Duration: 90 * time.Second}, // valid for calculation
		{FuelConsumed: 40, Number: 2, Duration: 90 * time.Second}, // valid for calculation
		{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
	}
	assert.Len(t, getAccountableLaps(gt7stats.Laps), 2)
	avg := gt7stats.GetFuelConsumptionPerMinute()

	// 30 avg fuel per lap / 1,5 avg duration = 20 fuel per minute
	assert.Equal(t, float32(20), avg)
}

func TestStats_GetAverageLapTime(t *testing.T) {

	t.Run("GetAverageLapTime", func(t *testing.T) {

		gt7stats := NewStats()
		gt7stats.Laps = []Lap{
			{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
			{FuelConsumed: 20, Number: 1, Duration: 90 * time.Second}, // valid for calculation
			{FuelConsumed: 40, Number: 2, Duration: 90 * time.Second}, // valid for calculation
			{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
		}
		assert.Equal(t, float64(90), gt7stats.GetAverageLapTime().Seconds())
	})

	t.Run("GetAverageLapTimeWithNoLaps", func(t *testing.T) {

		gt7stats := NewStats()
		gt7stats.Laps = []Lap{
			{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
			{FuelConsumed: -20, Number: 1, Duration: 90 * time.Second},
			{FuelConsumed: -40, Number: 2, Duration: 90 * time.Second},
			{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
		}
		assert.Equal(t, float64(0), gt7stats.GetAverageLapTime().Seconds())
	})
}

func TestStats_getLapsLeftInRace(t *testing.T) {

	t.Run("Is zero", func(t *testing.T) {
		gt7stats := NewStats()
		lapsLeftInRace := gt7stats.getLapsLeftInRace()
		assert.Equal(t, int16(0), lapsLeftInRace)
	})

	t.Run("Total Laps is set", func(t *testing.T) {
		gt7stats := NewStats()
		gt7stats.LastData.TotalLaps = 10
		gt7stats.LastData.CurrentLap = 5

		lapsLeftInRace := gt7stats.getLapsLeftInRace()
		assert.Equal(t, int16(6), lapsLeftInRace)
	})

}

func TestStats_getValidState(t *testing.T) {
	t.Run("Is not valid", func(t *testing.T) {
		s := NewStats()
		assert.False(t, s.getValidState())
	})

	t.Run("Is valid", func(t *testing.T) {
		s := NewStats()
		s.fuelConsumptionLastLap = 32
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.True(t, s.getValidState())
	})
}

func TestStats_getEndOfRaceType(t *testing.T) {

	t.Run("Is not valid", func(t *testing.T) {
		s := NewStats()
		s.LastData.TotalLaps = 10
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.Equal(t, BY_LAPS, s.getEndOfRaceType())
	})

	t.Run("Is not valid", func(t *testing.T) {
		s := NewStats()
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.Equal(t, BY_TIME, s.getEndOfRaceType())
	})
}

func TestStats_GetMessage(t *testing.T) {
	t.Run("No start yet", func(t *testing.T) {
		s := NewStats()
		assert.Equal(t, Message{
			Speed:                    "0",
			PackageID:                0,
			FuelLeft:                 "0.00",
			FuelConsumptionLastLap:   "0.00",
			TimeSinceStart:           NO_START_DETECTED,
			FuelNeededToFinishRace:   0,
			FuelConsumptionAvg:       "NaN",
			FuelDiv:                  "NaN",
			RaceTimeInMinutes:        0,
			ValidState:               false,
			LapsLeftInRace:           0,
			EndOfRaceType:            "By Time",
			FuelConsumptionPerMinute: "NaN",
		}, s.GetMessage())
	})

	t.Run("Start 10 Minutes ago, 60 Minutes Race", func(t *testing.T) {
		s := NewStats()
		s.LastData.CarSpeed = 100
		s.LastData.PackageID = 4711

		s.SetManualSetRaceDuration(30)
		s.SetRaceStartTime(time.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = []Lap{
			{
				FuelStart:    100,
				FuelEnd:      50,
				FuelConsumed: 50,
				Number:       0,
				Duration:     1*time.Minute + 31*time.Second,
			},
			{
				FuelStart:    50,
				FuelEnd:      25,
				FuelConsumed: 25,
				Number:       1,
				Duration:     1*time.Minute + 30*time.Second,
			},
		}

		assert.Equal(t, Message{
			Speed:                    "100",
			PackageID:                4711,
			FuelLeft:                 "0.00",
			FuelConsumptionLastLap:   "0.00",
			TimeSinceStart:           "10:00.500",
			FuelNeededToFinishRace:   0,
			FuelConsumptionAvg:       "25.00",
			FuelDiv:                  "NaN",
			RaceTimeInMinutes:        0,
			ValidState:               false,
			LapsLeftInRace:           0,
			EndOfRaceType:            "By Time",
			FuelConsumptionPerMinute: "16.67",
		}, s.GetMessage())
	})

	t.Run("Start 10 Minutes ago with 10 Laps in Total", func(t *testing.T) {
		s := NewStats()
		s.LastData.CarSpeed = 100
		s.LastData.PackageID = 4711
		s.LastData.TotalLaps = 10

		s.SetRaceStartTime(time.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = []Lap{
			{
				FuelStart:    100,
				FuelEnd:      50,
				FuelConsumed: 50,
				Number:       0,
				Duration:     1*time.Minute + 31*time.Second,
			},
			{
				FuelStart:    50,
				FuelEnd:      25,
				FuelConsumed: 25,
				Number:       1,
				Duration:     1*time.Minute + 30*time.Second,
			},
		}

		assert.Equal(t, Message{
			Speed:                    "100",
			PackageID:                4711,
			FuelLeft:                 "0.00",
			FuelConsumptionLastLap:   "0.00",
			TimeSinceStart:           "10:00.500",
			FuelNeededToFinishRace:   0,
			FuelConsumptionAvg:       "25.00",
			FuelDiv:                  "NaN",
			RaceTimeInMinutes:        0,
			ValidState:               false,
			LapsLeftInRace:           11,
			EndOfRaceType:            BY_LAPS,
			FuelConsumptionPerMinute: "16.67",
		}, s.GetMessage())
	})

}
