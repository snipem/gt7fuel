package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"log"
	"net/http"
	"time"
)

type Lap struct {
	fuel_start       float32
	fuel_end         float32
	fuel_consumption float32
	nr               int16
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

var race_time_in_minutes = 60

type Message struct {
	Speed                  string `json:"speed"`
	PackageID              int32  `json:"package_id"`
	FuelLeft               string `json:"fuel_left"`
	FuelConsumptionLastLap string `json:"fuel_consumption_last_lap"`
	TimeSinceStart         string `json:"time_since_start"`
	FuelNeededToFinishRace string `json:"fuel_needed_to_finish_race"`
	FuelConsumptionAvg     string `json:"fuel_consumption_avg"`
	FuelDiv                string `json:"fuel_div"`
}

var race_start_time time.Time

func getAverageFuelConsumption(laps []Lap) float32 {
	var totalFuelConsumption float32
	lapsAccountable := 0
	for _, lap := range laps {
		if lap.fuel_consumption > 0 && lap.nr > 0 {
			totalFuelConsumption += lap.fuel_consumption
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

		//"speed":      int(gt7c.LastData.CarSpeed),
		//	"package_id": int(gt7c.LastData.PackageID),
		//	"fuel_left":  int(gt7c.LastData.CurrentFuel)} +
		//map[string]float32{
		//	"fuel_consumption_last_lap": fuel_consumption_last_lap,
		//},

		timeSinceStart := time.Now().Sub(race_start_time)

		// it is best to use the last lap, since this will compensate for missed package etc.
		fuelNeededToFinishRace := calculateFuelNeededToFinishRace(
			timeSinceStart,
			time.Duration(race_time_in_minutes)*time.Minute,
			getDurationFromGT7Time(gt7c.LastData.BestLap),
			getDurationFromGT7Time(gt7c.LastData.LastLap),
			fuel_consumption_last_lap)

		ws.WriteJSON(Message{
			Speed:                  fmt.Sprintf("%.0f", gt7c.LastData.CarSpeed),
			PackageID:              gt7c.LastData.PackageID,
			FuelLeft:               fmt.Sprintf("%.2f", gt7c.LastData.CurrentFuel),
			FuelDiv:                fmt.Sprintf("%.2f", gt7c.LastData.CurrentFuel-fuelNeededToFinishRace),
			FuelConsumptionLastLap: fmt.Sprintf("%.2f", fuel_consumption_last_lap),
			FuelConsumptionAvg:     fmt.Sprintf("%.2f", getAverageFuelConsumption(laps)),
			TimeSinceStart:         getSportFormat(timeSinceStart),
			FuelNeededToFinishRace: fmt.Sprintf("%.1f", fuelNeededToFinishRace),
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
	return duration.String()
}

func main() {

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
					fuel_start: gt7c.LastData.CurrentFuel,
					nr:         gt7c.LastData.CurrentLap,
				})
			}

			if laps[len(laps)-1].nr != gt7c.LastData.CurrentLap {
				// Change of laps detected

				if gt7c.LastData.CurrentLap == 1 {
					// First crossing of the line
					race_start_time = time.Now()
					fmt.Printf("RACE START 🏁 %s \n", race_start_time.Format("2006-01-02 15:04:05"))
				}

				laps[len(laps)-1].fuel_end = gt7c.LastData.CurrentFuel
				fuel_consumption_last_lap = laps[len(laps)-1].fuel_start - laps[len(laps)-1].fuel_end
				laps[len(laps)-1].fuel_consumption = fuel_consumption_last_lap

				laps = append(laps, Lap{
					fuel_start: gt7c.LastData.CurrentFuel,
					nr:         gt7c.LastData.CurrentLap,
				})
				fmt.Println("Add new Lap")
				fmt.Println(gt7c.LastData.BestLap)

			}
			//fmt.Printf("Current Lap %d\n", gt7c.LastData.CurrentLap)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	fmt.Println("Server gestartet")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":9100", nil))

}
