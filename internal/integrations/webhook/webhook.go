package webhook

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/achetronic/autoheater/api/v1alpha1"
)

const (
	HttpEventPattern = `{"event":"%s","name":"%s","timestamp":"%s"}`
	HttpEventVerb    = "POST"

	HttpRequestCreationErrorMessage = "error creating http request: %s"
	HttpRequestSendingErrorMessage  = "error sending http request: %s"

	RequiredConfigFieldsMissingMessage = "some mandatory config field is missing on webhooks integration"
)

// --
func checkConfigFields(ctx *v1alpha1.Context) (err error) {
	webhookConfig := ctx.Config.Spec.Device.Integrations.Webhook

	// Check required fields to act
	if webhookConfig.URL == "" {
		return errors.New(RequiredConfigFieldsMissingMessage)
	}

	return err
}

// sendEvent send an HTTP request with the content '{"event":"%s","name":"%s","timestamp":"%s"}'
func sendEvent(ctx *v1alpha1.Context, event string) (httpResponse *http.Response, err error) {
	//
	httpClient := &http.Client{}

	webhookConfig := ctx.Config.Spec.Device.Integrations.Webhook

	// Crear una solicitud POST
	httpRequest, err := http.NewRequest(HttpEventVerb, webhookConfig.URL, nil)
	if err != nil {
		return httpResponse, errors.New(fmt.Sprintf(HttpRequestCreationErrorMessage, err))
	}

	// Agregar autenticación básica
	if webhookConfig.Auth.Username != "" && webhookConfig.Auth.Password != "" {
		httpRequest.SetBasicAuth(webhookConfig.Auth.Username, webhookConfig.Auth.Password)
	}

	// Add data to the request
	payload := []byte(fmt.Sprintf(HttpEventPattern, event, ctx.Config.Metadata.Name, time.Now().In(time.Local)))
	httpRequest.Body = io.NopCloser(bytes.NewBuffer(payload))
	httpRequest.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	httpResponse, err = httpClient.Do(httpRequest)
	if err != nil {
		return httpResponse, errors.New(fmt.Sprintf(HttpRequestSendingErrorMessage, err))
	}
	defer httpResponse.Body.Close()

	//
	return httpResponse, err
}

// SendStartDeviceEvent send a request to TODO
func SendStartDeviceEvent(ctx *v1alpha1.Context) (httpResponse *http.Response, err error) {

	err = checkConfigFields(ctx)
	if err != nil {
		return httpResponse, err
	}

	//
	httpResponse, err = sendEvent(ctx, "start")
	return httpResponse, err
}

// SendStopDeviceEvent send a request to TODO
func SendStopDeviceEvent(ctx *v1alpha1.Context) (httpResponse *http.Response, err error) {

	err = checkConfigFields(ctx)
	if err != nil {
		return httpResponse, err
	}

	//
	httpResponse, err = sendEvent(ctx, "stop")
	return httpResponse, err
}
