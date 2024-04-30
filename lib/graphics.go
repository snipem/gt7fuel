package lib

import (
	"bytes"
	svg "github.com/ajstarks/svgo"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"math"
)

func DrawLapToSVG(lap Lap) string {

	maxx, maxz, minx, minz := getMaxMinValuesForCoordinates(lap.DataHistory)

	buf := new(bytes.Buffer)
	// FIXME get this from the actual lap
	//width := 2 * int(maxx)
	//height := 2* int(maxz)

	canvas := svg.New(buf)
	//canvas.Start(width, height)
	canvas.Startview(int(maxx), int(maxz), int(minx), int(minz),
		int(math.Abs(float64(minz)))+int(maxz),
		int(math.Abs(float64(minx)))+int(maxx),
		)

	// higher is less detail
	detail := 5

	for i, _ := range lap.DataHistory {

		if i > detail && i%detail == 0 {
			x1 := int(lap.DataHistory[i-detail].PositionX)
			y1 := int(lap.DataHistory[i-detail].PositionZ)
			x2 := int(lap.DataHistory[i].PositionX)
			y2 := int(lap.DataHistory[i].PositionZ)
			//canvas.Circle(x1, y1, 1, "fill:white")
			canvas.Line(x1, y1, x2, y2, "stroke:white;stroke-width:10")
			//canvas.Polygon([]int{x1, x2, x2, x1}, []int{y1, y1, y2, y2}, "fill:none;stroke:white;stroke-width:7")
		}

		// Add Connection between last and first

	}

	//canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()

	return buf.String()

}

func getMaxMinValuesForCoordinates(history []gt7.GTData) (float64, float64, float64, float64) {
	maxz := float32(0)
	minz := float32(math.MaxFloat32)
	maxx := float32(0)
	minx := float32(math.MaxFloat32)

	for i := 0; i < len(history); i++ {
		if history[i].PositionZ > maxz {
			maxz = history[i].PositionZ
		}

		if history[i].PositionZ < minz {
			minz = history[i].PositionZ
		}

		if history[i].PositionX > maxx {
			maxx = history[i].PositionX
		}

		if history[i].PositionX < minx {
			minx = history[i].PositionX
		}
	}
	return float64(maxx), float64(maxz), float64(minx), float64(minz)
}
