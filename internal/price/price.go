// ATTENTION:
// This package uses ApagaLuz as a datasource for PVPC. This datasource uses official ESIOS' API from the spanish
// government, which return data using CEST timezone. As a consequence, all this package works using CEST timezone
// and the other packages must take this into account to convert it and do whatever they need.
// [ESIOS] Ref: https://www.esios.ree.es/es/pvpc

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

	// Location of retrieved data from ApagaLuz
	ApagaLuzApiTimeLocation = "Europe/Madrid"

	//
	dateLayout = "02/01/2006 15"

	//
	ActiveHoursOutOfRangeErrorMessage = "config.device.activeHours field must be a number between 1 and 24"
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

// GetApiData TODO
func GetApiData(ctx *v1alpha1.Context) (response *HourDataList, err error) {

	// Send the request and wait for the result
	dataApiUrl := ApagaLuzAPIUrl
	if ctx.Config.Spec.Price.Zone == "canaryislands" {
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

	// Discard passed hours when requested by config
	if ctx.Config.Spec.Global.IgnorePassedHours {
		currentHour := time.Now().In(time.Local).Hour()
		*response = (*response)[currentHour:24]
	}

	return response, err
}

// TODO
func GetApiDataByPrice(ctx *v1alpha1.Context) (response *HourDataList, err error) {

	response, err = GetApiData(ctx)
	if err != nil {
		return response, err
	}

	sort.Slice(*response, func(i, j int) bool {
		return (*response)[i].Price < (*response)[j].Price
	})

	return response, nil
}

// GetLimitedCorrelativeHourRanges return an array whose elements are lists of correlative hours.
// Those hours were previously sorted and selected by having the lowest price as criteria
func GetLimitedCorrelativeHourRanges(ctx *v1alpha1.Context) (correlativeRanges []HourDataList, err error) {

	// 1. Get all the data sorted by price
	response, err := GetApiDataByPrice(ctx)
	if err != nil {
		return correlativeRanges, err
	}

	// Check desired amount of hours. Must be between 0 than 24
	if ctx.Config.Spec.Device.ActiveHours < 1 || ctx.Config.Spec.Device.ActiveHours > 24 {
		return correlativeRanges, errors.New(ActiveHoursOutOfRangeErrorMessage)
	}

	// 2. Keep the N cheapest items, discard the others
	*response = (*response)[0:ctx.Config.Spec.Device.ActiveHours]

	// 3. Re-sort them by hour
	sort.Slice(*response, func(i, j int) bool {
		return (*response)[i].Hour < (*response)[j].Hour
	})

	// 4. Craft an array whose elements are lists of correlative hours
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

// GetBestSchedules return a list of schedules that meet 'active hours' config parameter
// Starts are delayed by 5 minutes, and stops are 5 minutes early. Done in purpose to avoid time collisions on
// parallel scheduling. This can be improved a lot. Are you willing to contribute?
func GetBestSchedules(ctx *v1alpha1.Context) (schedules []Schedule, err error) {

	limitedCorrelativeRanges, err := GetLimitedCorrelativeHourRanges(ctx)
	if err != nil {
		return schedules, err
	}

	// Data from API is coming located in CEST. It's needed to parse it in that way
	apiTimeLocation, err := time.LoadLocation(ApagaLuzApiTimeLocation)

	for _, rangeItem := range limitedCorrelativeRanges {

		// Sort by hour (minor to major)
		sort.Slice(rangeItem, func(i, j int) bool {
			return rangeItem[i].Hour < rangeItem[j].Hour
		})

		// Get timestamp for the first item (always present)
		startTime, err := time.ParseInLocation(dateLayout, fmt.Sprintf("%s %d", rangeItem[0].Day, rangeItem[0].Hour), apiTimeLocation)
		if err != nil {
			return schedules, err
		}

		startTime = startTime.Add(5 * time.Minute)

		var stopTime time.Time

		// Only one item: in 1 hour stop it
		if len(rangeItem) == 1 {
			stopTime = startTime.Add(50 * time.Minute).In(apiTimeLocation)
			goto appendSchedule
		}

		// More items: stop it on last item's timestamp
		stopTime, err = time.ParseInLocation(dateLayout, fmt.Sprintf("%s %d", rangeItem[len(rangeItem)-1].Day, rangeItem[len(rangeItem)-1].Hour), apiTimeLocation)
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
