package webserver

import "time"

type WebServerConfig struct {
	AddressBinding string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

func DefaultWebServerConfig() WebServerConfig {
	return WebServerConfig{
		AddressBinding: ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}
}
