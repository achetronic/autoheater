package schedules

import (
	"fmt"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"github.com/achetronic/autoheater/internal/price"
	"sync"
	"time"
)

const (
	RootSchedulerStartedMessage = "Ejecutando el programador de tareas (%s) \n"

	StartDeviceProgrammedActionMessage = "Tarea programada. Encender el dispositivo (%s) \n"
	StartDeviceExecutedActionMessage   = "Tarea ejecutada. Encender el dispositivo (%s) \n"

	StopDeviceProgrammedActionMessage = "Tarea programada. Apagar el dispositivo (%s) \n"
	StopDeviceExecutedActionMessage   = "Tarea ejecutada. Apagar el dispositivo (%s) \n"
)

// RunScheduler run scheduling function periodically.
// It's executed always in the beginning of the day as it's the moment when the PVPC prices are really known
func RunScheduler(autoheater *v1alpha1.Autoheater, schedules []price.Schedule) {

	for {
		currentTime := time.Now().In(time.Local)

		//
		fmt.Printf(RootSchedulerStartedMessage, time.Now().In(time.Local))
		ScheduleActions(autoheater, schedules)

		// Wait until next programmed hour (following day)
		// By default, next scheduling moment is 12:00 AM
		nextDay := currentTime.Add(24 * time.Hour)
		nextTargetTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 1, 0, 0, time.Local)
		duration := nextTargetTime.Sub(currentTime)
		time.Sleep(duration)
	}
}

// ScheduleActions create goroutines to execute actions delayed until moments given by schedules list.
// To change the timezone of programmed tasks, just set TZ environment variable to desired one, i.e: TZ=Atlantic/Canary
func ScheduleActions(autoheater *v1alpha1.Autoheater, schedules []price.Schedule) {

	// Get current time
	currentTime := time.Now().In(time.Local)

	//
	syncScheduleWait := sync.WaitGroup{}

	// Iterate over schedules
	for _, schedule := range schedules {

		// Create a goroutine to execute some actions in desired moment (start event)
		scheduleStartTime := schedule.Start.In(time.Local)
		durationUntilStart := scheduleStartTime.Sub(currentTime)

		if durationUntilStart > 0 {
			syncScheduleWait.Add(1)

			go func(startTime time.Time, startIn time.Duration) {
				fmt.Printf(StartDeviceProgrammedActionMessage, startTime)
				syncScheduleWait.Done()

				time.Sleep(startIn)
				ExecuteStartAction(autoheater)

				fmt.Printf(StartDeviceExecutedActionMessage, startTime)
			}(scheduleStartTime, durationUntilStart)
		}

		// Create a goroutine to execute some actions in desired moment (stop event)
		scheduleStopTime := schedule.Stop.In(time.Local)
		durationUntilStop := scheduleStopTime.Sub(currentTime)

		if durationUntilStop > 0 {
			syncScheduleWait.Add(1)

			go func(stopTime time.Time, stopIn time.Duration) {
				fmt.Printf(StopDeviceProgrammedActionMessage, stopTime)
				syncScheduleWait.Done()

				time.Sleep(stopIn)
				ExecuteStopAction(autoheater)

				fmt.Printf(StopDeviceExecutedActionMessage, stopTime)
			}(scheduleStopTime, durationUntilStop)
		}

		// Wait until both goroutines have been scheduled
		syncScheduleWait.Wait()
	}
}

// ---
func ExecuteStartAction(autoheater *v1alpha1.Autoheater) {
	fmt.Printf("Hola, soy una Goroutina que va a encenderte la vida \n")
}

// --
func ExecuteStopAction(autoheater *v1alpha1.Autoheater) {
	fmt.Printf("Hola, soy OTRA goroutine que va a apagarte la vida \n")
}
