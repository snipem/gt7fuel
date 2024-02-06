package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
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

var laps []Lap
var fuelConsumptionLastLap float32
var raceStartTime time.Time

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var gt7c *gt7.GT7Communication

var raceTimeInMinutes int

type Message struct {
	Speed                  string `json:"speed"`
	PackageID              int32  `json:"package_id"`
	FuelLeft               string `json:"fuel_left"`
	FuelConsumptionLastLap string `json:"fuel_consumption_last_lap"`
	TimeSinceStart         string `json:"time_since_start"`
	FuelNeededToFinishRace int32  `json:"fuel_needed_to_finish_race"`
	FuelConsumptionAvg     string `json:"fuel_consumption_avg"`
	FuelDiv                string `json:"fuel_div"`
	RaceTimeInMinutes      int32  `json:"race_time_in_minutes"`
	ValidState             bool   `json:"valid_state"`
	LapsLeftInRace         int16  `json:"laps_left_in_race"`
	EndOfRaceType          string `json:"end_of_race_type"`
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func getFuelConsumptionsAccountable(laps []Lap) []float32 {

	var fuelConsumptionAccountable []float32

	for _, lap := range laps {
		if lap.FuelConsumed > 0 && lap.Number > 0 {
			fuelConsumptionAccountable = append(fuelConsumptionAccountable, lap.FuelConsumed)
		}
	}
	return fuelConsumptionAccountable
}

func getAverageFuelConsumption(laps []Lap) float32 {
	var totalFuelConsumption float32
	lapsAccountable := getFuelConsumptionsAccountable(laps)

	for _, f := range lapsAccountable {
		totalFuelConsumption += f
	}
	return totalFuelConsumption / float32(len(lapsAccountable))
}

func getMedianFuelConsumption(laps []Lap) float32 {
	lapsAccountable := getFuelConsumptionsAccountable(laps)
	return lib.Median(lapsAccountable)
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()
	log.Println("Have websocket connection")

	counter := 0
	for {
		counter++

		timeSinceStart := time.Now().Sub(raceStartTime)

		fuelConsumptionAvg := getAverageFuelConsumption(laps)

		totalDurationOfRace := time.Duration(raceTimeInMinutes) * time.Minute

		endOfRaceType := ""
		lapsLeftInRace := int16(0)
		var raceDuration time.Duration
		if gt7c.LastData.TotalLaps > 0 {
			lapsLeftInRace = gt7c.LastData.TotalLaps - gt7c.LastData.CurrentLap + 1 // because the current lap is ongoing
			endOfRaceType = "By Laps"
			raceDuration = getDurationFromGT7Time(gt7c.LastData.BestLap) * time.Duration(gt7c.LastData.TotalLaps)
		} else {
			lapsLeftInRace = getLapsLeftInRace(timeSinceStart, totalDurationOfRace, getDurationFromGT7Time(gt7c.LastData.BestLap))
			endOfRaceType = "By Time"
			raceDuration = time.Duration(raceTimeInMinutes) * time.Minute
		}

		// it is best to use the last lap, since this will compensate for missed package etc.
		fuelNeededToFinishRaceInTotal := calculateFuelNeededToFinishRace(
			timeSinceStart,
			raceDuration,
			getDurationFromGT7Time(gt7c.LastData.BestLap),
			getDurationFromGT7Time(gt7c.LastData.LastLap),
			fuelConsumptionLastLap)

		fuelDiv := fuelNeededToFinishRaceInTotal - gt7c.LastData.CurrentFuel

		validState := true

		if timeSinceStart > 1000*time.Hour {
			validState = false
		} else if fuelConsumptionLastLap <= 0 {
			validState = false
		}

		err := ws.WriteJSON(Message{
			Speed:                  fmt.Sprintf("%.0f", gt7c.LastData.CarSpeed),
			PackageID:              gt7c.LastData.PackageID,
			FuelLeft:               fmt.Sprintf("%.2f", gt7c.LastData.CurrentFuel),
			FuelConsumptionLastLap: fmt.Sprintf("%.2f", fuelConsumptionLastLap),
			FuelConsumptionAvg:     fmt.Sprintf("%.2f", fuelConsumptionAvg),
			TimeSinceStart:         getSportFormat(timeSinceStart),
			FuelNeededToFinishRace: lib.RoundUpAlways(fuelNeededToFinishRaceInTotal),
			LapsLeftInRace:         lapsLeftInRace,
			EndOfRaceType:          endOfRaceType,
			FuelDiv:                fmt.Sprintf("%.0f", fuelDiv),
			RaceTimeInMinutes:      int32(raceDuration.Minutes()),
			ValidState:             validState,
		})
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func getLapsLeftInRace(timeInRace time.Duration, totalDurationOfRace time.Duration, bestLapTime time.Duration) int16 {
	lapsLeft := getTimeLeftInRaceWithExtraLap(timeInRace, totalDurationOfRace, bestLapTime) / bestLapTime
	if lapsLeft < 0 {
		return 0
	}
	return int16(lapsLeft)
}

func homePage(w http.ResponseWriter, r *http.Request) {

	m, _ := url.ParseQuery(r.URL.RawQuery)
	minsQuery := m.Get("min")
	if minsQuery != "" {
		convertedRacetimeInMinutes, err := strconv.Atoi(minsQuery)
		if err != nil {
			log.Printf("Cannot convert %s\n", minsQuery)
		} else {
			raceTimeInMinutes = convertedRacetimeInMinutes
		}
	}
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
	fuelConsumptionQuota := fuelconsumedlastlap / float32(lastlaptime.Milliseconds())

	timeLeftInRace := getTimeLeftInRaceWithExtraLap(timeInRace, totalDurationOfRace, bestlaptime)

	return float32(timeLeftInRace.Milliseconds()) * fuelConsumptionQuota

}

func getTimeLeftInRaceWithExtraLap(timeInRace time.Duration, totalDurationOfRace time.Duration, bestlaptime time.Duration) time.Duration {
	totalDurationPlusExtraLap := totalDurationOfRace + bestlaptime
	timeLeftInRace := totalDurationPlusExtraLap - timeInRace
	return timeLeftInRace
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

	go func() {
		for {

			if gt7c.LastData.CurrentLap == 0 {
				// Race reset

				fuelConsumptionLastLap = 0
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
					fuelConsumptionLastLap = 0
					fmt.Printf("RACE START 🏁 %s \n", raceStartTime.Format("2006-01-02 15:04:05"))
				}

				laps[len(laps)-1].FuelEnd = gt7c.LastData.CurrentFuel
				laps[len(laps)-1].Duration = getDurationFromGT7Time(gt7c.LastData.LastLap)

				// Do not log last laps fuel consumption in the first lap
				if gt7c.LastData.CurrentLap != 1 {
					fuelConsumptionLastLap = laps[len(laps)-1].FuelStart - laps[len(laps)-1].FuelEnd
					laps[len(laps)-1].FuelConsumed = fuelConsumptionLastLap
				}

				laps = append(laps, Lap{
					FuelStart: gt7c.LastData.CurrentFuel,
					Number:    gt7c.LastData.CurrentLap,
				})
				fmt.Printf("Add new Lap. Laps: %v\n", laps)

			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	port := ":9100"

	url := fmt.Sprintf("http://localhost%s", port)

	log.Printf("Server started at %s\n", url)
	open(url)
	setupRoutes()
	log.Fatal(http.ListenAndServe(port, nil))

}
