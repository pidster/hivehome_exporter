package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DevOpsFu/go-hivehome/hivehome"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

var (
	currentTemp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hivehome_thermostat_temperature_current",
			Help: "The current temperature reported by the thermostat",
		})

	targetTemp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hivehome_thermostat_temperature_target",
			Help: "The target temperature set on the thermostat",
		})

	heatingStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "hivehome_thermostat_heating_status",
			Help: "Reports whether the heating relay is on or off",
		})
)

var client *hivehome.Client
var thermostatZone string

func init() {

	viper.SetDefault("server.address", "")
	viper.SetDefault("server.port", "8000")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/hivehome_exporter/")
	viper.AddConfigPath("$HOME/.hivehome_exporter/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s", err))
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config changed")
	})
	username := viper.GetString("credentials.username")
	password := viper.GetString("credentials.password")
	thermostatZone = viper.GetString("metrics.thermostat_zone")
	client = hivehome.NewClient(username, password)
}

func main() {
	endpoint := viper.GetString("server.address") + ":" + viper.GetString("server.port")

	http.HandleFunc("/metrics", getMetricsHandler)

	log.Printf("Starting Hivehome exporter listening at %v", endpoint)
	panic(http.ListenAndServe(endpoint, nil))
}

func getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	getMetrics()
	promhttp.Handler().ServeHTTP(w, r)
}

func getMetrics() {
	log.Println("Retrieving metrics...")

	thermostatID, err := client.GetThermostatIDForZone(thermostatZone)

	if err != nil {
		panic(fmt.Errorf("Fatal error getting Thermostat ID: %s", err))
	}

	resp, err := client.GetNodeAttributes(thermostatID)

	if err != nil {
		panic(fmt.Errorf("Fatal error getting node attributes: %s", err))
	}

	currentTemp.Set(gjson.Get(resp, "temperature.reportedValue").Float())
	targetTemp.Set(gjson.Get(resp, "targetHeatTemperature.reportedValue").Float())

	switch gjson.Get(resp, "stateHeatingRelay.reportedValue").String() {
	case "OFF":
		heatingStatus.Set(0)
	case "ON":
		heatingStatus.Set(1)
	}

	log.Println("Retrieval complete")
}
