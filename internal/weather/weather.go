package weather

import (
	"encoding/json"
	"fmt"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"gonum.org/v1/gonum/stat"
)

const OpenMeteoAPIUrl = "https://api.open-meteo.com/v1/forecast"

// OpenMeteoResponseSpec represents the fields available to be sent to Open Meteo.
// DISCLAIMER: NOT all the fields are covered. Only those that are considered useful to know if it's hot
// Ref: https://open-meteo.com/en/docs
type OpenMeteoResponseSpec struct {

	// Geographical WGS84 coordinate of the location
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	GenerationTimeMs float64 `json:"generationtime_ms"`

	UTCOffsetSeconds float64 `json:"utc_offset_seconds"`

	Timezone             string `json:"timezone"`
	TimezoneAbbreviation string `json:"timezone_abbreviation"`

	Elevation float64 `json:"elevation"`

	HourlyUnits HourlyUnitsSpec `json:"hourly_units"`

	Hourly HourlySpec `json:"hourly"`
}

// TODO
type HourlyUnitsSpec struct {
	Time                string `json:"time"`
	Temperature2m       string `json:"temperature_2m"`
	ApparentTemperature string `json:"apparent_temperature"`
}

type HourlySpec struct {
	Time                []string  `json:"time"`
	Temperature2m       []float64 `json:"temperature_2m"`
	ApparentTemperature []float64 `json:"apparent_temperature"`
}

// TODO DEVOLVER EL ERROR PA EVALUAR
func GetApiData(autoheater *v1alpha1.Autoheater) (response *OpenMeteoResponseSpec, err error) {

	// Weather not enabled, just throw empty data
	if !autoheater.Spec.Weather.Enabled {
		return response, nil
	}

	// Check fields regarding coordinates
	if autoheater.Spec.Weather.Coordinates.Latitude == 0 || autoheater.Spec.Weather.Coordinates.Longitude == 0 {
		log.Fatal("coordinates section is required to evaluate weather")
	}

	// Check fields regarding temperature
	if autoheater.Spec.Weather.Temperature.Type == "" ||
		autoheater.Spec.Weather.Temperature.Unit == "" ||
		autoheater.Spec.Weather.Temperature.Threshold <= 0 {
		log.Fatal("Temperature section is required to evaluate weather")
	}

	// Select between apparent or real temperature
	parameterHourly := "apparent_temperature"
	if autoheater.Spec.Weather.Temperature.Type == "real" {
		parameterHourly = "temperature_2m"
	}

	// Select between celsius or fahrenheit
	parameterTemperatureUnit := "celsius"
	if autoheater.Spec.Weather.Temperature.Unit == "fahrenheit" {
		parameterTemperatureUnit = "fahrenheit"
	}

	// Convert floats values to string
	parameterLatitude := strconv.FormatFloat(autoheater.Spec.Weather.Coordinates.Latitude, 'g', 5, 64)
	parameterLongitude := strconv.FormatFloat(autoheater.Spec.Weather.Coordinates.Longitude, 'g', 5, 64)

	// Encode everything as URL
	params := url.Values{}
	params.Add("latitude", parameterLatitude)
	params.Add("longitude", parameterLongitude)
	params.Add("hourly", parameterHourly)
	params.Add("temperature_unit", parameterTemperatureUnit)

	requestUrl, err := url.Parse(OpenMeteoAPIUrl)
	if err != nil {
		fmt.Println("Error al analizar la URL base:", err)
		return
	}

	requestUrl.RawQuery = params.Encode()

	// Send the request and wait for the result
	resp, err := http.Get(requestUrl.String())
	if err != nil {
		fmt.Println("Error al hacer la solicitud HTTP:", err)
		return
	}
	defer resp.Body.Close()

	// Read the request's body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer el cuerpo de la respuesta:", err)
		return
	}

	// Decode response's JSON into a struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Print(err)
	}

	return response, nil
}

// TODO IsHeatingDay
func IsHeatingDay(autoheater *v1alpha1.Autoheater) (bool, error) {

	response, err := GetApiData(autoheater)
	if err != nil {
		return false, err
	}

	meanResult := stat.Mean(response.Hourly.ApparentTemperature, nil)
	if autoheater.Spec.Weather.Temperature.Type == "real" {
		meanResult = stat.Mean(response.Hourly.Temperature2m, nil)
	}

	if meanResult >= float64(autoheater.Spec.Weather.Temperature.Threshold) {
		return false, nil
	}

	return true, nil
}
