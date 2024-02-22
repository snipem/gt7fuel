package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRoundUpAlways(t *testing.T) {
	assert.Equal(t, int32(3), RoundUpAlways(2.4))
	assert.Equal(t, int32(3), RoundUpAlways(2.6))
	assert.Equal(t, int32(2), RoundUpAlways(2.0))
	assert.Equal(t, int32(3), RoundUpAlways(2.01))
	// higher is desired
	assert.Equal(t, int32(-2), RoundUpAlways(-2.01))
}
func Test_getDurationFromGT7Time(t *testing.T) {

	duration := GetDurationFromGT7Time(int32(84149))
	expectedDuration := 1*time.Minute +
		24*time.Second +
		149*time.Millisecond

	assert.Equal(t,
		expectedDuration,
		duration)

}
func Test_getSportFormat(t *testing.T) {
	t.Run("getSportFormat", func(t *testing.T) {
		formattedTime := GetSportFormat(1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "01:30.010", formattedTime)
	})

	t.Run("getSportFormatManyMilliseconds", func(t *testing.T) {
		formattedTime := GetSportFormat(1*time.Minute + 30*time.Second + 1010*time.Millisecond)
		assert.Equal(t, "01:31.010", formattedTime)
	})

	t.Run("getSportFormatWithHours", func(t *testing.T) {
		formattedTime := GetSportFormat(2*time.Hour + 1*time.Minute + 30*time.Second + 10*time.Millisecond)
		assert.Equal(t, "121:30.010", formattedTime)
	})
}
