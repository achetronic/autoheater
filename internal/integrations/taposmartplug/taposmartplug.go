package taposmartplug

import (
	"encoding/json"
	"errors"
	"github.com/achetronic/autoheater/api/v1alpha1"
	tapogotypes "github.com/achetronic/tapogo/api/types"
	"github.com/achetronic/tapogo/pkg/tapogo"
	"github.com/richardjennings/tapo/pkg/tapo"
	"time"
)

const (
	RequiredConfigFieldsMissingMessage = "some mandatory config field is missing on Tapo SmartPlug integration"

	RequestRetryAttempts = 3
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
			return tapoResponse, err
		}

		tapoResponse, err = tapoClient.TurnOn()

	default:
		tapoClientNew, err := tapogo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
		if err != nil {
			return tapoResponse, err
		}

		// New KLAP protocol throws random errors when the requests are done at speed.
		// Retrying mostly solve the issue
		tapoResponseNew := &tapogotypes.ResponseSpec{}
		for retryAttempt := 0; retryAttempt < RequestRetryAttempts; retryAttempt++ {
			tapoResponseNew, err = tapoClientNew.TurnOn()
			if err != nil {
				time.Sleep(time.Millisecond * 250) // Magic number 0.25s
				continue
			}
			break
		}

		if err != nil {
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
			return tapoResponse, err
		}

		tapoResponse, err = tapoClient.TurnOff()

	default:
		tapoClientNew, err := tapogo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
		if err != nil {
			return tapoResponse, err
		}

		// New KLAP protocol throws random errors when the requests are done at speed.
		// Retrying mostly solve the issue
		tapoResponseNew := &tapogotypes.ResponseSpec{}
		for retryAttempt := 0; retryAttempt < RequestRetryAttempts; retryAttempt++ {
			tapoResponseNew, err = tapoClientNew.TurnOff()
			if err != nil {
				time.Sleep(time.Millisecond * 250) // Magic number 0.25s
				continue
			}
			break
		}

		if err != nil {
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
