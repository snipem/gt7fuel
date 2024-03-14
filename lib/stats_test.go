package lib

import (
	"github.com/jmhodges/clock"
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
	avg, _ := gt7stats.GetFuelConsumptionPerMinute()

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
		averageLapTime, err := gt7stats.GetAverageLapTime()
		assert.NoError(t, err)
		assert.Equal(t, float64(90), averageLapTime.Seconds())
	})

	t.Run("GetAverageLapTimeWithNoLaps", func(t *testing.T) {

		gt7stats := NewStats()
		gt7stats.Laps = []Lap{
			{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
			{FuelConsumed: -20, Number: 1, Duration: 90 * time.Second},
			{FuelConsumed: -40, Number: 2, Duration: 90 * time.Second},
			{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
		}
		averageLapTime, err := gt7stats.GetAverageLapTime()
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), averageLapTime)
	})
}

func TestStats_getLapsLeftInRace(t *testing.T) {

	t.Run("Is zero", func(t *testing.T) {
		gt7stats := NewStats()
		lapsLeftInRace, err := gt7stats.getLapsLeftInRace()
		assert.Error(t, err)
		assert.Equal(t, int16(-1), lapsLeftInRace)
	})

	t.Run("Total Laps is set", func(t *testing.T) {
		gt7stats := NewStats()
		gt7stats.LastData.TotalLaps = 10
		gt7stats.LastData.CurrentLap = 5

		lapsLeftInRace, err := gt7stats.getLapsLeftInRace()
		assert.NoError(t, err)
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
		s.Laps = []Lap{
			{
				FuelStart:    100,
				FuelEnd:      80,
				FuelConsumed: 20,
				Number:       0,
				Duration:     0,
			},
			{
				FuelStart:    80,
				FuelEnd:      60,
				FuelConsumed: 20,
				Number:       0,
				Duration:     0,
			},
		}
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.True(t, s.getValidState())
	})
}

func TestStats_getEndOfRaceType(t *testing.T) {

	t.Run("Is not valid", func(t *testing.T) {
		s := NewStats()
		s.LastData.TotalLaps = 10
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.Equal(t, ByLaps, s.getEndOfRaceType())
	})

	t.Run("Is not valid", func(t *testing.T) {
		s := NewStats()
		s.raceStartTime = time.Now().Add(time.Duration(-10) * time.Minute)
		assert.Equal(t, ByTime, s.getEndOfRaceType())
	})
}

func TestStats_GetMessage(t *testing.T) {
	t.Run("No start yet", func(t *testing.T) {
		s := NewStats()
		s.setClock(clock.NewFake())
		assert.Equal(t, Message{
			Speed:                    "0",
			PackageID:                0,
			FuelLeft:                 "0.00",
			FuelConsumptionLastLap:   "-1.00",
			TimeSinceStart:           NoStartDetected,
			FuelNeededToFinishRace:   -1,
			FuelConsumptionAvg:       "NaN",
			FuelDiv:                  "-1",
			RaceTimeInMinutes:        0,
			ValidState:               false,
			LapsLeftInRace:           -1,
			EndOfRaceType:            "By Time",
			FuelConsumptionPerMinute: "-1.00",
			ErrorMessage:             "Laps left in race unknown: BestLap is 0, impossible to calculate laps left based on lap time\nFuel needed to finish race unknown: BestLap or LastLap is 0, impossible to calculate fuel needed to finish race\nFuel Div unknown: error getting fuel needed to finish race: BestLap or LastLap is 0, impossible to calculate fuel needed to finish race",
		}, s.GetMessage())
	})

	t.Run("Start 10 Minutes ago, 30 Minutes Race", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.CarSpeed = 100
		s.LastData.PackageID = 4711
		s.LastData.BestLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.LastLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.CurrentFuel = 20

		s.SetManualSetRaceDuration(30 * time.Minute)
		s.SetRaceStartTime(fakeClock.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = getReasonableLaps()

		assert.Equal(t, Message{
			Speed:                    "100",
			PackageID:                4711,
			FuelLeft:                 "20.00",
			FuelConsumptionLastLap:   "25.00",
			TimeSinceStart:           "10:00.500",
			FuelNeededToFinishRace:   192,
			FuelConsumptionAvg:       "25.00",
			FuelDiv:                  "172",
			RaceTimeInMinutes:        30,
			ValidState:               true,
			LapsLeftInRace:           7,
			EndOfRaceType:            "By Time",
			FuelConsumptionPerMinute: "16.67",
		}, s.GetMessage())
	})

	t.Run("Start 10 Minutes ago with 10 Laps in Total", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.CarSpeed = 100
		s.LastData.PackageID = 4711
		s.LastData.TotalLaps = 10
		s.LastData.BestLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.LastLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.CurrentFuel = 20

		s.SetRaceStartTime(fakeClock.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = getReasonableLaps()

		assert.Equal(t, Message{
			Speed:                    "100",
			PackageID:                4711,
			FuelLeft:                 "20.00",
			FuelConsumptionLastLap:   "25.00",
			TimeSinceStart:           "10:00.500",
			FuelNeededToFinishRace:   192,
			FuelConsumptionAvg:       "25.00",
			FuelDiv:                  "172",
			RaceTimeInMinutes:        30, // total laps * best lap
			ValidState:               true,
			LapsLeftInRace:           11,
			EndOfRaceType:            ByLaps,
			FuelConsumptionPerMinute: "16.67",
		}, s.GetMessage())
	})

}

func getReasonableLaps() []Lap {
	return []Lap{
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
}
