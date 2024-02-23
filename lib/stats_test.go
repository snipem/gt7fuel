package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_getAverageFuelConsumption(t *testing.T) {

	gt7stats := Stats{}
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

	gt7stats := Stats{}
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

		gt7stats := Stats{}
		gt7stats.Laps = []Lap{
			{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
			{FuelConsumed: 20, Number: 1, Duration: 90 * time.Second}, // valid for calculation
			{FuelConsumed: 40, Number: 2, Duration: 90 * time.Second}, // valid for calculation
			{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
		}
		assert.Equal(t, float64(90), gt7stats.GetAverageLapTime().Seconds())
	})

	t.Run("GetAverageLapTimeWithNoLaps", func(t *testing.T) {

		gt7stats := Stats{}
		gt7stats.Laps = []Lap{
			{FuelConsumed: -20, Number: 0, Duration: 80 * time.Second},
			{FuelConsumed: -20, Number: 1, Duration: 90 * time.Second},
			{FuelConsumed: -40, Number: 2, Duration: 90 * time.Second},
			{FuelConsumed: -200, Number: 3, Duration: 70 * time.Second},
		}
		assert.Equal(t, float64(0), gt7stats.GetAverageLapTime().Seconds())
	})
}
