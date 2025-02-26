package lib

type CarPosition struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Facing float32 `json:"facing"`
}

type RealTimeMessage struct {
	Speed                      string      `json:"speed"`
	PackageID                  int32       `json:"package_id"`
	FuelLeft                   string      `json:"fuel_left"`
	FuelConsumptionLastLap     string      `json:"fuel_consumption_last_lap"`
	TimeSinceStart             string      `json:"time_since_start"`
	FuelNeededToFinishRace     int32       `json:"fuel_needed_to_finish_race"`
	FuelConsumptionAvg         string      `json:"fuel_consumption_avg"`
	FuelDiv                    string      `json:"fuel_div"`
	RaceTimeInMinutes          int32       `json:"race_time_in_minutes"`
	ValidState                 bool        `json:"valid_state"`
	LapsLeftInRace             int16       `json:"laps_left_in_race"`
	EndOfRaceType              string      `json:"end_of_race_type"`
	FuelConsumptionPerMinute   string      `json:"fuel_consumption_per_minute"`
	LowestTireTemp             float32     `json:"lowest_tire_temp"`
	ErrorMessage               string      `json:"error_message"`
	NextPitStop                int16       `json:"next_pit_stop"`
	CurrentLapProgressAdjusted string      `json:"current_lap_progress_adjusted"`
	Tires                      string      `json:"tires"`
	LapTimeDeviation           string      `json:"lap_time_deviation"`
	TireTemperatures           []int       `json:"tire_temperatures"`
	TCSActive                  bool        `json:"tcs_active"`
	ASMActive                  bool        `json:"asma_active"`
	RisingTrailbreaking        bool        `json:"rising_trailbreaking"`
	Position                   CarPosition `json:"position"`
}

type HeavyMessage struct {
	FormattedLaps string `json:"formatted_laps"`
	LapSVG        string `json:"lap_svg"`
}
