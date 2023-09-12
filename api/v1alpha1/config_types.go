package v1alpha1

// ConfigSpec TODO
type ConfigSpec struct {
	ApiVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   MetadataSpec      `yaml:"metadata"`
	Spec       SpecificationSpec `yaml:"spec"`
}

// MetadataSpec TODO
type MetadataSpec struct {
	Name string `yaml:"name"`
}

// SpecificationSpec TODO
type SpecificationSpec struct {
	Global  GlobalSpec  `yaml:"global"`
	Device  DeviceSpec  `yaml:"device"`
	Weather WeatherSpec `yaml:"weather"`
	Price   PriceSpec   `yaml:"price"`
}

// GlobalSpec TODO
type GlobalSpec struct {
	IgnorePassedHours bool `yaml:"ignorePassedHours,omitempty"`
}

// DeviceSpec TODO
type DeviceSpec struct {
	Type         string           `yaml:"type"`
	ActiveHours  int              `yaml:"activeHours"`
	Integrations IntegrationsSpec `yaml:"integrations"`
}

// IntegrationsSpec TODO
type IntegrationsSpec struct {

	// TODO
	TapoSmartPlug TapoSmartPlugSpec `yaml:"tapoSmartPlug,omitempty"`

	// TODO
	Webhook WebhookSpec `yaml:"webhook,omitempty"`
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
