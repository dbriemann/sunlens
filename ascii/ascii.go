package ascii

import (
	"fmt"

	"github.com/dbriemann/sunlens/utils"
)

const (
	Empty     = ' '
	FormatStr = "\033[%sm%s\033[0m" //ansi escape sequence -- parameters: 1.formatting 2.value -- closes by resetting formatting to default
	None      = "0"
	Bold      = "1"
	Blink     = "5"
)

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

func (c *Canvas) SetAnsi(row, col int, str string) {
	row = c.rowMax - row
	if c.formatting[row][col] == None {
		c.formatting[row][col] = str
	} else {
		c.formatting[row][col] += ";" + str
	}
}

//SetColor transforms "mini" RGB values (0 to 5) to terminal ansi code
//and sets the corresponding color for the specified coordinate
func (c *Canvas) SetColor(row, col int, color utils.Color) { //uint8
	//sadly a lot of terminals don't support true colors yet..
	//in the future this would be a better alternative:
	//c.formatting[row][col] = fmt.Sprintf("38;2;%d;%d;%d;1", r, g, b) //+= ?
	//for now.. we use the basic(only) alternative
	number := 16 + 36*color.R + 6*color.G + color.B
	c.SetAnsi(row, col, fmt.Sprintf("38;5;%d", number))
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
