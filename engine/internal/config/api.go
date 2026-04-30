package config

type APIConfig struct {
	Port int `yaml:"port"`
}

func (c *APIConfig) Validate() error {
	return nil
}
