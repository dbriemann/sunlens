package utils

//Color defines a color in RGB with values from 0 to 5 each. -> 216 colors
type Color struct {
	R uint8
	G uint8
	B uint8
}

type HeatColor struct {
	Temperature float64
	Color       Color
}

func NewColorByTemp(temp float64, heatMap []HeatColor) Color {
	//min color
	if temp < heatMap[0].Temperature {
		return heatMap[0].Color
	}
	//max color
	if temp > heatMap[len(heatMap)-1].Temperature {
		return heatMap[len(heatMap)-1].Color
	}

	//color in between min and max
	loColor := heatMap[0]
	hiColor := heatMap[len(heatMap)-1]

	for i, col := range heatMap {
		if temp <= col.Temperature {
			//set color col as upper bound
			hiColor = col
			//set last color as lower bound if possible
			if i > 0 {
				loColor = heatMap[i-1]
			}
			//colors set, break
			break
		}
	}

	//create color in between bounds
	c := ColorByInterpolation(&loColor, &hiColor, temp)

	return c
}

func ColorByInterpolation(hcLo, hcHi *HeatColor, temp float64) Color {
	//normalize heat to [0,1]
	heat := (temp - hcLo.Temperature) / (hcHi.Temperature - hcLo.Temperature)

	c := Color{}

	c.R = uint8(float64(int(hcHi.Color.R)-int(hcLo.Color.R))*heat + float64(hcLo.Color.R))
	c.G = uint8(float64(int(hcHi.Color.G)-int(hcLo.Color.G))*heat + float64(hcLo.Color.G))
	c.B = uint8(float64(int(hcHi.Color.B)-int(hcLo.Color.B))*heat + float64(hcLo.Color.B))

	return c
}
