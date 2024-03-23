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
	avg, err := gt7stats.GetAverageFuelConsumptionPerLap()
	assert.NoError(t, err)

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

	//t.Run("GetAverageLapTimeIgnoreBoxLaps", func(t *testing.T) {
	//
	//	gt7stats := NewStats()
	//	gt7stats.Laps = []Lap{
	//		{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
	//		{FuelConsumed: 20, Number: 1, Duration: 90 * time.Second},  // This is a box lap
	//		{FuelConsumed: 20, Number: 1, Duration: 90 * time.Second},  // valid for calculation
	//		{FuelConsumed: -40, Number: 2, Duration: 90 * time.Second}, // This is the lap going into the box
	//		{FuelConsumed: 40, Number: 2, Duration: 90 * time.Second},  // This is a box lap
	//		{FuelConsumed: 40, Number: 2, Duration: 90 * time.Second},  // valid for calculation
	//		{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
	//	}
	//	averageLapTime, err := gt7stats.GetAverageLapTime()
	//	assert.NoError(t, err)
	//	assert.Equal(t, float64(90), averageLapTime.Seconds())
	//})
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

	t.Run("Max time is set, but no race start set", func(t *testing.T) {
		gt7stats := NewStats()
		gt7stats.LastData.CurrentLap = 5
		gt7stats.LastData.BestLap = 1 * 60 * 1000
		gt7stats.SetManualSetRaceDuration(10 * time.Minute)

		lapsLeftInRace, err := gt7stats.getLapsLeftInRace()
		assert.Error(t, err)
		assert.Equal(t, int16(-1), lapsLeftInRace)
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
			Speed:                      "0",
			PackageID:                  0,
			FuelLeft:                   "0.00",
			FuelConsumptionLastLap:     "-1.00",
			TimeSinceStart:             NoStartDetected,
			FuelNeededToFinishRace:     -1,
			FuelConsumptionAvg:         "-1.00",
			FuelDiv:                    "-1",
			RaceTimeInMinutes:          0,
			ValidState:                 false,
			LapsLeftInRace:             -1,
			EndOfRaceType:              "By Time",
			FuelConsumptionPerMinute:   "-1.00",
			ErrorMessage:               "Laps left in race unknown: error getting duration since start: race start time is not detected, cannot get time since start\nFuel needed to finish race unknown: error getting fuel consumption last lap: not enough laps to return fuel consumption of last lap, nr of laps: 0\nFuel Div unknown: error getting fuel needed to finish race: error getting fuel consumption last lap: not enough laps to return fuel consumption of last lap, nr of laps: 0",
			NextPitStop:                -1,
			CurrentLapProgressAdjusted: "-1.0",
		}, s.GetMessage())
	})

	t.Run("Start 10 Minutes ago, 30 Minutes Race", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.CarSpeed = 100
		s.LastData.CurrentLap = 5
		s.LastData.PackageID = 4711
		s.LastData.BestLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.LastLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.CurrentFuel = 20

		s.SetManualSetRaceDuration(30 * time.Minute)
		s.SetRaceStartTime(fakeClock.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = getReasonableLaps()
		s.OngoingLap = getReasonableOngoingLap()

		assert.Equal(t, Message{
			Speed:                      "100",
			PackageID:                  4711,
			FuelLeft:                   "20.00",
			FuelConsumptionLastLap:     "25.00",
			TimeSinceStart:             "10:00.500",
			FuelNeededToFinishRace:     192,
			FuelConsumptionAvg:         "25.00",
			FuelDiv:                    "172",
			RaceTimeInMinutes:          30,
			ValidState:                 true,
			LapsLeftInRace:             7,
			EndOfRaceType:              "By Time",
			FuelConsumptionPerMinute:   "16.67",
			NextPitStop:                5,
			CurrentLapProgressAdjusted: "5.3",
			ErrorMessage:               "",
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
		s.LastData.CurrentLap = 5

		s.SetRaceStartTime(fakeClock.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = getReasonableLaps()
		s.OngoingLap = getReasonableOngoingLap()

		assert.Equal(t, Message{
			Speed:                      "100",
			PackageID:                  4711,
			FuelLeft:                   "20.00",
			FuelConsumptionLastLap:     "25.00",
			TimeSinceStart:             "10:00.500",
			FuelNeededToFinishRace:     192,
			FuelConsumptionAvg:         "25.00",
			FuelDiv:                    "172",
			RaceTimeInMinutes:          30, // total laps * best lap
			ValidState:                 true,
			LapsLeftInRace:             6,
			EndOfRaceType:              ByLaps,
			FuelConsumptionPerMinute:   "16.67",
			NextPitStop:                5,
			CurrentLapProgressAdjusted: "5.3",
		}, s.GetMessage())
	})

	t.Run("No Fuel consumption", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.CarSpeed = 100
		s.LastData.PackageID = 4711
		s.LastData.BestLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.LastLap = 3 * 60 * 1000 // 3 minutes
		s.LastData.CurrentFuel = 100

		s.SetManualSetRaceDuration(30 * time.Minute)
		s.SetRaceStartTime(fakeClock.Now().Add(time.Duration(-10)*time.Minute + time.Duration(-500)*time.Millisecond))

		s.Laps = []Lap{
			{FuelStart: 100, FuelEnd: 100, FuelConsumed: 0, Number: 0, Duration: time.Minute * 1},
			{FuelStart: 100, FuelEnd: 100, FuelConsumed: 0, Number: 1, Duration: time.Minute * 1},
			{FuelStart: 100, FuelEnd: 100, FuelConsumed: 0, Number: 2, Duration: time.Minute * 1},
			{FuelStart: 100, FuelEnd: 100, FuelConsumed: 0, Number: 3, Duration: time.Minute * 1},
		}
		s.OngoingLap = getReasonableOngoingLap()

		assert.Equal(t, Message{
			Speed:                      "100",
			PackageID:                  4711,
			FuelLeft:                   "100.00",
			FuelConsumptionLastLap:     "0.00",
			TimeSinceStart:             "10:00.500",
			FuelNeededToFinishRace:     0,
			FuelConsumptionAvg:         "0.00",
			FuelDiv:                    "-100",
			RaceTimeInMinutes:          30,
			ValidState:                 true,
			LapsLeftInRace:             7,
			EndOfRaceType:              "By Time",
			FuelConsumptionPerMinute:   "0.00",
			ErrorMessage:               "",
			NextPitStop:                -1,
			CurrentLapProgressAdjusted: "0.3",
		}, s.GetMessage())
	})

}

func getReasonableLaps() []Lap {

	durationFirstLap := 1*time.Minute + 31*time.Second
	durationSecondLap := 1*time.Minute + 30*time.Second

	startOfOngoingLap := getReasonableOngoingLap().LapStart

	return []Lap{
		{
			FuelStart:    100,
			FuelEnd:      50,
			FuelConsumed: 50,
			Number:       0,
			Duration:     durationFirstLap,
			LapStart:     startOfOngoingLap.Add(-(durationFirstLap + durationSecondLap)),
		},
		{
			FuelStart:    50,
			FuelEnd:      25,
			FuelConsumed: 25,
			Number:       1,
			Duration:     durationSecondLap,
			LapStart:     startOfOngoingLap.Add(-(durationSecondLap)),
		},
	}
}

func getReasonableOngoingLap() Lap {
	return Lap{
		FuelStart: 25,
		Number:    2,
		LapStart:  clock.NewFake().Now().Add(-(1 * time.Minute)),
	}
}

func TestStats_getTotalLapsInRace(t *testing.T) {

	t.Run("Total Laps", func(t *testing.T) {
		s := NewStats()
		s.LastData.TotalLaps = 10

		totalLaps, err := s.getTotalLapsInRace()

		assert.NoError(t, err)
		assert.Equal(t, int16(10), totalLaps)
	})

	t.Run("Total Laps By Duration", func(t *testing.T) {
		s := NewStats()
		s.LastData.TotalLaps = 0
		s.LastData.BestLap = 2 * 60 * 1000 // 2 minutes
		s.ManualSetRaceDuration = 30 * time.Minute

		totalLaps, err := s.getTotalLapsInRace()

		assert.NoError(t, err)
		assert.Equal(t, int16(16), totalLaps)
	})
}

func TestStats_GetProgressAdjustedCurrentLap(t *testing.T) {
	t.Run("Total Laps", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.TotalLaps = 10
		s.LastData.CurrentLap = 2
		s.LastData.BestLap = 2 * 60 * 1000 // 2 minutes
		s.OngoingLap = getReasonableOngoingLap()
		s.Laps = getReasonableLaps()

		currentProgress, err := s.GetProgressAdjustedCurrentLap()

		assert.NoError(t, err)
		assert.Equal(t, float32(2.5), currentProgress)
	})
	t.Run("Total Time", func(t *testing.T) {
		s := NewStats()
		fakeClock := clock.NewFake()
		s.setClock(fakeClock)
		s.LastData.TotalLaps = 10
		s.LastData.CurrentLap = 2
		s.LastData.BestLap = 2 * 60 * 1000 // 2 minutes
		s.OngoingLap = getReasonableOngoingLap()
		s.OngoingLap.LapStart = clock.NewFake().Now().Add(-(10 * time.Minute)) // 10 minutes and going in current lap
		s.Laps = getReasonableLaps()

		currentProgress, err := s.GetProgressAdjustedCurrentLap()

		assert.NoError(t, err)
		assert.Less(t, currentProgress, float32(3)) // Should still be lower than 3
	})
}

