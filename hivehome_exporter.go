package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DevOpsFu/go-hivehome/hivehome"
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

func init() {

	viper.SetDefault("server.address", "")
	viper.SetDefault("server.port", "8000")

	viper.SetDefault("metrics.collect_interval", "10s")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/hivehome_exporter/")
	viper.AddConfigPath("$HOME/.hivehome_exporter/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s", err))
	}

	viper.WatchConfig()
}

func main() {

	go getMetrics()

	endpoint := viper.GetString("server.address") + ":" + viper.GetString("server.port")

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(endpoint, nil))
}

func getMetrics() {
	username := viper.GetString("credentials.username")
	password := viper.GetString("credentials.password")
	client := hivehome.NewClient(username, password)

	client.Login()

	thermostatZone := viper.GetString("metrics.thermostat_zone")
	thermostatID := client.GetThermostatIDForZone(thermostatZone)
	collectInterval := viper.GetString("metrics.collect_interval")

	collectionTimerDuration, _ := time.ParseDuration(collectInterval)

	for range time.Tick(collectionTimerDuration) {

		resp := client.GetNodeAttributes(thermostatID)

		currentTemp.Set(gjson.Get(resp, "temperature.reportedValue").Float())
		targetTemp.Set(gjson.Get(resp, "targetHeatTemperature.reportedValue").Float())

		switch gjson.Get(resp, "stateHeatingRelay.reportedValue").String() {
		case "OFF":
			heatingStatus.Set(0)
		case "ON":
			heatingStatus.Set(1)
		}
	}

}
