package config

import (
	"os"

	"github.com/achetronic/autoheater/api/v1alpha1"

	"gopkg.in/yaml.v3"
)

// Marshal TODO
func Marshal(config v1alpha1.Autoheater) (bytes []byte, err error) {
	bytes, err = yaml.Marshal(config)
	return bytes, err
}

// Unmarshal TODO
func Unmarshal(bytes []byte) (config v1alpha1.Autoheater, err error) {
	err = yaml.Unmarshal(bytes, &config)
	return config, err
}

// ReadFile TODO
func ReadFile(filepath string) (config v1alpha1.Autoheater, err error) {
	var fileBytes []byte
	fileBytes, err = os.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	config, err = Unmarshal(fileBytes)

	return config, err
}
