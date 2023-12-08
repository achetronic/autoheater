package v1alpha1

// --
type TapoSmartPlugSpec struct {
	Client  string `yaml:"client"`
	Address string `yaml:"address"`
	Auth    struct {
		Username string `yaml:"username" env:"TAPO_SMARTPLUG_USERNAME"`
		Password string `yaml:"password" env:"TAPO_SMARTPLUG_PASSWORD"`
	} `yaml:"auth"`
}

// --
type WebhookSpec struct {
	URL  string `yaml:"url"`
	Auth struct {
		Username string `yaml:"username" env:"WEBHOOK_USERNAME"`
		Password string `yaml:"password" env:"WEBHOOK_PASSWORD"`
	} `yaml:"auth"`
}
