package forecastio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

//various constant values...
const (
	//QueryURL format: https://api.forecast.io/forecast/APIKEY/LATITUDE,LONGITUDE[,TIME]?parameters..
	QueryURL = "https://api.darksky.net/forecast/%s/%.5f,%.5f?units=%s&lang=%s"

	CA   string = "ca"
	SI   string = "si"
	US   string = "us"
	UK   string = "uk"
	AUTO string = "auto"
)

//DataPoint represents the various weather phenomena occurring at a specific instant of time
type DataPoint struct {
	Time        int64  `json:"time"`
	Summary     string `json:"summary"`
	Code        string `json:"icon"`
	SunriseTime int64  `json:"sunriseTime"`
	SunsetTime  int64  `json:"sunsetTime"`

	//fractional part of the moon area showing: 0 == new moon, 0.25 == first quarter moon, 0.5 == full moon, 0.75 == last quarter moon
	//(The ranges in between these represent waxing crescent, waxing gibbous, waning gibbous, and waning crescent moons, respectively.)
	MoonPhase           float64 `json:"moonPhase,omitempty"`
	NearestStormDist    float64 `json:"nearestStormDistance,omitempty"`
	NearestStormBearing float64 `json:"nearestStormBearing,omitempty"`

	/*
		A very rough guide is that a value of
		0 in./hr. corresponds to no precipitation,
		0.002 in./hr. corresponds to very light precipitation,
		0.017 in./hr. corresponds to light precipitation,
		0.1 in./hr. corresponds to moderate precipitation,
		0.4 in./hr. corresponds to heavy precipitation.
	*/
	PrecipIntensity        float64 `json:"precipIntensity"`
	PrecipIntensityMax     float64 `json:"precipIntensityMax,omitempty"`
	PrecipIntensityMaxTime float64 `json:"precipIntensityMaxTime,omitempty"`
	PrecipProbability      float64 `json:"precipProbability"`

	//rain, snow, sleet (which applies to each of freezing rain, ice pellets, and “wintery mix”), or hail
	PrecipType string `json:"precipType,omitempty"`

	//snowfall accumulation
	PrecipAccumulation  float64 `json:"precipAccumulation,omitempty"`
	Temperature         float64 `json:"temperature,omitempty"`
	TemperatureMin      float64 `json:"temperatureMin,omitempty"`
	TemperatureMinTime  float64 `json:"temperatureMinTime,omitempty"`
	TemperatureMax      float64 `json:"temperatureMax,omitempty"`
	TemperatureMaxTime  float64 `json:"temperatureMaxTime,omitempty"`
	ApparentTemperature float64 `json:"apparentTemperature,omitempty"`
	DewPoint            float64 `json:"dewPoint"`
	WindSpeed           float64 `json:"windSpeed"`
	WindBearing         float64 `json:"windBearing,omitempty"`

	//0 corresponds to clear sky, 0.4 to scattered clouds, 0.75 to broken cloud cover, and 1 to completely overcast skies.
	CloudCover float64 `json:"cloudCover"`
	Humidity   float64 `json:"humidity"` //0.0 - 1.0
	Pressure   float64 `json:"pressure"`
	Visibility float64 `json:"visibility"`
	Ozone      float64 `json:"ozone"`
}

//DataBlock represents the various weather phenomena occurring over a period of time.
type DataBlock struct {
	Summary string      `json:"summary"`
	Code    string      `json:"icon"`
	Data    []DataPoint `json:"data"`
}

//Alert represents a sever weather warning
type Alert struct {
	Title       string `json:"title"`
	Expires     int64  `json:"expires"`
	Description string `json:"description"`
	URL         string `json:"uri"`
}

//Flags contains various metadata information related to the request
type Flags struct {
	DarkSkyUnavailable string   `json:"darksky-unavailable"`
	DarkSkyStations    []string `json:"darksky-stations"`
	DataPointStations  []string `json:"datapoint-stations"`
	IsdStations        []string `json:"isd-stations"`
	LampStations       []string `json:"lamp-stations"`
	MetarStations      []string `json:"metar-stations"`
	MetnoLicense       string   `json:"metno-license"`
	Sources            []string `json:"sources"`
	Units              string   `json:"units"`
}

//Forecast contains all data of a detailed weather forecast
type Forecast struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Offset    int       `json:"offset"`
	Currently DataPoint `json:"currently"`
	Minutely  DataBlock `json:"minutely"`
	Hourly    DataBlock `json:"hourly"`
	Daily     DataBlock `json:"daily"`
	Alerts    []Alert   `json:"alerts"`
	Flags     Flags     `json:"flags"`
}

//GetForecast queries the forecast.io server and returns a Forecast object
func GetForecast(key string, lat, lng float64, unitType, lang string) (*Forecast, error) {
	fc := &Forecast{}

	//build query url
	url := fmt.Sprintf(QueryURL, key, lat, lng, unitType, lang)
	//	fmt.Println(url)

	//http get forecast
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Problem talking to forecast.io API: " + err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("Problem handling forecast.io data: " + err.Error())
	}

	//decode json response
	if err = json.Unmarshal(body, fc); err != nil {
		return nil, errors.New("Problem parsing forecast.io API response: " + err.Error())
	}

	return fc, nil
}
