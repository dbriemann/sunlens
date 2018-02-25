package main

import (
	"os"
	"os/user"
	"path"

	"fmt"

	"github.com/dbriemann/geopard"
	"github.com/dbriemann/sunlens/config"
	"github.com/dbriemann/sunlens/forecastio"
	"github.com/dbriemann/sunlens/terminal"
)

const (
	configExtPath  = ".config/sunlens/"
	configFileName = "sunlens.cfg"
)

var (
	usrHome string
)

func init() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Could not detect user: " + err.Error())
		os.Exit(0)
	} else {
		usrHome = usr.HomeDir
	}

	//create config path directory if it does not exist yet
	basePath := path.Join(usr.HomeDir, configExtPath)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Println("Unable to create config directory: " + basePath + " Error: " + err.Error())
		os.Exit(0)
	}
}

func main() {
	//TODO
	//1. load or create config without location
	//2. get argument
	//3. is shortcut/location/nothing -> load/create/default location
	//4. x/add/add to config
	//parse command line arguments
	var loc config.Location
	if len(os.Args) > 1 {
		//take first argument if there is one, ignore all following
		locArg := os.Args[1]

		location, err := config.NewLocation(locArg)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		loc = location
	}

	//load config from file or create default config if none exists yet
	configPath := path.Join(usrHome, configExtPath, configFileName)
	conf, err := config.LoadConfig(configPath, loc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	config.Settings = conf

	//if location is not set use default location from config
	if loc.Shortcut == "" {
		loc = conf.Locations[conf.DefaultLocation]
	} else {
		location := loc
		//check if location already exists in config
		for _, c := range conf.Locations {
			if loc.Shortcut == c.Shortcut {
				location = c
			}
		}

		//and save a new location in the config file
		if location.Latitude != loc.Latitude && location.Longitude != loc.Longitude {
			fmt.Println("Saving new location: ", location)
			loc = location
		}
	}

	//request forecast data from forecast.io
	fc, err := forecastio.GetForecast(conf.ApiKey, loc.Latitude, loc.Longitude, conf.UnitFormat, conf.Language)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	//create terminal to render data in ascii
	term, err := terminal.NewTerminal(fc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Printf(" Weather for: %s [shortcut:%s]\n", loc.City, loc.Shortcut)

	term.Render()

	geopard.GetInstance().Destroy()
}
