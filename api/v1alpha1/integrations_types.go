package v1alpha1

// --
type TapoSmartPlugSpec struct {
	Address string `yaml:"address"`
	Auth    struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

// --
type WebhookSpec struct {
	URL  string `yaml:"url"`
	Auth struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}
