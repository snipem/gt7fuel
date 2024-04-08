package experimental

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

func Test_processImage(t *testing.T) {
	//t.Skipf("Skipping test")
	go fmt.Println(runStream("https://www.twitch.tv/videos/2079255269", "test"))
	time.Sleep(5 * time.Second)
	_, err := processImage("suzuka.jpg")
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

func Test_processImage1(t *testing.T) {
	tr, err := processImage("suzuka_test.jpg")
	assert.NoError(t, err)

	assert.Equal(t, 83, tr.FrontLeft)
	assert.Equal(t, 83, tr.FrontRight)
	assert.Equal(t, 86, tr.RearLeft)
	assert.Equal(t, 91, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}
func Test_processImage_nodata(t *testing.T) {
	tr, err := processImage("gt7fuelstream_live.jpg")
	assert.NoError(t, err)

	assert.Equal(t, 100, tr.FrontLeft)
	assert.Equal(t, 100, tr.FrontRight)
	assert.Equal(t, 100, tr.RearLeft)
	assert.Equal(t, 100, tr.RearRight)

	assert.NotNil(t, tr.LastWrite)
	assert.NotNil(t, tr.Filename)

}
