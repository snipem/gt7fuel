package lib

import (
	"github.com/snipem/gt7tools/lib/dump"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func Test_drawLap(t *testing.T) {

	lap := getLapFromDump()
	svg := DrawLapToSVG(lap)
	assert.Contains(t, svg, "<svg")
	// save sting to file
	f, err := os.Create(filepath.Join("testdata_out", "test_lap.svg"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(svg)
}

func getLapFromDump() Lap {
	lap := Lap{}
	glennrun, err := dump.ReadGT7Data(filepath.Join("..", "testdata", "gt7testdata", "watkinsglen.gob.gz"))
	if err != nil {
		panic(err)
	}

	for _, data := range glennrun {
		if data.CurrentLap == 2 {
			lap.DataHistory = append(lap.DataHistory, data)
		} else if data.CurrentLap > 1 {
			break
		}
	}

	return lap

}

func TestDrawLapEmpty(t *testing.T) {

	svg := DrawLapToSVG(Lap{})
	assert.Contains(t, svg, "<svg")
	assert.NotContains(t, svg, "path")
	// save sting to file
	f, err := os.Create(filepath.Join("testdata_out", "test_lap_empty.svg"))
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(svg)
}
