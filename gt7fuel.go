package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var gt7c *gt7.GT7Communication
var gt7stats *lib.Stats
var raceTimeInMinutes int

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func LogRace(c *gt7.GT7Communication, gt7stats *lib.Stats, i *int) {
	for {
		lib.LogTick(&c.LastData, gt7stats, i)
		time.Sleep(100 * time.Millisecond)
	}
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
func stayAwake() {

	if runtime.GOOS == "darwin" {
		log.Println("Staying wake on Mac")

		cmd := exec.Command("caffeinate", "-d")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Staying awake was ended by: %v", err)
		}
		log.Println("Staying awake ended")
	} else {
		log.Printf("Staying awake is only supported on Mac not on %s\n", runtime.GOOS)
	}

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
			log.Printf("Error writing JSON: %s\n", err)
			time.Sleep(10 * time.Second)
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
	for {
		run(raceTimeInMinutes)
		log.Println("Sleeping 10 seconds ...")
		time.Sleep(10 * time.Second)
	}

}

func run(int) {

	gt7c = gt7.NewGT7Communication("255.255.255.255")
	go func() {

		for {
			err := gt7c.Run()
			if err != nil {
				log.Printf("error running gt7c.Run(): %v", err)
			}
			log.Println("Sleeping 10 seconds before restarting gt7c.Run()")
			time.Sleep(10 * time.Second)
		}
	}()

	gt7stats = lib.NewStats()

	go LogRace(gt7c, gt7stats, &raceTimeInMinutes)

	port := ":9100"

	localurl := fmt.Sprintf("http://localhost%s", port)

	log.Printf("Server started at %s\n", localurl)

	go stayAwake()

	err := open(localurl)
	if err != nil {
		log.Fatalf("Error opening browser: %v", err)
	}
	setupRoutes()
	log.Println(http.ListenAndServe(port, nil))
}
