package schedules

import (
	"reflect"
	"sync"
	"time"

	"github.com/achetronic/autoheater/api/v1alpha1"
	"github.com/achetronic/autoheater/internal/integrations/taposmartplug"
	"github.com/achetronic/autoheater/internal/integrations/webhook"
	"github.com/achetronic/autoheater/internal/price"
	"github.com/achetronic/autoheater/internal/weather"
)

const (
	//
	RetryAttempts = 3
	RetryDelay    = 5 * time.Second

	//
	RootSchedulerStartedMessage = "task scheduler is running @ %s"
	WaitingNextDayMessage       = "waiting until next day to schedule actions"
	WeatherNotSuitableMessage   = "weather is not suitable to turn on the device"

	StartDeviceProgrammedActionMessage = "task programmed. device will be turned on @ %s"
	StartDeviceExecutedActionMessage   = "task completed. device has been turned on @ %s"

	// --
	StopDeviceProgrammedActionMessage = "task programmed. device will be turned off @ %s"
	StopDeviceExecutedActionMessage   = "task completed. device has been turned off @ %s"

	WeatherNotAvailableErrorMessage         = "impossible to determine whether it's cold in your coordinates"
	TapoStartExecutionFailedErrorMessage    = "error executing start action for 'tapo smartplug' integration: %s"
	TapoStopExecutionFailedErrorMessage     = "error executing stop action for 'tapo smartplug' integration: %s"
	WebhookStartExecutionFailedErrorMessage = "error executing start action for 'webhook' integration: %s"
	WebhookStopExecutionFailedErrorMessage  = "error executing stop action for 'tapo webhook' integration: %s"
)

// RunScheduler run scheduling function periodically.
// It's executed always in the beginning of the day as it's the moment when the PVPC prices are really known
func RunScheduler(ctx *v1alpha1.Context) {

	var err error
	var retryFunctionErr error

	var isCold bool
	var schedules []price.Schedule

	for {
		currentTime := time.Now().In(time.Local)

		// Disable the scheduler in (hot days for heaters) && (cold days for coolers)
		if ctx.Config.Spec.Weather.Enabled {

			retryFunctionErr = retry(func() error {
				isCold, err = weather.IsColdDay(ctx)
				return err
			}, RetryAttempts, RetryDelay)

			if retryFunctionErr != nil {
				ctx.Logger.Infof(WeatherNotAvailableErrorMessage)
				goto waitNextDay
			}

			// Warm day, not enable the heater
			// Cold day, not enable the cooler
			if (ctx.Config.Spec.Device.Type == "heater" && !isCold) ||
				(ctx.Config.Spec.Device.Type == "cooler" && isCold) {
				ctx.Logger.Infof(WeatherNotSuitableMessage)
				goto waitNextDay
			}
		}

		// Get the sections with the best prices to satisfy the hours required by the user
		retryFunctionErr = retry(func() error {
			schedules, err = price.GetBestSchedules(ctx)
			return err
		}, RetryAttempts, RetryDelay)

		if retryFunctionErr != nil {
			ctx.Logger.Infof(price.PricesNotAvailableErrorMessage)
			goto waitNextDay
		}

		//
		ctx.Logger.Infof(RootSchedulerStartedMessage, time.Now().In(time.Local).Format(time.RFC822))
		ScheduleActions(ctx, schedules)

	waitNextDay:
		// Wait until next programmed hour (following day)
		// By default, next scheduling moment is 12:00 AM
		ctx.Logger.Infof(WaitingNextDayMessage)
		nextDay := currentTime.Add(24 * time.Hour)
		nextTargetTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 1, 0, 0, time.Local)
		duration := nextTargetTime.Sub(currentTime)
		time.Sleep(duration)
	}
}

