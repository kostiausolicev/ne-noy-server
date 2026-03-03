package config

type VKClientConfig struct {
	ServiceKey string `mapstructure:"serviceKey"`
	BaseURL    string `mapstructure:"baseUrl"`
}
