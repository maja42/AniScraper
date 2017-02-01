package main

import "github.com/maja42/AniScraper/webserver"

type Config struct {
	WebServerConfig webserver.WebServerConfig
}

func DefaultConfig() *Config {
	return &Config{
		WebServerConfig: webserver.DefaultWebServerConfig(),
	}
}