// ScheduleActions create goroutines to execute actions delayed until moments given by schedules list.
// To change the timezone of programmed tasks, just set TZ environment variable to desired one, i.e: TZ=Atlantic/Canary
func ScheduleActions(ctx *v1alpha1.Context, schedules []price.Schedule) {

	// Send a signal to stop the device before scheduling new goroutines.
	// This is to avoid keeping the device turned on in expensive hours in case this CLI failed in the middle
	// of some time range, and restarted after the range finished
	ExecuteStopAction(ctx)

	// Get current time
	currentTime := time.Now().In(time.Local)

	//
	syncScheduleWait := sync.WaitGroup{}

	// Iterate over schedules
	for _, schedule := range schedules {

		//
		scheduleStartTime := schedule.Start.In(time.Local)
		durationUntilStart := scheduleStartTime.Sub(currentTime)

		scheduleStopTime := schedule.Stop.In(time.Local)
		durationUntilStop := scheduleStopTime.Sub(currentTime)

		//
		beforeTheRange := durationUntilStart > 0
		insideTheRange := durationUntilStart < 0 && durationUntilStop > 0

		// Create a goroutine to execute some actions in desired moment (start event)
		if beforeTheRange || insideTheRange {
			syncScheduleWait.Add(1)

			// Starting moment is passed, but still have time to start
			if insideTheRange {
				durationUntilStart = 0
			}

			go func(startTime time.Time, startIn time.Duration) {
				ctx.Logger.Infof(StartDeviceProgrammedActionMessage, startTime.Format(time.RFC822))
				syncScheduleWait.Done()

				time.Sleep(startIn)
				ExecuteStartAction(ctx)

				ctx.Logger.Infof(StartDeviceExecutedActionMessage, startTime.Format(time.RFC822))
			}(scheduleStartTime, durationUntilStart)
		}

		// Create a goroutine to execute some actions in desired moment (stop event)
		if durationUntilStop > 0 {
			syncScheduleWait.Add(1)

			go func(stopTime time.Time, stopIn time.Duration) {
				ctx.Logger.Infof(StopDeviceProgrammedActionMessage, stopTime.Format(time.RFC822))
				syncScheduleWait.Done()

				time.Sleep(stopIn)
				ExecuteStopAction(ctx)

				ctx.Logger.Infof(StopDeviceExecutedActionMessage, stopTime.Format(time.RFC822))
			}(scheduleStopTime, durationUntilStop)
		}

		// Wait until both goroutines have been scheduled
		syncScheduleWait.Wait()
	}
}

// ExecuteStartAction execute an action for each defined integration on 'start' events
func ExecuteStartAction(ctx *v1alpha1.Context) {
	var err error

	// Execute the action for Tapo Smart plug device when its config is present
	if !reflect.ValueOf(ctx.Config.Spec.Device.Integrations.TapoSmartPlug).IsZero() {
		_, err := taposmartplug.TurnOnDevice(ctx)
		if err != nil {
			ctx.Logger.Infof(TapoStartExecutionFailedErrorMessage, err)
		}
	}

	// Execute the action for webhook device when its config is present
	if !reflect.ValueOf(ctx.Config.Spec.Device.Integrations.Webhook).IsZero() {
		_, err = webhook.SendStartDeviceEvent(ctx)
		if err != nil {
			ctx.Logger.Infof(WebhookStartExecutionFailedErrorMessage, err)
		}
	}
}

// ExecuteStopAction execute an action for each defined integration on 'stop' events
func ExecuteStopAction(ctx *v1alpha1.Context) {
	var err error

	// Execute the action for Tapo Smartplug device when its config is present
	if !reflect.ValueOf(ctx.Config.Spec.Device.Integrations.TapoSmartPlug).IsZero() {
		_, err := taposmartplug.TurnOffDevice(ctx)
		if err != nil {
			ctx.Logger.Infof(TapoStopExecutionFailedErrorMessage, err)
		}
	}

	// Execute the action for webhook device when its config is present
	if !reflect.ValueOf(ctx.Config.Spec.Device.Integrations.Webhook).IsZero() {
		_, err = webhook.SendStopDeviceEvent(ctx)
		if err != nil {
			ctx.Logger.Infof(WebhookStopExecutionFailedErrorMessage, err)
		}
	}
}
