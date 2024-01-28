package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Lap struct {
	FuelStart    float32
	FuelEnd      float32
	FuelConsumed float32
	Number       int16
	Duration     time.Duration
}

type Stats struct {
	laps []Lap
}

var laps []Lap
var fuel_consumption_last_lap float32
var raceStartTime time.Time

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var gt7c *gt7.GT7Communication

// func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
//   ws, err := upgrader.Upgrade(w, r, nil)
//   if err != nil {
//     log.Println("Error upgrading to WebSocket:", err)
//     return
//   }
//   defer ws.Close()

//   // Handle WebSocket messages here...
// }

var raceTimeInMinutes int

type Message struct {
	Speed                  string `json:"speed"`
	PackageID              int32  `json:"package_id"`
	FuelLeft               string `json:"fuel_left"`
	FuelConsumptionLastLap string `json:"fuel_consumption_last_lap"`
	TimeSinceStart         string `json:"time_since_start"`
	FuelNeededToFinishRace string `json:"fuel_needed_to_finish_race"`
	FuelConsumptionAvg     string `json:"fuel_consumption_avg"`
	FuelDiv                string `json:"fuel_div"`
	RaceTimeInMinutes      int32  `json:"race_time_in_minutes"`
}

var race_start_time time.Time

func getAverageFuelConsumption(laps []Lap) float32 {
	var totalFuelConsumption float32
	lapsAccountable := 0
	for _, lap := range laps {
		if lap.FuelConsumed > 0 && lap.Number > 0 {
			totalFuelConsumption += lap.FuelConsumed
			lapsAccountable += 1
		}
	}
	return totalFuelConsumption / float32(lapsAccountable)
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()

	counter := 0
	for {
		counter++

		timeSinceStart := time.Now().Sub(raceStartTime)

		fuelConsumptionAvg := getAverageFuelConsumption(laps)

		// it is best to use the last lap, since this will compensate for missed package etc.
		fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
			timeSinceStart,
			time.Duration(raceTimeInMinutes)*time.Minute,
			getDurationFromGT7Time(gt7c.LastData.BestLap),
			getDurationFromGT7Time(gt7c.LastData.LastLap),
			fuel_consumption_last_lap)

		fuelDiv := fuelNeededToFinishRaceInTotal - gt7c.LastData.CurrentFuel

		ws.WriteJSON(Message{
			Speed:                  fmt.Sprintf("%.0f", gt7c.LastData.CarSpeed),
			PackageID:              gt7c.LastData.PackageID,
			FuelLeft:               fmt.Sprintf("%.2f", gt7c.LastData.CurrentFuel),
			FuelConsumptionLastLap: fmt.Sprintf("%.2f", fuel_consumption_last_lap),
			FuelConsumptionAvg:     fmt.Sprintf("%.2f", fuelConsumptionAvg),
			TimeSinceStart:         getSportFormat(timeSinceStart),
			FuelNeededToFinishRace: fmt.Sprintf("%.1f", fuelNeededToFinishRaceInTotal),
			FuelDiv:                fmt.Sprintf("%.1f", fuelDiv),
			RaceTimeInMinutes:      int32(raceTimeInMinutes),
		})

		time.Sleep(100 * time.Millisecond)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleWebSocketConnection)
}

func calculateFuelNeededInTotal(totalDuration float64, bestlaptime float64, lastlaptime float64, fuelconsumedlastlap float64) float64 {

	totalDurationPlusExtraLap := totalDuration + bestlaptime
	fuelConsumptionQuota := fuelconsumedlastlap / lastlaptime
	totalFuelNeeded := totalDurationPlusExtraLap * fuelConsumptionQuota

	return totalFuelNeeded

}

func calculateFuelQuota(lastlaptime float64, fuelconsumedlastlap float64) float64 {
	return fuelconsumedlastlap / lastlaptime
}

func calculateFuelNeededToFinishRace(timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration, lastlaptime time.Duration, fuelconsumedlastlap float32) float32 {
	totalDurationPlusExtraLap := totalDurationOfRace + bestlaptime
	fuelConsumptionQuota := fuelconsumedlastlap / float32(lastlaptime.Milliseconds())

	timeLeftInRace := totalDurationPlusExtraLap - timeInRace

	return float32(timeLeftInRace.Milliseconds()) * fuelConsumptionQuota

}

func getDurationFromGT7Time(gt7time int32) time.Duration {
	seconds := gt7time / 1000
	milliseconds := gt7time % 1000

	return time.Duration(seconds)*time.Second + time.Duration(milliseconds*int32(time.Millisecond))

}

func getSportFormat(duration time.Duration) string {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	milliseconds := duration.Milliseconds() % 1000

	// If hours are present, accumulate them into minutes
	minutes += hours * 60
	return fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, milliseconds)

}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <race_time_in_minutes>")
		os.Exit(1)
	}

	// Get the first command-line argument
	arg := os.Args[1]

	// Convert the argument to an integer
	var err error
	raceTimeInMinutes, err = strconv.Atoi(arg)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	gt7c = gt7.NewGT7Communication("255.255.255.255")
	go gt7c.Run()
	//for true {
	//	fmt.Println(gt7c.LastData.CarSpeed)
	//}

	go func() {
		for {

			if gt7c.LastData.CurrentLap == 0 {
				// Race reset
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if len(laps) == 0 {
				laps = append(laps, Lap{
					FuelStart: gt7c.LastData.CurrentFuel,
					Number:    gt7c.LastData.CurrentLap,
				})
			}

			if laps[len(laps)-1].Number != gt7c.LastData.CurrentLap {
				// Change of laps detected

				if gt7c.LastData.CurrentLap == 1 {
					// First crossing of the line
					raceStartTime = time.Now()
					fmt.Printf("RACE START 🏁 %s \n", raceStartTime.Format("2006-01-02 15:04:05"))
				}

				laps[len(laps)-1].FuelEnd = gt7c.LastData.CurrentFuel
				fuel_consumption_last_lap = laps[len(laps)-1].FuelStart - laps[len(laps)-1].FuelEnd
				laps[len(laps)-1].FuelConsumed = fuel_consumption_last_lap

				laps = append(laps, Lap{
					FuelStart: gt7c.LastData.CurrentFuel,
					Number:    gt7c.LastData.CurrentLap,
				})
				fmt.Println("Add new Lap")

			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	log.Println("Server started")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":9100", nil))

}
