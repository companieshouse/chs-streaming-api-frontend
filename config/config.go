package config

import (
	"github.com/companieshouse/gofigure"
)

// Config represents the frontend application configuration.
type Config struct {
	gofigure                interface{} `order:"env,flag"`
	BindAddr                string      `env:"BIND_ADDR"                    flag:"bind-addr"              flagDesc:"Bind address"`
	CertFile                string      `env:"CERT_FILE"                    flag:"cert-file"              flagDesc:"Certificate file"`
	KeyFile                 string      `env:"KEY_FILE"                     flag:"key-file"               flagDesc:"Key file"`
	EricURL                 string      `env:"ERIC_LOCAL_URL"               flag:"eric-url"               flagDesc:"Eric url"`
	FilingLogs              string      `env:"LOG_TOPIC"                    flag:"filing-logs"            flagDesc:"Log frontend streaming api"`
	RequestTimeout          int         `env:"REQUEST_TIMEOUT"              flag:"request-timeout"        flagDesc:"Request timeout in seconds"`
	HeartbeatInterval       int         `env:"HEARTBEAT_INTERVAL"           flag:"heartbeat-interval"     flagDesc:"Heartbeat interval in seconds"`
	CacheBrokerURL          string      `env:"CACHE_BROKER_URL"             flag:"cache-broker-url"       flagDesc:"Cache broker url"`
	OfficersEndpointFlag    bool        `env:"OFFICERS_ENDPOINT_ENABLED"    flag:"officers-endpoint-flag" flagDesc:"Flag to enable/disable the Officers stream"`
	PSCsEndpointFlag        bool        `env:"PSCS_ENDPOINT_ENABLED"        flag:"pscs-endpoint-flag"     flagDesc:"Flag to enable/disable the PSCs stream"`
}

// ServiceConfig returns a ServiceConfig interface for Config.
func (c Config) ServiceConfig() ServiceConfig {
	return ServiceConfig{c}
}

// ServiceConfig wraps Config to implement service.Config.
type ServiceConfig struct {
	Config
}

// BindAddr implements service.Config.BindAddr.
func (cfg ServiceConfig) BindAddr() string {
	return cfg.Config.BindAddr
}

// CertFile implements service.Config.CertFile.
func (cfg ServiceConfig) CertFile() string {
	return cfg.Config.CertFile
}

// KeyFile implements service.Config.KeyFile.
func (cfg ServiceConfig) KeyFile() string {
	return cfg.Config.KeyFile
}

// Namespace implements service.Config.Namespace.
func (cfg ServiceConfig) Namespace() string {
	return "chs-streaming-api-frontend"
}

// Namespace implements service.Config.Namespace
func (c *Config) Namespace() string {
	return "chs"
}

var cfg *Config

// Get configures the application and returns the default configuration.
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:          ":3124",
		CertFile:          "",
		KeyFile:           "",
		EricURL:           "",
		FilingLogs:        "",
		RequestTimeout:    86400,
		HeartbeatInterval: 30,
		CacheBrokerURL:    "",
	}

	err := gofigure.Gofigure(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
