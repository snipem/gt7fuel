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

var fuelConsumptionLastLap float32
var raceStartTime time.Time

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var gt7c *gt7.GT7Communication

var raceTimeInMinutes int

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
		gt7stats.LastData = &gt7c.LastData
		gt7stats.SetManualSetRaceDuration(time.Duration(raceTimeInMinutes) * time.Minute)
		gt7stats.SetRaceStartTime(raceStartTime)
		gt7stats.SetFuelConsumptionLastLap(fuelConsumptionLastLap)

		err := ws.WriteJSON(gt7stats.GetMessage())

		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}
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

//func calculateFuelNeededInTotal(totalDuration float64, bestlaptime float64, lastlaptime float64, fuelconsumedlastlap float64) float64 {
//
//	totalDurationPlusExtraLap := totalDuration + bestlaptime
//	fuelConsumptionQuota := fuelconsumedlastlap / lastlaptime
//	totalFuelNeeded := totalDurationPlusExtraLap * fuelConsumptionQuota
//
//	return totalFuelNeeded
//
//}
//
//func calculateFuelQuota(lastlaptime float64, fuelconsumedlastlap float64) float64 {
//	return fuelconsumedlastlap / lastlaptime
//}

var gt7stats lib.Stats

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

	gt7c := gt7.NewGT7Communication("255.255.255.255")
	go gt7c.Run()

	gt7stats.Init()

	go logGt7(gt7c)

	port := ":9100"

	localurl := fmt.Sprintf("http://localhost%s", port)

	log.Printf("Server started at %s\n", localurl)
	open(localurl)
	setupRoutes()
	log.Fatal(http.ListenAndServe(port, nil))

}

func logGt7(c *gt7.GT7Communication) {
	for {

		if c.LastData.CurrentLap == 0 {
			// Race reset

			fuelConsumptionLastLap = 0
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if len(gt7stats.Laps) == 0 {
			gt7stats.Laps = append(gt7stats.Laps, lib.Lap{
				FuelStart: c.LastData.CurrentFuel,
				Number:    c.LastData.CurrentLap,
			})
		}

		if gt7stats.Laps[len(gt7stats.Laps)-1].Number != c.LastData.CurrentLap {
			// Change of laps detected

			if c.LastData.CurrentLap == 1 {
				// First crossing of the line
				raceStartTime = time.Now()
				fuelConsumptionLastLap = 0
				fmt.Printf("RACE START 🏁 %s \n", raceStartTime.Format("2006-01-02 15:04:05"))
			}

			gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd = c.LastData.CurrentFuel
			gt7stats.Laps[len(gt7stats.Laps)-1].Duration = lib.GetDurationFromGT7Time(c.LastData.LastLap)

			// Do not log last laps fuel consumption in the first lap
			if c.LastData.CurrentLap != 1 {
				fuelConsumptionLastLap = gt7stats.Laps[len(gt7stats.Laps)-1].FuelStart - gt7stats.Laps[len(gt7stats.Laps)-1].FuelEnd
				gt7stats.Laps[len(gt7stats.Laps)-1].FuelConsumed = fuelConsumptionLastLap
			}

			fmt.Printf("Add new Lap. Last Lap was: %s\n", getLastLap(gt7stats.Laps))

			newLap := lib.Lap{
				FuelStart: c.LastData.CurrentFuel,
				Number:    c.LastData.CurrentLap,
			}
			gt7stats.Laps = append(gt7stats.Laps, newLap)

		}
		// FIXME Use deep copy here
		gt7stats.LastLoggedData.FuelCapacity = c.LastData.FuelCapacity
		time.Sleep(100 * time.Millisecond)
	}
}

func getLastLap(l []lib.Lap) *lib.Lap {
	return &gt7stats.Laps[len(gt7stats.Laps)-1]
}
