package terminal

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dbriemann/sunlens/ascii"
	"github.com/dbriemann/sunlens/forecastio"
)

const (
	terminalMinCols  = 80
	terminalMinRows  = 24
	leftSideBarWidth = 6
	hourWidth        = 4
)

type dayData struct {
	hourly []hourData
	tm     *time.Time
}

type hourData struct {
	tm    time.Time
	feels float64
	temp  float64
}

//Terminal represents the basic type to render ascii weather
type Terminal struct {
	rows      int
	cols      int
	hours     int
	maxTemp   float64
	minTemp   float64
	tempRange int
	days      []dayData
	forecast  *forecastio.Forecast
	//canvas represents the weather curve area
	canvas *ascii.Canvas
}

//NewTerminal creates a new terminal renderer with forecast data
func NewTerminal(fc *forecastio.Forecast) (*Terminal, error) {
	t := &Terminal{}
	//check terminal size
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	resp, err := cmd.Output()
	if err != nil {
		return nil, errors.New("Cannot detect terminal: " + err.Error())
	}

	size := strings.Fields(string(resp))
	if t.rows, err = strconv.Atoi(size[0]); err != nil {
		return nil, errors.New("Cannot read terminal size: " + err.Error())
	}
	if t.rows < terminalMinRows {
		return nil, errors.New("Terminal is to small: Number of rows must be at least " + strconv.Itoa(terminalMinRows))
	}
	if t.cols, err = strconv.Atoi(size[1]); err != nil {
		return nil, errors.New("Cannot read terminal size: " + err.Error())
	}
	if t.cols < terminalMinCols {
		return nil, errors.New("Terminal is to small: Number of columns must be at least " + strconv.Itoa(terminalMinCols))
	}

	t.forecast = fc
	t.init()

	return t, nil
}

func (t *Terminal) init() {
	t.days = make([]dayData, 0)
	t.hours = (t.cols - leftSideBarWidth) / hourWidth
	if timeRange := len(t.forecast.Hourly.Data); timeRange < t.hours {
		t.hours = timeRange
	}

	t.maxTemp = -1000.0
	t.minTemp = 1000.0

	day := dayData{hourly: make([]hourData, 0), tm: nil}
	//get all data points for the presentable time interval
	for i := 0; i < t.hours; i++ {
		hData := t.forecast.Hourly.Data[i]
		tim := time.Unix(hData.Time, 0).Local()

		if tim.Hour() == 0 && len(day.hourly) > 0 {
			//the last day is over and is not an empty dummy object..
			t.days = append(t.days, day)
			day = dayData{hourly: make([]hourData, 0), tm: nil} //create a new day
		}

		if day.tm == nil {
			day.tm = &tim
		}

		if hData.Temperature > t.maxTemp {
			t.maxTemp = hData.Temperature
		}
		if hData.Temperature < t.minTemp {
			t.minTemp = hData.Temperature
		}
		day.hourly = append(day.hourly, hourData{tm: tim, temp: hData.Temperature, feels: hData.ApparentTemperature})
	}
	//add the last non-finished day if it is no dummy..
	if len(day.hourly) > 0 {
		t.days = append(t.days, day)
	}

	//calculate range for temperature scale
	t.maxTemp = math.Ceil(t.maxTemp)
	absmod := ((int(t.maxTemp) % 2) + 2) % 2 //double modulo to force positive result
	oe := 2 - absmod                         //1 if maxTemp is odd and 2 if maxTemp is even
	t.maxTemp += float64(oe)

	t.minTemp = math.Floor(t.minTemp)
	absmod = ((int(t.minTemp) % 2) + 2) % 2 //double modulo to force positive result
	oe = 2 - absmod                         //1 if minTemp is odd and 2 if minTemp is even
	t.minTemp -= float64(oe)

	t.tempRange = int(t.maxTemp - t.minTemp + 1) //bounds inclusive

	t.canvas = ascii.NewCanvas(t.tempRange, t.hours*hourWidth)
}

