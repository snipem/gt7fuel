package experimental

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"image"
	"image/png"
	"os"
	"path"
	"testing"
	"time"
)

func Test_processImage(t *testing.T) {
	//t.Skipf("Skipping test")
	go fmt.Println(runStream("https://www.twitch.tv/videos/2079255269", "test"))
	time.Sleep(5 * time.Second)
	_, _, _, _, _, err := readTireDataFromImage("suzuka.jpg")
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
	tr, _, _, _, _, err := readTireDataFromImage("suzuka_test.jpg")
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

	assert.Equal(t, 100, tr.FrontLeft)
	assert.Equal(t, 100, tr.FrontRight)
	assert.Equal(t, 100, tr.RearLeft)
	assert.Equal(t, 100, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}

func Test_readTireDataFromImage(t *testing.T) {
	_, flimg, frimg, rlimg, rrimg, err := readTireDataFromImage("testdata_in/SHARE_20240428_1236290_marked.png")
	if err != nil {
		return
	}
	writeToFile("testdata/SHARE_20240428_1236290.jpeg_fl.jpeg", flimg)
	writeToFile("testdata/SHARE_20240428_1236290.jpeg_fr.jpeg", frimg)
	writeToFile("testdata/SHARE_20240428_1236290.jpeg_rl.jpeg", rlimg)
	writeToFile("testdata/SHARE_20240428_1236290.jpeg_rr.jpeg", rrimg)
}

func writeToFile(filename string, img image.Rectangle) {

	// Create a new file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Encode the image as PNG and write it to the file
	if err := png.Encode(file, img); err != nil {
		panic(err)
	}
}

func Test_getRelativePositionForTires(t *testing.T) {
	topLeft, topRight, bottomLeft, bottomRight := getRelativePositionForTires(1920, 1080)

	assert.Equal(t, 391, topLeft)
	assert.Equal(t, 476, topRight)
	assert.Equal(t, 960, bottomLeft)
	assert.Equal(t, 1011, bottomRight)
}

func Test_getRelativePositionForTires4k(t *testing.T) {
	topLeft, topRight, bottomLeft, bottomRight := getRelativePositionForTires(2*1920, 2*1080)

	assert.Equal(t, 391, topLeft)
	assert.Equal(t, 476, topRight)
	assert.Equal(t, 960, bottomLeft)
	assert.Equal(t, 1011, bottomRight)
}
