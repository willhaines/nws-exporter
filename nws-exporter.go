package main

import (
	"fmt"
	"strings"
	"regexp"
	"sort"
	"flag"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

type weather struct {
    Id string `json:'id'`
	Properties map[string]struct {
		UnitCode string `json:'unitCode'`
		Value float64 `json:'value'`
	} `json:'properties'`
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
func ToSnakeCase(str string) string {
    snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
    snake  = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
    return strings.ToLower(snake)
}

func main() {
	var station string
	var port int
	flag.StringVar(&station, "s", "KSKX", "Specify NWS station. Default is KSKX (Taos, NM).")
	flag.IntVar(&port, "p", 2112, "Specify port to listen on. Default is 2112.")
	flag.Parse()

    http.HandleFunc("/metrics", func (w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://api.weather.gov/stations/%s/observations/latest?require_qc=true", station)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		my_weather := weather{}
		err = json.Unmarshal(body, &my_weather)
		if err != nil {
			fmt.Println(err)
		}

		keys := make([]string, 0, len(my_weather.Properties))
		for k := range my_weather.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, prop := range keys {
			val := my_weather.Properties[prop]
			if val.UnitCode != "" {
				unit := strings.Split(val.UnitCode, ":")[1]
				property := ToSnakeCase(prop)
				fmt.Fprintf(w, "nws_%s{unit=\"%s\", station=\"%s\"} %f\n", property, unit, station, val.Value)
			}
		}
    })
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}