//Render renders the weather forecast to the terminal
func (t *Terminal) Render() {
	dateLineTop := "\u250C%s\u2510"
	dateLineMiddle := "\u2502%s%s\u2502"
	dateLineBottom := "\u2514%s\u2518"
	headerTop := strings.Repeat(" ", leftSideBarWidth)
	headerMiddle := strings.Repeat(" ", leftSideBarWidth)
	headerBottom := strings.Repeat(" ", leftSideBarWidth)

	hourCount := 0
	hours := make([]int, 0)

	for _, day := range t.days {
		hoursLeft := len(day.hourly)
		totalChars := hoursLeft * hourWidth

		dayDesc := day.tm.Weekday().String()
		descs := []string{
			dayDesc + ", " + day.tm.Month().String() + " " + strconv.Itoa(day.tm.Day()),
			dayDesc,
			dayDesc[0:3],
			dayDesc[0 : hourWidth-2],
		}
		dayDesc = ""
		for _, desc := range descs {
			if len(desc) <= totalChars-2 {
				dayDesc = desc
				break
			}
		}

		topFill := strings.Repeat("\u2500", totalChars-2)
		bottomFill := strings.Repeat("\u2500", totalChars-2)
		middleFill := strings.Repeat(" ", totalChars-2-len(dayDesc))

		headerTop += fmt.Sprintf(dateLineTop, topFill)
		headerMiddle += fmt.Sprintf(dateLineMiddle, dayDesc, middleFill)
		headerBottom += fmt.Sprintf(dateLineBottom, bottomFill)

		//build canvas with hours
		for _, hour := range day.hourly {
			scaleTemp := int(hour.temp - t.minTemp)
			color := ascii.NewColorByTemp(hour.temp)
			t.canvas.SetColor(scaleTemp, hourCount*hourWidth+hourWidth/2, color)
			if math.Floor(hour.temp+0.5) > math.Floor(hour.feels+0.5) {
				t.canvas.Set(scaleTemp, hourCount*hourWidth+hourWidth/2, '\u2533')
			} else if math.Floor(hour.temp+0.5) < math.Floor(hour.feels+0.5) {
				t.canvas.Set(scaleTemp, hourCount*hourWidth+hourWidth/2, '\u253B')
			} else {
				t.canvas.Set(scaleTemp, hourCount*hourWidth+hourWidth/2, '\u2501') //\u2501 \u254B
			}

			if hour.tm.Hour() == 0 || hourCount == 0 {
				//set vertical ..
				t.canvas.SetVerticalBar(hourCount*hourWidth, '\u2502')
			}

			hourCount++

			hours = append(hours, hour.tm.Hour())
		}
	}

	fmt.Println(headerTop)
	fmt.Println(headerMiddle)
	fmt.Println(headerBottom)

	for i := int(t.maxTemp); i >= int(t.minTemp); i-- {
		if i%2 == 0 {
			fmt.Printf("%3d°C %s\n", i, t.canvas.Row(i-int(t.minTemp)))
		} else {
			fmt.Printf("%s%s\n", strings.Repeat(" ", leftSideBarWidth), t.canvas.Row(i-int(t.minTemp)))
		}
	}

	outerScale := strings.Repeat(" ", leftSideBarWidth) +
		fmt.Sprintf("\u2514%s%s\u2534%s\u2518", strings.Repeat("\u2500", hourWidth-1), "%s", strings.Repeat("\u2500", hourWidth-1))
	innerScale := strings.Repeat(fmt.Sprintf("\u2534%s", strings.Repeat("\u2500", hourWidth-1)), hourCount-2)
	hourScale := fmt.Sprintf(outerScale, innerScale)
	fmt.Println(hourScale)
	fmt.Printf(strings.Repeat(" ", leftSideBarWidth))
	for _, h := range hours {
		fmt.Printf("%02d%s", h, strings.Repeat(" ", hourWidth-2))
	}
	fmt.Printf("\n")
}
