package experimental

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"time"
)

func runStream(stream string, outputfolder string) (string, error) {

	if _, err := os.Stat(outputfolder); !os.IsNotExist(err) {
		err := os.RemoveAll(outputfolder)
		if err != nil {
			return "", err
		}
	}

	err := os.Mkdir(outputfolder, 0755)
	if err != nil {
		return "", err
	}
	cmd := "streamlink " + stream + " best -O | ffmpeg -i pipe:0 -r 1 " + outputfolder + "/output_%01d.jpg"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd), fmt.Errorf("error: %v", err)
	}
	return string(out), nil
}

func ReadTireDataFromStream(tr *TireData, streamurl string, filename string) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			trRead, err := ProcessImagesInFolder(filename)

			tr.FrontRight = trRead.FrontRight
			tr.FrontLeft = trRead.FrontLeft
			tr.RearLeft = trRead.RearLeft
			tr.RearRight = trRead.RearRight
			tr.LastWrite = trRead.LastWrite
			tr.Filename = trRead.Filename

			tr.AvgTireDataFrom = trRead.AvgTireDataFrom

			if err != nil {
				log.Printf("Error reading file '%s': %v\n", filename, err)
			}
		}
	}()

	for {
		response, err := runStream(streamurl, filename)
		log.Println(response)
		log.Printf("Error while starting stream of %s, %v\n", streamurl, err)
		time.Sleep(15 * time.Second) // wait 15s before restart
		log.Println("Restarting stream")
	}

}

type TireDelta struct {
	FrontLeft  int
	FrontRight int
	RearLeft   int
	RearRight  int
}

func (t *TireDelta) String() string {
	return fmt.Sprintf("FL: %d, FR: %d, RR: %d, RR: %d", t.FrontLeft, t.FrontRight, t.RearLeft, t.RearRight)
}

type TireData struct {
	FrontLeft       int
	FrontRight      int
	RearLeft        int
	RearRight       int
	Filename        string
	LastWrite       time.Time
	AvgTireDataFrom []TireData
}

func (t *TireData) String() string {
	return fmt.Sprintf("FrontLeft: %d, FrontRight: %d, RearLeft: %d, RearRight: %d", t.FrontLeft, t.FrontRight, t.RearLeft, t.RearRight)
}

// Html gives a html table for the tires relative to their position
func (t *TireData) Html() string {
	return fmt.Sprintf(
		"<table class='tiretable'>"+
			"<tr class='tirerow'>"+
			"<td>%d<td>"+
			"<td>%d<td>"+
			"</tr>"+
			"<tr class='tirerow'>"+
			"<td>%d<td>"+
			"<td>%d<td>"+
			"</tr>"+
			"</table>"+
			"", t.FrontLeft, t.FrontRight, t.RearLeft, t.RearRight,
	)
}

func ProcessImagesInFolder(folder string) (TireData, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return TireData{}, err
	}

	//sort files by date
	files = sortFilesByDate(files)
	maxReadings := 5
	tdReadings := []TireData{}

	if len(files) < maxReadings {
		log.Printf("Not enough files in folder %s\n", folder)
		return TireData{}, nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		imageFilename := path.Join(folder, file.Name())
		//log.Printf("Processing image %s\n", imageFilename)
		td, _, _, _, _, err := readTireDataFromImage(imageFilename)
		if err != nil {
			return TireData{}, err
		}

		//log.Printf("Got data %v from image %s\n", td, imageFilename)
		tdReadings = append(tdReadings, td)

		if len(tdReadings) >= maxReadings {
			avgReading := avgTdReading(tdReadings)
			log.Printf("Got %d readings\n: %d , %d, %d, %d", len(tdReadings), avgReading.FrontLeft, avgReading.FrontRight, avgReading.RearLeft, avgReading.RearRight)
			return avgReading, nil
		}

	}
	// FIXME maybe use an average
	return TireData{}, nil
}

func avgTdReading(readings []TireData) TireData {

	td := TireData{}
	for _, r := range readings {
		td.FrontLeft += r.FrontLeft
		td.FrontRight += r.FrontRight
		td.RearLeft += r.RearLeft
		td.RearRight += r.RearRight
		td.AvgTireDataFrom = append(td.AvgTireDataFrom, r)
	}
	td.FrontLeft /= len(readings)
	td.FrontRight /= len(readings)
	td.RearLeft /= len(readings)
	td.RearRight /= len(readings)

	return td
}

func sortFilesByDate(files []os.DirEntry) []os.DirEntry {
	//sort files by date
	sort.Slice(files, func(i, j int) bool {
		file1info, _ := files[i].Info()
		file2info, _ := files[j].Info()
		return file1info.ModTime().After(file2info.ModTime())
	})
	return files

}

func readTireDataFromImage(filename string) (TireData, image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle, error) {
	img, creationTime, err := loadImage(filename)
	if err != nil {
		return TireData{}, image.Rectangle{}, image.Rectangle{}, image.Rectangle{}, image.Rectangle{}, nil
	}

	tireBoxLeft, tireBoxRight, tireBoxTop, tireBoxBottom := getRelativePositionForTires(img.Bounds().Max.X, img.Bounds().Max.Y)

	imgfl, fl := getTireReading(img, tireBoxLeft, tireBoxTop)
	imgfr, fr := getTireReading(img, tireBoxRight, tireBoxTop)
	imgrl, rl := getTireReading(img, tireBoxLeft, tireBoxBottom)
	imgrr, rr := getTireReading(img, tireBoxRight, tireBoxBottom)

	tr := TireData{
		FrontLeft:  fl,
		FrontRight: fr,
		RearLeft:   rl,
		RearRight:  rr,
		Filename:   filename,
		LastWrite:  creationTime,
	}

	return tr, imgfl, imgfr, imgrl, imgrr, nil
}

func loadImage(filename string) (image.Image, time.Time, error) {
	// Open the JPG file
	file, err := os.Open(filename)
	if err != nil {
		return nil, time.Time{},  err
	}
	defer file.Close()

	// Decode the JPG image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, time.Time{},  err
	}

	filestat, err := file.Stat()
	if err != nil {
		return nil, time.Time{},  err
	}
	return img, filestat.ModTime(),  nil
}

func getRelativePositionForTires(maxX int, maxY int) (int, int, int, int) {
	relativePositionTireBoxTopLeft := float32(391) / float32(1920)
	relativePositionTireBoxTopRight := float32(476) / float32(1920)

	relativePositionTireBoxBottomLeft := float32(960) / float32(1080)
	relativePositionTireBoxBottomRight := float32(1011) / float32(1080)

	return int(relativePositionTireBoxTopLeft * float32(maxX)), int(relativePositionTireBoxTopRight * float32(maxX)), int(relativePositionTireBoxBottomLeft * float32(maxY)), int(relativePositionTireBoxBottomRight * float32(maxY))
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func getTireReading(img image.Image, x int, y int) (image.Image, int) {
	tireHeight := 36
	tireWidth := 1
	topBar := image.Rect(x, y, x+tireWidth, y+tireHeight) // Top bar, 10 pixels height
	croppedImage := img.(SubImager).SubImage(image.Point{tireHeight, tireWidth})

	// Calculate the total pixels and reddish pixels in each bar
	tireTotalPixels := topBar.Dy()
	tireReddishPixels := countReddishPixels(img, topBar)
	tireReddishPercentage := float64(tireReddishPixels) / float64(tireTotalPixels) * 100
	return croppedImage, int(100 - tireReddishPercentage)
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
