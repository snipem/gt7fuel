package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_getDurationFromGT7Time(t *testing.T) {

	duration := getDurationFromGT7Time(int32(84149))
	expectedDuration := 1*time.Minute +
		24*time.Second +
		149*time.Millisecond

	assert.Equal(t,
		expectedDuration,
		duration)

}

func Test_calculateFuelNeededToFinishRace(t *testing.T) {

	t.Run("calculateFuelNeededToFinishRace", func(t *testing.T) {

		fuelneededToFinish := calculateFuelNeededToFinishRace(
			20*time.Minute+30*time.Second+10*time.Millisecond,
			60*time.Minute,
			1*time.Minute+30*time.Second,
			1*time.Minute+45*time.Second,
			2,
		)
		assert.Equal(t, float32(46.856953), fuelneededToFinish)
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
			1*time.Minute,
			2,
		)
		assert.Equal(t, float32(82), fuelneededToFinish)
	})
}

func Test_getSportFormat(t *testing.T) {
	t.Run("getSportFormat", func(t *testing.T) {
		formattedTime := getSportFormat(1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "01:30.010", formattedTime)
	})

	t.Run("getSportFormatManyMilliseconds", func(t *testing.T) {
		formattedTime := getSportFormat(1*time.Minute + 30*time.Second + 1010*time.Millisecond)
		assert.Equal(t, "01:31.010", formattedTime)
	})

	t.Run("getSportFormatWithHours", func(t *testing.T) {
		formattedTime := getSportFormat(2*time.Hour + 1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "121:30.010", formattedTime)
	})
}

func Test_getAverageFuelConsumption(t *testing.T) {
	avg := getAverageFuelConsumption(
		[]Lap{
			{
				FuelConsumed: -20,
				Number:       0,
			},
			{
				FuelConsumed: 20,
				Number:       1,
			},
			{
				FuelConsumed: 40,
				Number:       2,
			},
			{
				FuelConsumed: -200,
				Number:       3,
			},
		})

	assert.Equal(t, float32(30), avg)
}
