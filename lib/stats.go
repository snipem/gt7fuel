package lib

import (
	"fmt"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"time"
)

type Stats struct {
	LastLoggedData gt7.GTData
	Laps           []Lap
}

type Lap struct {
	FuelStart    float32
	FuelEnd      float32
	FuelConsumed float32
	Number       int16
	Duration     time.Duration
}

func (l Lap) String() string {
	return fmt.Sprintf("Lap %d: FuelStart=%.2f, FuelEnd=%.2f, FuelConsumed=%.2f, Duration=%s",
		l.Number, l.FuelStart, l.FuelEnd, l.FuelConsumed, l.Duration)
}

func (s Stats) Init() {
	s.LastLoggedData = gt7.GTData{}
}
func (s Stats) GetAverageFuelConsumption() float32 {
	var totalFuelConsumption float32
	lapsAccountable := GetAccountableFuelConsumption(s.Laps)

	for _, f := range lapsAccountable {
		totalFuelConsumption += f
	}
	return totalFuelConsumption / float32(len(lapsAccountable))
}

func (s Stats) GetFuelConsumptionPerMinute() float32 {
	return s.GetAverageFuelConsumption() / float32(s.GetAverageLapTime().Minutes())
}

func (s Stats) GetAverageLapTime() time.Duration {
	var totalDuration time.Duration
	accountableLaps := getAccountableLaps(s.Laps)
	for _, lap := range accountableLaps {
		totalDuration += lap.Duration
	}
	return totalDuration / time.Duration(len(accountableLaps))
}
