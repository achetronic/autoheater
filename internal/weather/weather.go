package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"gonum.org/v1/gonum/stat"
)

const (
	OpenMeteoAPIUrl = "https://api.open-meteo.com/v1/forecast"

	//
	CoordinatesNotFoundErrorMessage    = "coordinates section is required to evaluate weather"
	TemperatureNotFoundErrorMessage    = "temperature section is required to evaluate weather"
	HttpUrlParsingErrorMessage         = "error configuring http request: %s"
	HttpRequestFailedErrorMessage      = "error performing http request: %s"
	HttpResponseReadFailedErrorMessage = "error reading http response's body: %s"
)

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

// HourlyUnitsSpec TODO
type HourlyUnitsSpec struct {
	Time                string `json:"time"`
	Temperature2m       string `json:"temperature_2m"`
	ApparentTemperature string `json:"apparent_temperature"`
}

// HourlySpec TODO
type HourlySpec struct {
	Time                []string  `json:"time"`
	Temperature2m       []float64 `json:"temperature_2m"`
	ApparentTemperature []float64 `json:"apparent_temperature"`
}

// GetApiData TODO
func GetApiData(ctx *v1alpha1.Context) (response *OpenMeteoResponseSpec, err error) {

	// Weather not enabled, just throw empty data
	if !ctx.Config.Spec.Weather.Enabled {
		return response, nil
	}

	// Check fields regarding coordinates
	if ctx.Config.Spec.Weather.Coordinates.Latitude == 0 || ctx.Config.Spec.Weather.Coordinates.Longitude == 0 {
		ctx.Logger.Fatal(CoordinatesNotFoundErrorMessage)
	}

	// Check fields regarding temperature
	if ctx.Config.Spec.Weather.Temperature.Type == "" ||
		ctx.Config.Spec.Weather.Temperature.Unit == "" ||
		ctx.Config.Spec.Weather.Temperature.Threshold <= 0 {
		ctx.Logger.Fatal(TemperatureNotFoundErrorMessage)
	}

	// Select between apparent or real temperature
	parameterHourly := "apparent_temperature"
	if ctx.Config.Spec.Weather.Temperature.Type == "real" {
		parameterHourly = "temperature_2m"
	}

	// Select between celsius or fahrenheit
	parameterTemperatureUnit := "celsius"
	if ctx.Config.Spec.Weather.Temperature.Unit == "fahrenheit" {
		parameterTemperatureUnit = "fahrenheit"
	}

	// Convert floats values to string
	parameterLatitude := strconv.FormatFloat(ctx.Config.Spec.Weather.Coordinates.Latitude, 'g', 5, 64)
	parameterLongitude := strconv.FormatFloat(ctx.Config.Spec.Weather.Coordinates.Longitude, 'g', 5, 64)

	// Encode everything as URL
	params := url.Values{}
	params.Add("latitude", parameterLatitude)
	params.Add("longitude", parameterLongitude)
	params.Add("hourly", parameterHourly)
	params.Add("temperature_unit", parameterTemperatureUnit)

	requestUrl, err := url.Parse(OpenMeteoAPIUrl)
	if err != nil {
		return response, errors.New(fmt.Sprintf(HttpUrlParsingErrorMessage, err))
	}

	requestUrl.RawQuery = params.Encode()

	// Send the request and wait for the result
	resp, err := http.Get(requestUrl.String())
	if err != nil {
		return response, errors.New(fmt.Sprintf(HttpRequestFailedErrorMessage, err))
	}
	defer resp.Body.Close()

	// Read the request's body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, errors.New(fmt.Sprintf(HttpResponseReadFailedErrorMessage, err))
	}

	// Decode response's JSON into a struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Print(err)
	}

	return response, nil
}

// IsColdDay return true when temperature's mean for the whole day is under the threshold defined on config
func IsColdDay(ctx *v1alpha1.Context) (bool, error) {

	response, err := GetApiData(ctx)
	if err != nil {
		return false, err
	}

	meanResult := stat.Mean(response.Hourly.ApparentTemperature, nil)
	if ctx.Config.Spec.Weather.Temperature.Type == "real" {
		meanResult = stat.Mean(response.Hourly.Temperature2m, nil)
	}

	if meanResult >= float64(ctx.Config.Spec.Weather.Temperature.Threshold) {
		return false, nil
	}

	return true, nil
}
