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
		waitTime := time.Duration(1) * time.Minute
		log.Println("Waiting " + waitTime.String() + " before trying to restart stream")
		time.Sleep(waitTime) // wait 15s before restart
		log.Println("Attempt to restarting stream")
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

func (t TireData) Format() string {
	return fmt.Sprintf("FL: %d, FR: %d, RL: %d, RR: %d", t.FrontLeft, t.FrontRight, t.RearLeft, t.RearRight)
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

func (t *TireData) Diff(end TireData) TireData {
	return TireData{
		FrontLeft:  t.FrontLeft - end.FrontLeft,
		FrontRight: t.FrontRight - end.FrontRight,
		RearLeft:   t.RearLeft - end.RearLeft,
		RearRight:  t.RearRight - end.RearRight,
	}
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
		//log.Printf("Not enough files in folder %s\n", folder)
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

func readTireDataFromImage(filename string) (TireData, image.Image, image.Image, image.Image, image.Image, error) {
	img, creationTime, err := loadImage(filename)
	if err != nil {
		return TireData{}, image.Rectangle{}, image.Rectangle{}, image.Rectangle{}, image.Rectangle{}, err
	}

	tireBoxLeft, tireBoxRight, tireBoxTop, tireBoxBottom, tireHeight, tireWidth := getRelativePositionForTires(img.Bounds().Max.X, img.Bounds().Max.Y)

	imgfl, fl := getTireReading(img, tireBoxLeft, tireBoxTop, tireHeight, tireWidth)
	imgfr, fr := getTireReading(img, tireBoxRight, tireBoxTop, tireHeight, tireWidth)
	imgrl, rl := getTireReading(img, tireBoxLeft, tireBoxBottom, tireHeight, tireWidth)
	imgrr, rr := getTireReading(img, tireBoxRight, tireBoxBottom, tireHeight, tireWidth)

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
		return nil, time.Time{}, err
	}
	defer file.Close()

	// Decode the JPG image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, time.Time{}, err
	}

	filestat, err := file.Stat()
	if err != nil {
		return nil, time.Time{}, err
	}
	return img, filestat.ModTime(), nil
}

func getRelativePositionForTires(maxX int, maxY int) (int, int, int, int, int, int) {
	// We used 4k to read the initial values
	const referenceX = float32(3840)
	const referenceY = float32(2160)

	relativeTireHeight := float32(70) / float32(referenceY)
	relativeTireWidth := float32(22) / float32(referenceX)

	// the position of the left and the right tire box (x)
	relativePositionTireBoxLeft := float32(772) / float32(referenceX)
	relativePositionTireBoxRight := float32(940) / float32(referenceX)

	// The hight of the tire boxes on the back (y)
	relativePositionTireBoxTop := float32(1920) / float32(referenceY)
	relativePositionTireBoxBottom := float32(2028) / float32(referenceY)

	return int(relativePositionTireBoxLeft * float32(maxX)), int(relativePositionTireBoxRight * float32(maxX)), int(relativePositionTireBoxTop * float32(maxY)), int(relativePositionTireBoxBottom * float32(maxY)), int(relativeTireHeight * float32(maxY)), int(relativeTireWidth * float32(maxX))
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func getTireReading(img image.Image, x int, y int, tireHeight int, tireWidth int) (image.Image, int) {
	cropSize := image.Rect(x, y, x+tireWidth, y+tireHeight)

	croppedImage := img.(SubImager).SubImage(cropSize)

	// Calculate the total pixels and reddish pixels in each bar
	tireTotalPixels := cropSize.Dy() * cropSize.Dx()
	tireReddishPixels := countReddishPixels(img, cropSize)
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
