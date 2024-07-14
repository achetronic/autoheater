package schedules

import (
	"time"
)

func retry(function func() error, attempts int, timeBetweenAttempts time.Duration) (err error) {

	var functionError error

	for attempt := 0; attempt < attempts; attempt++ {
		functionError = function()
		if functionError == nil {
			break
		}

		time.Sleep(timeBetweenAttempts)
	}

	if functionError != nil {
		return functionError
	}

	return nil
}
