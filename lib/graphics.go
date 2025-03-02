package lib

import (
	"bytes"
	"fmt"
	svg "github.com/ajstarks/svgo"
	gt7 "github.com/snipem/go-gt7-telemetry/lib"
	"math"
)

func DrawLapToSVG(lap Lap) string {

	// add pixels to left and right side,
	// height will scale due to kept aspect ratio
	padding := 20
	// higher is less detail
	detail := 5

	maxx, maxz, minx, minz := getMaxMinValuesForCoordinates(lap.DataHistory)

	buf := new(bytes.Buffer)

	widthOfTrack := int(math.Abs(minx) + math.Abs(maxx))
	heightOfTrack := int(math.Abs(minz) + math.Abs(maxz))

	//fmt.Printf("%d %d %d %d %d %d",
	//	int(maxx),
	//	int(maxz),
	//	int(minx),
	//	int(minz),
	//	int(widthOfTrack),
	//	int(heightOfTrack))

	canvas := svg.New(buf)
	//canvas.Start(width, height)
	canvas.Startview(
		widthOfTrack, heightOfTrack,
		int(minx)-padding,
		int(minz),
		widthOfTrack+2*padding, // has to double since we already subtracted pixels on the left side
		heightOfTrack,
	)

	path := ""

	for i, _ := range lap.DataHistory {

		if i > detail && i%detail == 0 {
			x1 := int(lap.DataHistory[i-detail].PositionX)
			y1 := int(lap.DataHistory[i-detail].PositionZ)
			x2 := int(lap.DataHistory[i].PositionX)
			y2 := int(lap.DataHistory[i].PositionZ)

			path += fmt.Sprintf("M %d,%d L %d,%d ", x1, y1, x2, y2)
		}
	}

	if len(lap.DataHistory) > 0 {
		// Close gap
		path += fmt.Sprintf("M %d,%d L %d,%d z", int(lap.DataHistory[len(lap.DataHistory)-1].PositionX), int(lap.DataHistory[len(lap.DataHistory)-1].PositionZ), int(lap.DataHistory[0].PositionX), int(lap.DataHistory[0].PositionZ))

		// https://www.w3.org/TR/SVG11/paths.html
		canvas.Path(path, "fill:none;stroke:black;stroke-width:14")
		canvas.Path(path, "fill:none;stroke:white;stroke-width:10")
	}

	canvas.End()

	svgoutput := buf.String()
	// for debugging
	//svg = strings.Replace(svg, "xmlns=\"http://www.w3.org/2000/svg\"", "xmlns=\"http://www.w3.org/2000/svg\" style=\"background-color:green\"", 1)

	return svgoutput

}

func getMaxMinValuesForCoordinates(history []gt7.GTData) (float64, float64, float64, float64) {

	if len(history) == 0 {
		return 0, 0, 0, 0
	}

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
