package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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

func Test_getLapsLeftInRace(t *testing.T) {
	t.Run("getLapsLeftInRace", func(t *testing.T) {

		lapsLeftInRace := getLapsLeftInRace(1*time.Minute+30*time.Second+10*time.Millisecond, 60*time.Minute, 1*time.Minute+45*time.Second)
		assert.Equal(t, int16(34), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceSimple", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(0, 100*time.Minute, 1*time.Minute)
		// 100 laps by lap time and 1 additional
		assert.Equal(t, int16(100)+int16(1), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRace30SecondsToGo", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(99*time.Minute+30*time.Second, 100*time.Minute, 1*time.Minute)
		assert.Equal(t, int16(1), lapsLeftInRace)
	})

	//t.Run("getLapsLeftInRaceNoTimeLeft", func(t *testing.T) {
	//	timeInRace := 100 * time.Minute
	//	lapsLeftInRace := getLapsLeftInRace(timeInRace, timeInRace, 1*time.Minute)
	//	assert.Equal(t, int16(0), lapsLeftInRace)
	//})

	t.Run("getLapsLeftInRaceCheckAddedLaps", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(100*time.Minute+30*time.Second, 100*time.Minute, 1*time.Minute)
		// In last lap
		assert.Equal(t, int16(0), lapsLeftInRace)
	})

	t.Run("getLapsLeftInRaceEndOfRace", func(t *testing.T) {
		lapsLeftInRace := getLapsLeftInRace(101*time.Minute, 100*time.Minute, 1*time.Minute)
		// No lap left
		assert.Equal(t, int16(0), lapsLeftInRace)
	})
}
