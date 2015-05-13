package main

import (
	"fmt"
	"os"
	"os/user"
	"path"

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
	//load config from file or create default config if none exists yet
	configPath := path.Join(usrHome, configExtPath, configFileName)
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	//request forecast data from forecast.io
	location := conf.Locations[conf.DefaultLocation]
	fc, err := forecastio.GetForecast(conf.ApiKey, location.Latitude, location.Longitude, conf.UnitFormat, conf.Language)
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

	fmt.Printf("Weather for: %s [shortcut:#%s]\n", location.City, location.Shortcut)

	term.Render()
}
