package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dbriemann/sunlens/utils"
	"github.com/jasonmoo/geo"
)

//global settings
var (
	Settings *Config
)

type Location struct {
	City      string
	Latitude  float64
	Longitude float64
	Shortcut  string
}

//NewLocation creates a location from its shortcut description
//by querying the google api for lat, long and exact name
func NewLocation(shortcut string) (Location, error) {
	loc := Location{Shortcut: shortcut}

	//find geo coordinates
	add, err := geo.Geocode(shortcut)
	if err != nil {
		return loc, errors.New("Unable to get latitude and longitude for: " + shortcut + " Error: " + err.Error())
	}

	loc.City = add.Address
	loc.Latitude = add.Lat
	loc.Longitude = add.Lng

	return loc, nil
}

//Config stores all basic settings the user should adjust.
type Config struct {
	ApiKey string
	//	City             string
	//	Latitude         float64
	//	Longitude        float64
	UnitFormat string //us(farenheit, miles..), si(celsius, meters..), ca, uk, auto(location dependent)
	HeatMap    []utils.HeatColor
	/*
		language may be one of the following:
		bs (Bosnian), de (German), en (English, which is the default),
		es (Spanish), fr (French), it (Italian), nl (Dutch), pl (Polish),
		pt (Portuguese), ru (Russian), tet (Tetum), or x-pig-latin (Igpay Atinlay)
	*/
	Language        string     //language code.. see above
	DefaultLocation int        //sets the number of the default location in "Location" slice
	Locations       []Location //saves all queried locations
}

//LoadConfig creates a new Config object from a json file.
func LoadConfig(path string) (*Config, error) {
	c := &Config{}
	b, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &c)
		if err == nil {
			if c.ApiKey == "" {
				return nil, errors.New("Please set your ApiKey in config file.")
			}
		}
		return c, err
	}
	if _, ok := err.(*os.PathError); ok {
		//default fallback values if no config is found
		c.ApiKey = ""
		c.UnitFormat = "auto"
		c.Language = "en"
		c.HeatMap = []utils.HeatColor{
			utils.HeatColor{Temperature: -10, Color: utils.Color{R: 0, G: 0, B: 5}}, //blue
			utils.HeatColor{Temperature: 0, Color: utils.Color{R: 0, G: 5, B: 5}},   //cyan
			utils.HeatColor{Temperature: 10, Color: utils.Color{R: 0, G: 5, B: 0}},  //green
			utils.HeatColor{Temperature: 20, Color: utils.Color{R: 5, G: 5, B: 0}},  //yellow
			utils.HeatColor{Temperature: 30, Color: utils.Color{R: 5, G: 0, B: 0}},  //red
		}

		c.Locations = make([]Location, 1)
		if c.Locations[0], err = NewLocation("Darmstadt"); err != nil {
			return nil, errors.New("Could not create default location: " + err.Error())
		}

		if err := c.Save(path); err != nil {
			return nil, errors.New("Could not save default config: " + err.Error())
		} else {
			return nil, errors.New(fmt.Sprintf("Created default config file: %s\n%s", path, "Please edit it and change the parameters."))
		}
	} else if err != nil {
		return nil, errors.New("Error in config file: " + path + " : " + err.Error())
	}
	return c, err
}

//Save saves the Config object c to a json file.
func (c *Config) Save(path string) error {
	j, err := json.MarshalIndent(c, "", "\t")
	if err == nil {
		return ioutil.WriteFile(path, j, 0600)
	}
	return err
}
