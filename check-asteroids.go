package main

import (
	"encoding/json"
	"fmt"
	"github.com/yieldbot/sensuplugin/sensuutil"
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"net/http"
)

var ok = "ok"
var warning = "warning"
var critical = "critical"

type Asteroid struct {
	Name                              string `json:"name"`
	Nasa_Jpl_Url                      string `json:"nasa_jpl_url"`
	Is_Potentially_Hazardous_Asteroid bool   `json:"is_potentially_hazardous_asteroid"`
}

type Data struct {
	Element_Count      int `json:"element_count"`
	Near_Earth_Objects map[string]*json.RawMessage
}

func main() {
	type Config struct {
		Credentials struct {
			Apikey string
		}
	}

	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "config.cfg")
	if err != nil {
		sensuutil.Exit(critical, fmt.Sprintf("Failed to parse gcfg data: %s", err))
	}

	var api_key = cfg.Credentials.Apikey
	url := "https://api.nasa.gov/neo/rest/v1/feed/today?api_key="
	url += api_key

	// conect to Nasa API
	resp, err := http.Get(url)
	if err != nil {
		sensuutil.Exit(warning, (fmt.Sprintf("Error: %s", err)))
	}
	defer resp.Body.Close()

	// confirm that we received an OK status
	if resp.StatusCode != http.StatusOK {
		sensuutil.Exit(warning, (fmt.Sprintf("Error: %s", resp.StatusCode)))
	}

	// read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sensuutil.Exit(warning, (fmt.Sprintf("Error reading body: %s", err)))
	}

	var entries Data
	if err := json.Unmarshal(body, &entries); err != nil {
		sensuutil.Exit(warning, (fmt.Sprintf("Error decoding JSON: %s", err)))
	}

	for _, asteroids := range entries.Near_Earth_Objects {
		as := make([]Asteroid, 0)
		json.Unmarshal(*asteroids, &as)

		for _, a := range as {
			sensuutil.Exit(checkIsHazardous(a))
		}
	}
}

func checkIsHazardous(asteroid Asteroid) (string, string) {
	if asteroid.Is_Potentially_Hazardous_Asteroid == true {
		msg := "Potentially Hazardous Asteroid detected!" + asteroid.Nasa_Jpl_Url
		return critical, msg
	}
	return ok, "No Hazardous Asteroids today."
}
