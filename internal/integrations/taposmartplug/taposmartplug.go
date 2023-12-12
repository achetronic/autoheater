package taposmartplug

import (
	"encoding/json"
	"errors"
	"github.com/achetronic/autoheater/api/v1alpha1"
	tapogotypes "github.com/achetronic/tapogo/api/types"
	"github.com/achetronic/tapogo/pkg/tapogo"
	"github.com/avast/retry-go"
	"github.com/richardjennings/tapo/pkg/tapo"
	"time"
)

const (
	// Global messages
	RequiredConfigFieldsMissingMessage = "some mandatory config field is missing on Tapo smartplug integration"

	// Error messages
	TurningOffDuringRetriesError = "error turning off tapo smartplug device (retries left?): %s"
	TurningOnDuringRetriesError  = "error turning on tapo smartplug device (retries left?): %s"
	TurningOffError              = "error turning off tapo smartplug device: %s"
	TurningOnError               = "error turning on tapo smartPlug device: %s"
	ClientCreationError          = "tapo client failed on creation: %s"

	// Default values
	RequestRetryAttempts      = 10
	RequestRetryDelayDuration = time.Second * 2
)

// --
func checkConfigFields(ctx *v1alpha1.Context) (err error) {
	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	// Check required fields to act
	if tapoConfig.Address == "" ||
		tapoConfig.Auth.Username == "" ||
		tapoConfig.Auth.Password == "" ||
		tapoConfig.Client == "" {
		return errors.New(RequiredConfigFieldsMissingMessage)
	}

	return err
}

// TurnOnDevice send a request to tapo API to turn on the device
func TurnOnDevice(ctx *v1alpha1.Context) (tapoResponse map[string]interface{}, err error) {

	err = checkConfigFields(ctx)
	if err != nil {
		return tapoResponse, err
	}

	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	switch tapoConfig.Client {
	case "legacy":
		tapoClient, err := tapo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
		if err != nil {
			ctx.Logger.Errorf(ClientCreationError, err)
			return tapoResponse, err
		}

		tapoResponse, err = tapoClient.TurnOn()
		if err != nil {
			ctx.Logger.Errorf(TurningOnError, err)
		}

	default:
		// New KLAP protocol throws random errors when the requests are done at speed.
		// Retrying with a new token mostly solve the issue (jaquecoso...)
		tapoResponseNew := &tapogotypes.ResponseSpec{}
		err = retry.Do(
			func() (err error) {
				tapoClientNew, err := tapogo.NewTapo(tapoConfig.Address,
					tapoConfig.Auth.Username,
					tapoConfig.Auth.Password,
					&tapogo.TapoOptions{})
				if err != nil {
					ctx.Logger.Errorf(ClientCreationError, err)
					return err
				}

				tapoResponseNew, err = tapoClientNew.TurnOn()
				if err != nil {
					ctx.Logger.Errorf(TurningOnDuringRetriesError, err)
				}
				return err
			},
			retry.Attempts(RequestRetryAttempts),
			retry.Delay(RequestRetryDelayDuration))

		if err != nil {
			ctx.Logger.Errorf(TurningOnError, err)
			return tapoResponse, err
		}

		// Make response compatible with the global interface
		jsonBytes, err := json.Marshal(tapoResponseNew)
		if err != nil {
			return tapoResponse, err
		}

		err = json.Unmarshal(jsonBytes, &tapoResponse)
		if err != nil {
			return tapoResponse, err
		}
	}

	return tapoResponse, err
}

// TurnOffDevice send a request to tapo API to turn off the device
func TurnOffDevice(ctx *v1alpha1.Context) (tapoResponse map[string]interface{}, err error) {

	err = checkConfigFields(ctx)
	if err != nil {
		return tapoResponse, err
	}

	//
	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	switch tapoConfig.Client {
	case "legacy":
		tapoClient, err := tapo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
		if err != nil {
			ctx.Logger.Errorf(ClientCreationError, err)
			return tapoResponse, err
		}

		tapoResponse, err = tapoClient.TurnOff()
		if err != nil {
			ctx.Logger.Errorf(TurningOffError, err)
		}

	default:
		// New KLAP protocol throws random errors when the requests are done at speed.
		// Retrying with a new token mostly solve the issue (jaquecoso...)
		tapoResponseNew := &tapogotypes.ResponseSpec{}
		err = retry.Do(
			func() (err error) {
				tapoClientNew, err := tapogo.NewTapo(tapoConfig.Address,
					tapoConfig.Auth.Username,
					tapoConfig.Auth.Password,
					&tapogo.TapoOptions{})

				if err != nil {
					ctx.Logger.Errorf(ClientCreationError, err)
					return err
				}

				tapoResponseNew, err = tapoClientNew.TurnOff()
				if err != nil {
					ctx.Logger.Errorf(TurningOffDuringRetriesError, err)
				}
				return err
			},
			retry.Attempts(RequestRetryAttempts),
			retry.Delay(RequestRetryDelayDuration))

		if err != nil {
			ctx.Logger.Errorf(TurningOffError, err)
			return tapoResponse, err
		}

		// Make response compatible with the global interface
		jsonBytes, err := json.Marshal(tapoResponseNew)
		if err != nil {
			return tapoResponse, err
		}

		err = json.Unmarshal(jsonBytes, &tapoResponse)
		if err != nil {
			return tapoResponse, err
		}
	}

	return tapoResponse, err
}
