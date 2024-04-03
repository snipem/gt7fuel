package lib

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_processImage(t *testing.T) {
	t.Skipf("Skipping test")
	go fmt.Println(runStream("https://www.twitch.tv/videos/2079255269", "suzuka.jpg"))
	time.Sleep(5 * time.Second)
	_, err := processImage("suzuka.jpg")
	assert.NoError(t, err)
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
