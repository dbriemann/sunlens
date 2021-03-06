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
	"github.com/dbriemann/sunlens/config"
	"github.com/dbriemann/sunlens/forecastio"
	"github.com/dbriemann/sunlens/utils"
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
	tm                time.Time
	feels             float64
	temp              float64
	precipIntensity   float64
	precipProbability float64
	precipType        string
	cloudCover        float64
}

// Terminal represents the basic type to render ascii weather
type Terminal struct {
	rows      int
	cols      int
	hours     int
	maxTemp   float64
	minTemp   float64
	tempRange int
	tempUnit  string
	days      []dayData
	forecast  *forecastio.Forecast
	conf      *config.Config
	// canvas represents the weather curve area
	canvas *ascii.Canvas
}

// NewTerminal creates a new terminal renderer with forecast data
func NewTerminal(fc *forecastio.Forecast) (*Terminal, error) {
	t := &Terminal{}
	// check terminal size
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
	switch t.forecast.Flags.Units {
	case "si":
		t.tempUnit = "C"
	case "us":
		t.tempUnit = "F"
	case "uk":
		t.tempUnit = "C"
	case "ca":
		t.tempUnit = "C"

	}
	// add spacing to right border and avoid linebreaks for certain widths
	t.cols--

	t.days = make([]dayData, 0)
	t.hours = (t.cols - leftSideBarWidth) / hourWidth
	if timeRange := len(t.forecast.Hourly.Data); timeRange < t.hours {
		t.hours = timeRange
	}

	t.maxTemp = -1000.0
	t.minTemp = 1000.0

	day := dayData{hourly: make([]hourData, 0), tm: nil}
	// get all data points for the presentable time interval
	for i := 0; i < t.hours; i++ {
		hData := t.forecast.Hourly.Data[i]
		tim := time.Unix(hData.Time, 0).Local()

		if tim.Hour() == 0 && len(day.hourly) > 0 {
			// the last day is over and is not an empty dummy object..
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
		day.hourly = append(day.hourly, hourData{
			tm:                tim,
			temp:              hData.Temperature,
			feels:             hData.ApparentTemperature,
			precipIntensity:   hData.PrecipIntensity,
			precipProbability: hData.PrecipProbability,
			precipType:        hData.PrecipType,
			cloudCover:        hData.CloudCover,
		})
	}

	// add the last non-finished day if it is no dummy..
	if len(day.hourly) > 0 {
		t.days = append(t.days, day)
	}

	// calculate range for temperature scale
	t.maxTemp = math.Round(t.maxTemp)
	//	if int(t.maxTemp)%2 != 0 {
	//		t.maxTemp++
	//	}

	t.minTemp = math.Round(t.minTemp)
	//	if int(t.minTemp)%2 != 0 {
	//		t.minTemp--
	//	}

	t.tempRange = int(t.maxTemp - t.minTemp + 1) // bounds inclusive

	t.canvas = ascii.NewCanvas(t.tempRange, t.hours*hourWidth)
}

// Render renders the weather forecast to the terminal
func (t *Terminal) Render() {
	dateLineTop := "\u250C%s\u2510"
	dateLineMiddle := "\u2502%s%s\u2502"
	dateLineBottom := "\u2514%s\u2518"
	headerTop := strings.Repeat(" ", leftSideBarWidth)
	headerMiddle := strings.Repeat(" ", leftSideBarWidth)
	headerBottom := strings.Repeat(" ", leftSideBarWidth)

	hourCount := 0
	var hours []int

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

		// build canvas with hours
		for _, hour := range day.hourly {
			scaleTemp := int(math.Round(hour.temp - t.minTemp))
			color := utils.NewColorByTemp(hour.temp, config.Settings.HeatMap, t.tempUnit)

			column := hourCount*hourWidth + hourWidth/2

			t.canvas.SetColor(scaleTemp, column, color)
			t.canvas.SetAnsi(scaleTemp, column, ascii.Bold)

			htemp := int(math.Round(hour.temp))
			hfeels := int(math.Round(hour.feels))

			if htemp > hfeels {
				t.canvas.Set(scaleTemp, column, '\u2533')
			} else if htemp < hfeels {
				t.canvas.Set(scaleTemp, column, '\u253B')
			} else {
				t.canvas.Set(scaleTemp, column, '\u2501') //\u2501 \u254B
			}

			// set weather indicators -> sunny, rainy, snowy, cloudy...
			// TODO -- put this in a function
			//			t.canvas.SetAnsi(0, column, ascii.Bold)
			//			if hour.precipProbability > 0.3 { // rainy / snowy
			//				if hour.precipType == "rain" {
			//					t.canvas.SetColor(0, column, utils.Color{R: 0, G: 5, B: 5})
			//					t.canvas.Set(0, column, '\u2614')
			//				} else { // freezy
			//					// snow, hail, ..
			//					t.canvas.SetColor(0, column, utils.Color{R: 5, G: 5, B: 5})
			//					t.canvas.Set(0, column, '*')
			//				}
			//			} else if hour.cloudCover >= 0.4 { // cloudy
			//				t.canvas.SetColor(0, column, utils.Color{R: 4, G: 4, B: 4})
			//				t.canvas.Set(0, column, '\u2601')
			//			} else { // sunny
			//				t.canvas.SetColor(0, column, utils.Color{R: 5, G: 5, B: 0})
			//				t.canvas.Set(0, column, '\u2600')
			//			}

			if hour.tm.Hour() == 0 || hourCount == 0 {
				// set vertical ..
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
		fmt.Printf("%3d°%s %s\n", i, t.tempUnit, t.canvas.Row(i-int(t.minTemp)))
		//		if i%2 == 0 {
		//			fmt.Printf("%3d°%s %s\n", i, t.tempUnit, t.canvas.Row(i-int(t.minTemp)))
		//		} else {
		//			fmt.Printf("%s%s\n", strings.Repeat(" ", leftSideBarWidth), t.canvas.Row(i-int(t.minTemp)))
		//		}
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
