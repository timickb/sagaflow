package config

type BackofficeConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `yaml:"port"`
}
