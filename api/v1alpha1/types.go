package v1alpha1

// Autoheater TODO
type Autoheater struct {
	ApiVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   AutoheaterMetadataT `yaml:"metadata"`
	Spec       AutoheaterSpec      `yaml:"spec"`
}

// TODO
type AutoheaterMetadataT struct {
	Name string `yaml:"name"`
}

// AutoheaterSpec TODO
type AutoheaterSpec struct {
	Synchronization SynchronizationSpec `yaml:"synchronization"`
	Device          DeviceSpec          `yaml:"device"`
	Weather         WeatherSpec         `yaml:"weather"`
	Price           PriceSpec           `yaml:"price"`
}

// SynchronizationSpec TODO
type SynchronizationSpec struct {
	Time string `yaml:"time"`
}

// --
type DeviceSpec struct {
	ActiveHours    int                `yaml:"activeHours"`
	Implementation ImplementationSpec `yaml:"implementation"`
}

type ImplementationSpec struct {

	// TODO
	TapoSmartPlug struct {
		Address string `yaml:"address"`
		Auth    struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"auth"`
	} `yaml:"tapoSmartPlug,omitempty"`

	// TODO
	Webhook struct {
		URL string `yaml:"url"`
	} `yaml:"webhook,omitempty"`
}

// WeatherSpec TODO
type WeatherSpec struct {
	Enabled     bool            `yaml:"enabled"`
	Coordinates CoordinatesSpec `yaml:"coordinates,omitempty"`
	Temperature TemperatureSpec `yaml:"temperature,omitempty"`
}

// CoordinatesSpec TODO
type CoordinatesSpec struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

// TemperatureSpec TODO
type TemperatureSpec struct {
	Type      string `yaml:"type"`
	Unit      string `yaml:"unit"`
	Threshold int    `yaml:"threshold"`
}

// PriceSpec TODO
type PriceSpec struct {
	Zone string `yaml:"zone"`
}