func TestStats_GetNextNecessaryPitStopInLap(t *testing.T) {

	s := NewStats()
	fakeClock := clock.NewFake()
	s.setClock(fakeClock)
	s.Laps = getReasonableLaps()
	s.OngoingLap = getReasonableOngoingLap()
	s.LastData.TotalLaps = 10
	s.LastData.CurrentLap = 2
	s.LastData.BestLap = 2 * 60 * 1000 // 2 minutes

	lap, err := s.GetNextNecessaryPitStopAtEndOfLap()
	assert.NoError(t, err)
	assert.Equal(t, int(2), lap)

}

func Test_getDurationInLap(t *testing.T) {
	t.Run("Total Laps", func(t *testing.T) {
		lap, err := getDurationInLap(time.Duration(3)*time.Minute, time.Duration(4)*time.Minute, 0)
		assert.NoError(t, err)
		assert.Equal(t, float32(0.74999994), lap)
	})

	t.Run("Total Laps - Slow Lap", func(t *testing.T) {
		lap, err := getDurationInLap(time.Duration(5)*time.Minute, time.Duration(4)*time.Minute, 0)
		assert.NoError(t, err)
		assert.Less(t, lap, float32(1))
	})
}

func Test_getNextPitStop(t *testing.T) {
	// now: 3.5l at Lap 5 + 0.5, with 3l consumption per lap should pit at end of lap 5
	assert.Equal(t, 5, getNextPitStop(3.5, 3, 5.5), "Simple")
	// now: 3.5l at Lap 5 close to finish, with 3l consumption per lap should pit at end of lap 6
	assert.Equal(t, 6, getNextPitStop(3.5, 3, 5.9), "Simple")

	assert.Equal(t, 0, getNextPitStop(0, 5, 0.5), "Simple")
	assert.Equal(t, 25, getNextPitStop(100, 5, 5.5), "Simple")
	assert.Equal(t, -1, getNextPitStop(100, 0, 5.5), "Simple")
}
