package price

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"io"
	"net/http"
	"sort"
	"time"
)

const (
	ApagaLuzAPIUrl       = "https://raw.githubusercontent.com/jorgeatgu/apaga-luz/main/public/data/today_price.json"
	ApagaLuzCanaryAPIUrl = "https://raw.githubusercontent.com/jorgeatgu/apaga-luz/main/public/data/canary_price.json"

	//
	dateLayout = "02/01/2006 15"
)

// Schedule represents a time range to start and stop an external device
type Schedule struct {
	Start time.Time
	Stop  time.Time
}

// HourData represents each of the individual items in the response retrieved from ApagaLuz API
type HourData struct {
	Day   string  `json:"day"`
	Hour  int     `json:"hour"`
	Price float64 `json:"price"`
	Zone  string  `json:"zone"`
}

// HourDataList represents the entire response retrieved from ApagaLuz API
type HourDataList []HourData

// GetApiData TODO DEVOLVER EL ERROR PA EVALUAR
func GetApiData(autoheater *v1alpha1.Autoheater) (response *HourDataList, err error) {

	// Weather not enabled, just throw empty data
	if !autoheater.Spec.Weather.Enabled {
		return response, nil
	}

	// Check fields regarding coordinates
	if autoheater.Spec.Weather.Coordinates.Latitude == 0 || autoheater.Spec.Weather.Coordinates.Longitude == 0 {
		return response, errors.New("coordinates section is required to evaluate weather")
	}

	// Send the request and wait for the result
	dataApiUrl := ApagaLuzAPIUrl
	if autoheater.Spec.Price.Zone == "canaryislands" {
		dataApiUrl = ApagaLuzCanaryAPIUrl
	}
	resp, err := http.Get(dataApiUrl)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	// Read the request's body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	// Decode response's JSON into a struct
	err = json.Unmarshal(body, &response)

	return response, err
}

// TODO
func GetApiDataByPrice(autoheater *v1alpha1.Autoheater) (response *HourDataList, err error) {

	response, err = GetApiData(autoheater)
	if err != nil {
		return response, err
	}

	sort.Slice(*response, func(i, j int) bool {
		return (*response)[i].Price < (*response)[j].Price
	})

	return response, nil
}

// GetCorrelativeHourRangesByPrice return a list of correlative hour ranges, sorted by price (from lowest to highest)
// This is what we want to achieve (best priced are always in the first ranges):
// [[hour_1, hour_2], [hour_5], [hour_7, hour_8]]
func GetCorrelativeHourRangesByPrice(autoheater *v1alpha1.Autoheater) (correlativeRanges []HourDataList, err error) {

	response, err := GetApiDataByPrice(autoheater)
	if err != nil {
		return correlativeRanges, err
	}

	correlativeRangesIndex := 0

	for index, item := range *response {

		// Add the first hour data to a new range directly
		if index == 0 {
			correlativeRanges = append(correlativeRanges, HourDataList{})
			correlativeRanges[correlativeRangesIndex] = append(correlativeRanges[correlativeRangesIndex], item)
			continue
		}

		// Get the previous element to compare if their hours are correlatives.
		// On correlatives, add current item to the same list of correlatives. If not, it's added in a new range
		previousItem := HourData{}
		previousItem = (*response)[index-1]

		if item.Hour == (previousItem.Hour+1) || item.Hour == (previousItem.Hour-1) {
			correlativeRanges[correlativeRangesIndex] = append(correlativeRanges[correlativeRangesIndex], item)
		} else {
			correlativeRangesIndex++
			correlativeRanges = append(correlativeRanges, HourDataList{})

			correlativeRanges[correlativeRangesIndex] = append(correlativeRanges[correlativeRangesIndex], item)
		}
	}

	return correlativeRanges, err

}

// GetLimitedCorrelativeHourRanges return as many correlative hour ranges (previously sorted by price) as needed
// to cover required active hours given by the config file
func GetLimitedCorrelativeHourRanges(autoheater *v1alpha1.Autoheater) (correlativeRanges []HourDataList, err error) {

	correlativeRangesByPrice, err := GetCorrelativeHourRangesByPrice(autoheater)
	if err != nil {
		return correlativeRanges, err
	}

	// Select as many hours as required by config
	requiredHoursLeft := autoheater.Spec.Price.ActiveHours

	for _, rangeItem := range correlativeRangesByPrice {

		// Current range can cover all required hours left
		if len(rangeItem) == requiredHoursLeft {
			correlativeRanges = append(correlativeRanges, rangeItem)
			break
		}

		if len(rangeItem) > requiredHoursLeft {
			correlativeRanges = append(correlativeRanges, rangeItem[0:requiredHoursLeft])
			break
		}

		correlativeRanges = append(correlativeRanges, rangeItem)
		requiredHoursLeft = requiredHoursLeft - len(rangeItem)
	}

	return correlativeRanges, err
}

// GetBestSchedules return a list of schedules that meet 'active hours' config parameter
// Starts are delayed by 5 minutes, and stops are 5 minutes early. Done in purpose to avoid time collisions on
// parallel scheduling. This can be improved a lot. Are you willing to contribute?
func GetBestSchedules(autoheater *v1alpha1.Autoheater) (schedules []Schedule, err error) {

	limitedCorrelativeRanges, err := GetLimitedCorrelativeHourRanges(autoheater)
	if err != nil {
		return schedules, err
	}

	for _, rangeItem := range limitedCorrelativeRanges {

		// Sort by hour (minor to major)
		sort.Slice(rangeItem, func(i, j int) bool {
			return rangeItem[i].Hour < rangeItem[j].Hour
		})

		// Get timestamp for the first item (always present)
		startTime, err := time.Parse(dateLayout, fmt.Sprintf("%s %d", rangeItem[0].Day, rangeItem[0].Hour))
		if err != nil {
			return schedules, err
		}

		startTime = startTime.Add(5 * time.Minute)

		var stopTime time.Time

		// Only one item: in 1 hour stop it
		if len(rangeItem) == 1 {
			stopTime = startTime.Add(50 * time.Minute)
			goto appendSchedule
		}

		// More items: stop it on last item's timestamp
		stopTime, err = time.Parse(dateLayout, fmt.Sprintf("%s %d", rangeItem[len(rangeItem)-1].Day, rangeItem[len(rangeItem)-1].Hour))
		if err != nil {
			return schedules, err
		}

		stopTime = stopTime.Add(55 * time.Minute)

		//
	appendSchedule:

		schedules = append(schedules, Schedule{
			Start: startTime,
			Stop:  stopTime,
		})
	}

	return schedules, err
}
