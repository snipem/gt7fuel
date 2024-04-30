package lib

import (
	"bytes"
	svg "github.com/ajstarks/svgo"
)

func DrawLapToSVG(lap Lap) string {

	buf := new(bytes.Buffer)
	// FIXME get this from the actual lap
	width := 1000
	height := 1000
	canvas := svg.New(buf)
	//canvas.Start(width, height)
	canvas.Startview(width, height, -500,-500, 500, 500)

	for i, _ := range lap.DataHistory {

		if i > 0 {
		//&& i % 10 == 0 {
			x1 := int(lap.DataHistory[i-1].PositionX)
			y1 := int(lap.DataHistory[i-1].PositionZ)
			x2 := int(lap.DataHistory[i].PositionX)
			y2 := int(lap.DataHistory[i].PositionZ)
			//canvas.Circle(x1, y1, 1, "fill:white")
			canvas.Line(x1, y1, x2, y2, "stroke:white")
		}

	}

	//canvas.Text(width/2, height/2, "Hello, SVG", "text-anchor:middle;font-size:30px;fill:white")
	canvas.End()

	return buf.String()

}
