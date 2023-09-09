package price

import (
	"encoding/json"
	"fmt"
	"github.com/achetronic/autoheater/api/v1alpha1"
	"io"
	"log"
	"net/http"
	"sort"
)

const ApagaLuzAPIUrl = "https://raw.githubusercontent.com/jorgeatgu/apaga-luz/main/public/data/today_price.json"

// ApagaLuzResponseSpec represents TODO
type ApagaLuzResponseSpec struct {
	Day   string  `json:"day"`
	Hour  int     `json:"hour"`
	Price float64 `json:"price"`
	Zone  string  `json:"zone"`
}

// ApagaLuzResponseSpec represents TODO
type ApagaLuzResponseListSpec []ApagaLuzResponseSpec

// TODO DEVOLVER EL ERROR PA EVALUAR
func GetApiData(autoheater *v1alpha1.Autoheater) (response *ApagaLuzResponseListSpec, err error) {

	// Weather not enabled, just throw empty data
	if !autoheater.Spec.Weather.Enabled {
		return response, nil
	}

	// Check fields regarding coordinates
	if autoheater.Spec.Weather.Coordinates.Latitude == 0 || autoheater.Spec.Weather.Coordinates.Longitude == 0 {
		log.Fatal("coordinates section is required to evaluate weather")
	}

	// Send the request and wait for the result
	resp, err := http.Get(ApagaLuzAPIUrl)
	if err != nil {
		fmt.Println("Error al hacer la solicitud HTTP:", err)
		return
	}
	defer resp.Body.Close()

	// Read the request's body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer el cuerpo de la respuesta:", err)
		return
	}

	// Decode response's JSON into a struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Print(err)
	}

	log.Print(response)

	sort.Slice(*response, func(i, j int) bool {
		return (*response)[i].Price < (*response)[j].Price
	})

	log.Print("ordenadooooo")
	log.Print(response)

	return response, nil
}
