package ascii

const (
	Empty = ' '
)

/*
   (r,0)       (r,c)
		+-----+
		|     |
		+-----+
   (0,0)       (0,c)
*/
type Canvas struct {
	colMax int
	rowMax int
	values [][]rune
}

func NewCanvas(rows, cols int) *Canvas {
	c := &Canvas{}

	c.colMax = cols - 1
	c.rowMax = rows - 1
	c.values = make([][]rune, rows)

	for row, _ := range c.values {
		c.values[row] = make([]rune, cols)
		for col, _ := range c.values[row] {
			c.values[row][col] = Empty
		}
	}
	return c
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
	return string(c.values[r])
}

func (c Canvas) String() string {
	str := ""
	for _, row := range c.values {
		str += string(row) + "\n"
	}
	return str
}
