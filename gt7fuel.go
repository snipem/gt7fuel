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
	"reflect"
	"runtime"
	"strconv"
	"time"
)

var GitCommit string

var gt7c *gt7.GT7Communication
var gt7stats *lib.Stats
var raceTimeInMinutes int

var WaitTime = 100 * time.Millisecond

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func LogRace(c *gt7.GT7Communication, gt7stats *lib.Stats, i *int) {
	for gt7stats.ShallRun {
		lib.LogTick(&c.LastData, gt7stats, i)
		wait()
	}
}

func wait() {
	time.Sleep(WaitTime)
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
				//log.Printf("Staying awake is only supported on Mac not on %s\n", runtime.GOOS)
			}
		} else {
			log.Println("GT7 Connection is not active")
		}
		time.Sleep(10 * time.Second)
	}

}
func handleHeavyWebSocketConnection(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()
	log.Println("Have websocket connection")

	counter := 0
	firstContact := true
	for {
		if gt7stats.HeavyMessageNeedsRefresh || firstContact {
			counter++

			message := gt7stats.GetHeavyMessage()
			err = ws.WriteJSON(message)
			log.Printf("Sent a heavy message")
			gt7stats.HeavyMessageNeedsRefresh = false
			firstContact = false
			if err != nil {
				log.Printf("Error writing JSON: %s, ending connection\n", err)
				return // return browser has to reestablish connection
			}

			wait()
			counter = 0
		}
	}
}

func handleRealtimeWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer ws.Close()
	log.Println("Have websocket connection")

	lastMessage := lib.RealTimeMessage{}
	counter := 0
	equalMessageCounter := 0
	oldPackageId := int32(0)
	for {
		if oldPackageId != gt7c.LastData.PackageID {
			counter++

			gt7stats.SetManualSetRaceDuration(time.Duration(raceTimeInMinutes) * time.Minute)

			message := gt7stats.GetRealTimeMessage()
			if !reflect.DeepEqual(lastMessage, message) {
				err = ws.WriteJSON(message)
				if err != nil {
					log.Printf("Error writing JSON: %s, ending connection\n", err)
					return // return browser has to reestablish connection
				}
				lastMessage = message
			} else {
				equalMessageCounter++
			}
			if counter%1000 == 0 {
				log.Printf("Saved %0.f%% messages because of check for equality", 100*float64(equalMessageCounter)/float64(counter))
			}

			wait()
			oldPackageId = gt7c.LastData.PackageID
			counter = 0
		}
	}
}

func isEqualMessage(m lib.RealTimeMessage, m2 lib.RealTimeMessage) bool {
	// TODO maybe implement me
	return false
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
	http.HandleFunc("/realtimews", handleRealtimeWebSocketConnection)
	http.HandleFunc("/heavyws", handleHeavyWebSocketConnection)
}

func main() {

	//FIXME delete the oldest pictures in the tire parsing dir
	parseTwitch := flag.Bool("parse-twitch", true, "Set to true to enable parsing Twitch")
	raceTime := flag.Int("race-time", 60, "Race time in minutes")
	twitchUrl := flag.String("twitch-url", "", "Twitch channel URL to parse")

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

	// set global var from parameter
	raceTimeInMinutes = raceTime

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
