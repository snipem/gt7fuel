package lib

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

func runStream(stream string, filename string) string {
	cmd := "rm " + filename + " ; streamlink " + stream + " best -O | ffmpeg -i pipe:0 -filter:v fps=1 -update 1 " + filename
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	return string(out)
}

func ReadStream(tr *TireData) {
	filename := path.Join(os.TempDir(), "gt7fuelstream_live.jpg")

	go func() {
		for {
			fmt.Println(runStream("https://www.twitch.tv/snimat", filename))
			time.Sleep(15 * time.Second) // wait 15s before restart
		}
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			trRead, err := processImage(filename)
			tr = &trRead
			if err != nil {
				log.Printf("Error reading %s: %v\n", filename, err)
			}
		}
	}()

}

type TireData struct {
	FrontLeft  int
	FrontRight int
	RearLeft   int
	RearRight  int
	Filename   string
	LastWrite  time.Time
}

func (t *TireData) String() string {
	return fmt.Sprintf("FrontLeft: %d, FrontRight: %d, RearLeft: %d, RearRight: %d, Filename: %s, Last Write: %s", t.FrontLeft, t.FrontRight, t.RearLeft, t.RearRight, t.Filename, t.LastWrite)
}

func processImage(filename string) (TireData, error) {
	// Open the JPG file
	file, err := os.Open(filename)
	if err != nil {
		return TireData{}, err
	}
	defer file.Close()

	// Decode the JPG image
	img, _, _ := image.Decode(file)
	//if err != nil {
	//	return TireData{}, err
	//}

	filestat, err := file.Stat()
	if err != nil {
		return TireData{}, err
	}

	tr := TireData{
		FrontLeft:  getTireReading(img, 391, 960),
		FrontRight: getTireReading(img, 476, 960),
		RearLeft:   getTireReading(img, 391, 1011),
		RearRight:  getTireReading(img, 476, 1011),
		Filename:   filename,
		LastWrite:  filestat.ModTime(),
	}

	return tr, nil
}

func getTireReading(img image.Image, x int, y int) int {
	tireHeight := 36
	tireWidth := 1
	topBar := image.Rect(x, y, x+tireWidth, y+tireHeight) // Top bar, 10 pixels height

	// Calculate the total pixels and reddish pixels in each bar
	tireTotalPixels := topBar.Dy()
	tireReddishPixels := countReddishPixels(img, topBar)
	tireReddishPercentage := float64(tireReddishPixels) / float64(tireTotalPixels) * 100
	return int(100 - tireReddishPercentage)
}

func countReddishPixels(img image.Image, rect image.Rectangle) int {
	reddishPixels := 0

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Define a threshold for the red component
			if r > 0x7FFF && g < 0x7FFF && b < 0x7FFF {
				reddishPixels++
			}
		}
	}

	return reddishPixels
}
