package taposmartplug

import (
	"errors"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"github.com/richardjennings/tapo/pkg/tapo"
)

const (
	RequiredConfigFieldsMissingMessage = "some mandatory config field is missing on Tapo SmartPlug integration"
)

// --
func checkConfigFields(ctx *v1alpha1.Context) (err error) {
	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	// Check required fields to act
	if tapoConfig.Address == "" || tapoConfig.Auth.Username == "" || tapoConfig.Auth.Password == "" {
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

	//
	var tapoClient *tapo.Tapo

	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	tapoClient, err = tapo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
	if err != nil {
		return tapoResponse, err
	}

	tapoResponse, err = tapoClient.TurnOn()

	return tapoResponse, err
}

// TurnOffDevice send a request to tapo API to turn off the device
func TurnOffDevice(ctx *v1alpha1.Context) (tapoResponse map[string]interface{}, err error) {

	err = checkConfigFields(ctx)
	if err != nil {
		return tapoResponse, err
	}

	//
	var tapoClient *tapo.Tapo

	tapoConfig := ctx.Config.Spec.Device.Integrations.TapoSmartPlug

	tapoClient, err = tapo.NewTapo(tapoConfig.Address, tapoConfig.Auth.Username, tapoConfig.Auth.Password)
	if err != nil {
		return tapoResponse, err
	}

	tapoResponse, err = tapoClient.TurnOff()

	return tapoResponse, err
}
