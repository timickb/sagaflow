package config

type BackofficeConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `yaml:"port"`
}

func (c *BackofficeConfig) Validate() error {
	return nil
}
