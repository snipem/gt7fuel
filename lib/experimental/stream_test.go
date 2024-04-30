package experimental

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func Test_processImage(t *testing.T) {
	//t.Skipf("Skipping test")
	go fmt.Println(runStream("https://www.twitch.tv/videos/2079255269", "test"))
	time.Sleep(5 * time.Second)
	_, _, _, _, _, err := readTireDataFromImage("testdata_in/suzuka.jpg")
	assert.NoError(t, err)
}

func TestProcessHighlightVideo(t *testing.T) {
	//t.Skipf("Skipping test")
	filename := path.Join("testdata")
	tr := &TireData{}
	go ReadTireDataFromStream(tr, "https://clips.twitch.tv/DeafAuspiciousKittenCorgiDerp-7jrOJ2ywt21QhNc1", filename)
	time.Sleep(15 * time.Second)
	assert.NotNil(t, tr.LastWrite)
	assert.LessOrEqual(t, tr.LastWrite.Unix(), time.Now().Unix())
	assert.LessOrEqual(t, tr.FrontLeft, 51)
	assert.LessOrEqual(t, tr.FrontRight, 50)
	assert.LessOrEqual(t, tr.RearLeft, 75)
	assert.LessOrEqual(t, tr.RearRight, 75)

}

func Test_processPSAppImage(t *testing.T) {
	tr, _, _, _, _, err := readTireDataFromImage("testdata_in/SHARE_20240428_1236290.png")
	assert.NoError(t, err)

	assert.Equal(t, 83, tr.FrontLeft)
	assert.Equal(t, 83, tr.FrontRight)
	assert.Equal(t, 86, tr.RearLeft)
	assert.Equal(t, 91, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}

func Test_processImage1(t *testing.T) {
	tr, _, _, _, _, err := readTireDataFromImage("testdata_in/suzuka_test.jpg")
	assert.NoError(t, err)

	assert.Equal(t, 83, tr.FrontLeft)
	assert.Equal(t, 83, tr.FrontRight)
	assert.Equal(t, 86, tr.RearLeft)
	assert.Equal(t, 91, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}
func Test_processImage_nodata(t *testing.T) {
	tr, _, _, _, _, err := readTireDataFromImage("gt7fuelstream_live.jpg")
	assert.NoError(t, err)

	assert.Equal(t, 0, tr.FrontLeft)
	assert.Equal(t, 0, tr.FrontRight)
	assert.Equal(t, 0, tr.RearLeft)
	assert.Equal(t, 0, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}

func Test_readTireDataFromImage(t *testing.T) {
	inputfilename := "testdata_in/SHARE_20240428_1236290_marked.png"
	outnames := "testdata/SHARE_20240428_1236290_marked.png"

	writeTiresToDisk(nil, inputfilename, outnames)
}

func Test_readTireDataFromImageSuzuka(t *testing.T) {
	inputfilename := "testdata_in/suzuka_test.jpg"
	outnames := "testdata_out/suzuka_test.jpg"

	td := writeTiresToDisk(t, inputfilename, outnames)
	assert.Equal(t, 83, td.FrontLeft)
	assert.Equal(t, 88, td.FrontRight)
	assert.Equal(t, 91, td.RearLeft)
	assert.Equal(t, 95, td.RearRight)
}

func writeTiresToDisk(t *testing.T, inputfilename string, outnames string) TireData {
	td, flimg, frimg, rlimg, rrimg, err := readTireDataFromImage(inputfilename)
	assert.NoError(t, err)

	filetype := "jpg"
	if strings.HasSuffix(strings.ToLower(inputfilename), ".png") {
		filetype = "png"
	}

	writeToFile(outnames+"_fl."+filetype, flimg)
	writeToFile(outnames+"_fr."+filetype, frimg)
	writeToFile(outnames+"_rl."+filetype, rlimg)
	writeToFile(outnames+"_rr."+filetype, rrimg)

	return td
}

func writeToFile(filename string, img image.Image) {

	// Create a new file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		// Encode the image as PNG and write it to the file
		if err := png.Encode(file, img); err != nil {
			panic(err)
		}
	} else {
		// Encode the image as JPEG and write it to the file
		if err := jpeg.Encode(file, img, nil); err != nil {
			panic(err)
		}
	}

}

func Test_getRelativePositionForTires(t *testing.T) {
	topLeft, topRight, bottomLeft, bottomRight, tireHeight, tireWidth := getRelativePositionForTires(1920, 1080)

	assert.Equal(t, 386, topLeft)
	assert.Equal(t, 470, topRight)
	assert.Equal(t, 960, bottomLeft)
	assert.Equal(t, 1014, bottomRight)
	assert.Equal(t, 35, tireHeight)
	assert.Equal(t, 11, tireWidth)
}

func Test_getRelativePositionForTires4k(t *testing.T) {
	topLeft, topRight, bottomLeft, bottomRight, tireHeight, tireWidth := getRelativePositionForTires(2*1920, 2*1080)

	assert.Equal(t, 772, topLeft)
	assert.Equal(t, 940, topRight)
	assert.Equal(t, 1920, bottomLeft)
	assert.Equal(t, 2028, bottomRight)
	assert.Equal(t, 70, tireHeight)
	assert.Equal(t, 22, tireWidth)
}
