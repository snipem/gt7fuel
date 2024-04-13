package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"github.com/snipem/gt7fuel/lib"
	"github.com/snipem/gt7fuel/lib/experimental"
	"github.com/snipem/gt7tools/lib/dump"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"
)

var GitCommit string

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
func stayAwakeIfConnectionActive(s *lib.Stats) {

	for {
		if s.ConnectionActive {

			if runtime.GOOS == "darwin" {
				log.Println("Staying wake on Mac")
				cmd := exec.Command("caffeinate", "-d", "-t", "600") // 10 minutes
				if err := cmd.Run(); err != nil {
					log.Fatalf("Staying awake was ended by: %v", err)
				}
				//log.Println("Staying awake ended")
			} else {
				log.Printf("Staying awake is only supported on Mac not on %s\n", runtime.GOOS)
			}
		} else {
			log.Println("GT7 Connection is not active")
		}
		time.Sleep(10 * time.Second)
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

	parseTwitch := flag.Bool("parse-twitch", false, "Set to true to enable parsing Twitch")
	raceTime := flag.Int("race-time", 60, "Race time in minutes")
	twitchUrl := flag.String("twitch-url", "https://www.twitch.tv/snimat", "Twitch URL to parse")

	dumpFile := flag.String("dump-file", "", "Dump file for loading dumped data instead of real telemetry")

	// Parse command-line flags
	flag.Parse()

	fmt.Printf("Version: https://github.com/snipem/gt7fuel/commit/%s\n", GitCommit)

	for {
		run(*raceTime, *parseTwitch, *twitchUrl, *dumpFile)
		log.Println("Sleeping 10 seconds ...")
		time.Sleep(10 * time.Second)
	}

}

func run(raceTime int, parseTwitch bool, twitchResource string, dumpFilePath string) {

	gt7c = gt7.NewGT7Communication("255.255.255.255")

	if dumpFilePath != "" {

		gt7dump, err := dump.NewGT7Dump(dumpFilePath, gt7c)
		if err != nil {
			log.Fatalf("Error loading dump file: %v", err)
		}
		log.Println("Using dump file: ", dumpFilePath)
		go gt7dump.Run()

	} else {
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
	}

	gt7stats = lib.NewStats()

	if parseTwitch {
		log.Printf("Parsing Twitch for Tire Data")
		go experimental.ReadTireDataFromStream(gt7stats.LastTireData, twitchResource, path.Join(os.TempDir(), "gt7fuel"))
	}
	go LogRace(gt7c, gt7stats, &raceTimeInMinutes)

	port := ":9100"

	localurl := fmt.Sprintf("http://localhost%s", port)

	log.Printf("Server started at %s\n", localurl)

	go stayAwakeIfConnectionActive(gt7stats)

	err := open(localurl)
	if err != nil {
		log.Fatalf("Error opening browser: %v", err)
	}
	setupRoutes()
	log.Println(http.ListenAndServe(port, nil))
}
