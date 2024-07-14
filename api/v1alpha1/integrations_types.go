package v1alpha1

// --
type TapoSmartPlugSpec struct {
	Client  string `yaml:"client"`
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
	} `yaml:"auth, omitempty"`
}
