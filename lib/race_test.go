package lib

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_logTick(t *testing.T) {
	raceTimeInMinutes := 6
	s := NewStats()
	ld := &gt7.GTData{
		PackageID:         0,
		BestLap:           0,
		LastLap:           0,
		CurrentLap:        0,
		CurrentGear:       0,
		SuggestedGear:     0,
		FuelCapacity:      0,
		CurrentFuel:       0,
		Boost:             0,
		TyreDiameterFL:    0,
		TyreDiameterFR:    0,
		TyreDiameterRL:    0,
		TyreDiameterRR:    0,
		TypeSpeedFL:       0,
		TypeSpeedFR:       0,
		TypeSpeedRL:       0,
		TyreSpeedRR:       0,
		CarSpeed:          0,
		TyreSlipRatioFL:   "",
		TyreSlipRatioFR:   "",
		TyreSlipRatioRL:   "",
		TyreSlipRatioRR:   "",
		TimeOnTrack:       gt7.Duration{},
		TotalLaps:         0,
		CurrentPosition:   0,
		TotalPositions:    0,
		CarID:             0,
		Throttle:          0,
		RPM:               0,
		RPMRevWarning:     0,
		Brake:             0,
		RPMRevLimiter:     0,
		EstimatedTopSpeed: 0,
		Clutch:            0,
		ClutchEngaged:     0,
		RPMAfterClutch:    0,
		OilTemp:           0,
		WaterTemp:         0,
		OilPressure:       0,
		RideHeight:        0,
		TyreTempFL:        0,
		TyreTempFR:        0,
		SuspensionFL:      0,
		SuspensionFR:      0,
		TyreTempRL:        0,
		TyreTempRR:        0,
		SuspensionRL:      0,
		SuspensionRR:      0,
		Gear1:             0,
		Gear2:             0,
		Gear3:             0,
		Gear4:             0,
		Gear5:             0,
		Gear6:             0,
		Gear7:             0,
		Gear8:             0,
		PositionX:         0,
		PositionY:         0,
		PositionZ:         0,
		VelocityX:         0,
		VelocityY:         0,
		VelocityZ:         0,
		RotationPitch:     0,
		RotationYaw:       0,
		RotationRoll:      0,
		AngularVelocityX:  0,
		AngularVelocityY:  0,
		AngularVelocityZ:  0,
		IsPaused:          false,
		InRace:            false,
	}

	// First tick, pre race
	ld.CurrentFuel = 100
	ld.LastLap = 0
	ld.PackageID += 1

	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 0)

	// Do another log tick, laps should still be 0
	ld.CurrentFuel = 99
	ld.LastLap = 0
	ld.PackageID += 1
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 0)

	// Race Start Start Lap 1
	ld.CurrentFuel = 98
	ld.LastLap = 0
	ld.CurrentLap = 1 // RACE START FROM NO ON!
	ld.PackageID += 1
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 0) // Should have lap now, the ongoing

	// Start Lap 2
	ld.CurrentFuel = 95
	ld.CurrentLap = 2
	ld.LastLap = 3 * 60 * 1000
	ld.PackageID += 1
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 1) // Should have lap now, the last and the ongoing

	// Start Lap 3
	ld.CurrentFuel = 93
	ld.CurrentLap = 3
	ld.LastLap = 2 * 60 * 1000
	ld.PackageID += 1
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 2) // Should have lap now, the last and the ongoing

	// No should have only 2 laps, because lap 3 is not completed yet

	assert.False(t, s.raceStartTime.IsZero()) // check if race start time has been logged

	assert.Equal(t, float32(2.5), s.GetAverageFuelConsumption())
	averageLapTime, err := s.GetAverageLapTime()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(2*time.Minute+30*time.Second), averageLapTime)

	// Start a new race
	// First tick, pre race
	ld.CurrentFuel = 100
	ld.CurrentLap = 0

	ld.PackageID += 1
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.Len(t, s.Laps, 0)

	// First time package id is not incremented, connection be active now
	assert.True(t, s.ConnectionActive)
	_ = LogTick(ld, s, &raceTimeInMinutes)
	assert.False(t, s.ConnectionActive)

}
