package ascii

import "fmt"

const (
	Empty     = ' '                 //empty character == space
	FormatStr = "\033[%sm%s\033[0m" //ansi escape sequence -- parameters: 1.formatting 2.value -- closes by resetting formatting to default
	None      = "0"
)

var (
	baseHeatColors = []HeatColor{
		HeatColor{Temperature: -10, Color: Color{R: 0, G: 0, B: 5}}, //blue
		HeatColor{Temperature: 0, Color: Color{R: 0, G: 5, B: 5}},   //cyan
		HeatColor{Temperature: 10, Color: Color{R: 0, G: 5, B: 0}},  //green
		HeatColor{Temperature: 20, Color: Color{R: 5, G: 5, B: 0}},  //yellow
		HeatColor{Temperature: 30, Color: Color{R: 5, G: 0, B: 0}},  //red
	}
)

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

func NewColorByTemp(temp float64) Color {
	loColor := baseHeatColors[0]
	hiColor := baseHeatColors[len(baseHeatColors)-1]

	for i, col := range baseHeatColors {
		if temp <= col.Temperature {
			//set color col as upper bound
			hiColor = col
			//set last color as lower bound if possible
			if i > 0 {
				loColor = baseHeatColors[i-1]
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

//func HeatColor(heat float64) (r, g, b uint8) {
//	var lowR, lowG, lowB int8 = 0, 0, 5 //blue
//	var hiR, hiG, hiB int8 = 5, 0, 0    //red

//	r = uint8(int8(float64(hiR-lowR)*heat) + lowR)
//	g = uint8(int8(float64(hiG-lowG)*heat) + lowG)
//	b = uint8(int8(float64(hiB-lowB)*heat) + lowB)

//	return r, g, b
//}

/*
  	(r,0)       (r,c)
		+-----+
		|     |
		+-----+
  	(0,0)       (0,c)
*/
type Canvas struct {
	colMax     int
	rowMax     int
	values     [][]rune
	formatting [][]string
}

func NewCanvas(rows, cols int) *Canvas {
	c := &Canvas{}

	c.colMax = cols - 1
	c.rowMax = rows - 1
	c.values = make([][]rune, rows)
	c.formatting = make([][]string, rows)

	for row, _ := range c.values {
		c.values[row] = make([]rune, cols)
		c.formatting[row] = make([]string, cols)
		for col, _ := range c.values[row] {
			c.values[row][col] = Empty
			c.formatting[row][col] = None
		}
	}
	return c
}

//SetColor transforms "mini" RGB values (0 to 5) to terminal ansi code
//and sets the corresponding color for the specified coordinate
func (c *Canvas) SetColor(row, col int, color Color) { //uint8
	row = c.rowMax - row
	//sadly a lot of terminals don't support true colors yet..
	//in the future this would be a better alternative:
	//c.formatting[row][col] = fmt.Sprintf("38;2;%d;%d;%d;1", r, g, b) //+= ?
	//for now.. we use the basic(only) alternative
	number := 16 + 36*color.R + 6*color.G + color.B
	c.formatting[row][col] = fmt.Sprintf("38;5;%d", number) //+= ?
}

func (c *Canvas) SetVerticalBar(col int, ru rune) {
	for r := 0; r <= c.rowMax; r++ {
		c.SoftSet(r, col, ru)
	}
}

func (c *Canvas) Set(row, col int, ru rune) {
	row = c.rowMax - row
	c.values[row][col] = ru
}

func (c *Canvas) SoftSet(row, col int, ru rune) {
	row = c.rowMax - row
	if c.values[row][col] == Empty {
		c.values[row][col] = ru
	}
}

func (c *Canvas) Get(row, col int) rune {
	row = c.rowMax - row
	return c.values[row][col]
}

func (c *Canvas) Row(r int) string {
	r = c.rowMax - r
	return c.row(r)
	//	return string(c.values[r])
}

func (c *Canvas) row(r int) string {
	result := ""
	for i, ch := range c.values[r] {
		one := fmt.Sprintf(FormatStr, c.formatting[r][i], string(ch))
		//		fmt.Println(one)
		result += one
	}
	return result
}

//func (c Canvas) String() string {
//	str := ""
//	for _, row := range c.values {
//		str += string(row) + "\n"
//	}
//	return str
//}
