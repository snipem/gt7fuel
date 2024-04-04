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
	cmd := "streamlink " + stream + " best -O | ffmpeg -i pipe:0 -r 1 " + outputfolder + "/output_%02d.jpg"
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

			if err != nil {
				log.Printf("Error reading file '%s': %v\n", filename, err)
			}
		}
	}()

	for {
		response, err := runStream(streamurl, filename)
		fmt.Println(response)
		fmt.Printf("Error while starting stream of %s, %v\n", streamurl, err)
		time.Sleep(15 * time.Second) // wait 15s before restart
		fmt.Println("Restarting stream")
	}

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

func ProcessImagesInFolder(folder string) (TireData, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return TireData{}, err
	}

	//sort files by date
	files = sortFilesByDate(files)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		imageFilename := path.Join(folder, file.Name())
		log.Printf("Processing image %s\n", imageFilename)
		td, err := processImage(imageFilename)
		log.Printf("Got data %v from image %s\n", td, imageFilename)
		if err != nil {
			return TireData{}, err
		}
		return td, nil
	}
	// FIXME maybe use an average
	return TireData{}, nil
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

func processImage(filename string) (TireData, error) {
	// Open the JPG file
	file, err := os.Open(filename)
	if err != nil {
		return TireData{}, err
	}
	defer file.Close()

	// Decode the JPG image
	img, _, err := image.Decode(file)
	if err != nil {
		return TireData{}, err
	}

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
