package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRoundUpAlways(t *testing.T) {
	assert.Equal(t, int32(3), RoundUpAlways(2.4))
	assert.Equal(t, int32(3), RoundUpAlways(2.6))
	assert.Equal(t, int32(2), RoundUpAlways(2.0))
	assert.Equal(t, int32(3), RoundUpAlways(2.01))
	// higher is desired
	assert.Equal(t, int32(-2), RoundUpAlways(-2.01))
}
func Test_getDurationFromGT7Time(t *testing.T) {

	duration := GetDurationFromGT7Time(int32(84149))
	expectedDuration := 1*time.Minute +
		24*time.Second +
		149*time.Millisecond

	assert.Equal(t,
		expectedDuration,
		duration)

}
func Test_getSportFormat(t *testing.T) {
	t.Run("getSportFormat", func(t *testing.T) {
		formattedTime := GetSportFormat(1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "01:30.010", formattedTime)
	})

	t.Run("getSportFormatManyMilliseconds", func(t *testing.T) {
		formattedTime := GetSportFormat(1*time.Minute + 30*time.Second + 1010*time.Millisecond)
		assert.Equal(t, "01:31.010", formattedTime)
	})

	t.Run("getSportFormatWithHours", func(t *testing.T) {
		formattedTime := GetSportFormat(2*time.Hour + 1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "121:30.010", formattedTime)
	})
}
func Test_calculateFuelNeededToFinishRace(t *testing.T) {

	t.Run("calculateFuelNeededToFinishRace", func(t *testing.T) {

		fuelneededToFinish := calculateFuelNeededToFinishRace(
			20*time.Minute+30*time.Second+10*time.Millisecond,
			60*time.Minute,
			1*time.Minute+30*time.Second,
			2,
		)
		// FIME calculate
		assert.Equal(t, float32(52.666447), fuelneededToFinish)
	})

	t.Run("calculateFuelNeededToFinishRaceSimple", func(t *testing.T) {

		// 20 minutes in the race
		// 60 minutes in total + 1 minute added (extra lap)
		// = 41 minutes to go
		// 1 minute per lap
		// 2 fuel consumed per lap
		// = 2 fuel per minute / lap

		// fuel needed: 41 * 2 = 82

		fuelneededToFinish := calculateFuelNeededToFinishRace(
			20*time.Minute,
			60*time.Minute,
			1*time.Minute,
			2,
		)
		assert.Equal(t, float32(80), fuelneededToFinish)
	})
}

func Test_getLapsLeftInRace(t *testing.T) {
	t.Run("getLapsLeftInRace", func(t *testing.T) {

		lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(1*time.Minute+30*time.Second+10*time.Millisecond, 60*time.Minute, 1*time.Minute+45*time.Second)
		assert.NoError(t, err)
		assert.Equal(t, int16(33), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceSimple", func(t *testing.T) {
		lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(0, 100*time.Minute, 1*time.Minute)
		assert.NoError(t, err)
		// 100 laps by lap time and 1 additional
		assert.Equal(t, int16(100), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceSimpleOngoing", func(t *testing.T) {
		lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(20*time.Minute, 50*time.Minute, 2*time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int16(15), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRace30SecondsToGo", func(t *testing.T) {
		lapsLeftInRace, err := GetLapsLeftInRaceBasedOnTotalRaceDuration(99*time.Minute+30*time.Second, 100*time.Minute, 1*time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int16(0), lapsLeftInRace)
	})

	//t.Run("getLapsLeftInRaceNoTimeLeft", func(t *testing.T) {
	//	timeInRace := 100 * time.Minute
	//	lapsLeftInRace := getLapsLeftInRace(timeInRace, timeInRace, 1*time.Minute)
	//	assert.Equal(t, int16(0), lapsLeftInRace)
	//})

	t.Run("getLapsLeftInRaceCheckAddedLaps", func(t *testing.T) {
		lapsLeftInRace, _ := GetLapsLeftInRaceBasedOnTotalRaceDuration(100*time.Minute+30*time.Second, 100*time.Minute, 1*time.Minute)
		// In last lap
		assert.Equal(t, int16(0), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceEndOfRace", func(t *testing.T) {
		lapsLeftInRace, _ := GetLapsLeftInRaceBasedOnTotalRaceDuration(101*time.Minute, 100*time.Minute, 1*time.Minute)
		// No lap left
		assert.Equal(t, int16(0), lapsLeftInRace)
	})
}

func Test_getAccountableLaps(t *testing.T) {
	accountableLaps := getAccountableLaps([]Lap{
		{Number: 0, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 1, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 2, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 3, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 4, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 5, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 6, FuelStart: 5, FuelEnd: 10, Duration: 0},
		{Number: 7, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 8, FuelStart: 10, FuelEnd: 5, Duration: 0},
		{Number: 9, FuelStart: 10, FuelEnd: 5, Duration: 0},
	})
	assert.Len(t, accountableLaps, 8) // first and sixth lap is not accountable

	assert.NotContains(t, accountableLaps, Lap{Number: 6, FuelStart: 5, FuelEnd: 10, Duration: 0})
	assert.NotContains(t, accountableLaps, Lap{Number: 0, FuelStart: 10, FuelEnd: 5, Duration: 0})
	assert.Contains(t, accountableLaps, Lap{Number: 1, FuelStart: 10, FuelEnd: 5, Duration: 0})
}

func TestGetSportFormat(t *testing.T) {
	assert.Equal(t, "01:00:00.000", GetSportFormat(time.Duration(time.Hour*1)))
	assert.Equal(t, "10:00.000", GetSportFormat(time.Duration(time.Minute*10)))
	assert.Equal(t, "-10:10.000", GetSportFormat(time.Duration(-time.Minute*10-time.Second*10)))
}
