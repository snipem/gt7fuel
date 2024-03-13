package lib

import (
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_logTick(t *testing.T) {
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

	returnValue := false
	returnValue = logTick(ld, s)
	assert.False(t, returnValue)
	assert.Len(t, s.Laps, 0)

	ld.CurrentLap = 1
	returnValue = logTick(ld, s)
	assert.True(t, returnValue)
}
