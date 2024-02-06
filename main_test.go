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

func Test_getLapsLeftInRace(t *testing.T) {
	t.Run("getLapsLeftInRace", func(t *testing.T) {

		lapsLeftInRace := getLapsLeftInRace(1*time.Minute+30*time.Second+10*time.Millisecond, 60*time.Minute, 1*time.Minute+45*time.Second)
		assert.Equal(t, int32(34), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceSimple", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(0, 100*time.Minute, 1*time.Minute)
		// 100 laps by lap time and 1 additional
		assert.Equal(t, int32(100)+int32(1), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceCheckAddedLaps", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(100*time.Minute+30*time.Second, 100*time.Minute, 1*time.Minute)
		// Max 1 lap left
		assert.Equal(t, int32(1), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceEndOfRace", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(101*time.Minute, 100*time.Minute, 1*time.Minute)
		// No lap left
		assert.Equal(t, int32(0), lapsLeftInRace)
	})
}

func Test_getMedianFuelConsumption(t *testing.T) {
	assert.Equal(t, float32(5), getMedianFuelConsumption([]Lap{
		{Number: 1, FuelConsumed: float32(5)},
		{Number: 2, FuelConsumed: float32(5)},
	}))

	assert.Equal(t, float32(5), getMedianFuelConsumption([]Lap{
		{Number: 1, FuelConsumed: float32(50)},
		{Number: 2, FuelConsumed: float32(51)},
		{Number: 3, FuelConsumed: float32(5)},
	}))

	assert.Equal(t, float32(5), getMedianFuelConsumption([]Lap{
		{Number: 0, FuelConsumed: float32(50)},
		{Number: 1, FuelConsumed: float32(5)},
		{Number: 2, FuelConsumed: float32(5)},
	}))
}
