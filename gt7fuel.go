package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var gt7c *gt7.GT7Communication
var gt7stats *lib.Stats
var fuelConsumptionLastLap float32
var raceStartTime time.Time
var raceTimeInMinutes int

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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
	gt7stats.LastData = &gt7c.LastData
	for {
		counter++
		gt7stats.SetManualSetRaceDuration(time.Duration(raceTimeInMinutes) * time.Minute)

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

func main() {

	raceTimeInMinutes = 60

	gt7c = gt7.NewGT7Communication("255.255.255.255")
	go gt7c.Run()

	gt7stats = &lib.Stats{}

	go lib.LogRace(gt7c, gt7stats)

	port := ":9100"

	localurl := fmt.Sprintf("http://localhost%s", port)

	log.Printf("Server started at %s\n", localurl)
	err := lib.Open(localurl)
	if err != nil {
		log.Fatal(err)
	}
	setupRoutes()
	log.Fatal(http.ListenAndServe(port, nil))

}
