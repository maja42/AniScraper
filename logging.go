package main

import (
	lcf "github.com/Robpol86/logrus-custom-formatter"
	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/aniscraper"
	"github.com/maja42/AniScraper/utils"
	"github.com/maja42/AniScraper/webserver"
)

var log = logrus.New()

func SetupLogger(consoleLevel logrus.Level) {
	log.Level = consoleLevel

	lcf.WindowsEnableNativeANSI(true)
	template := "%[shortLevelName]s[%04[relativeCreated]d] %-45[message]s%[fields]s\n"
	log.Formatter = lcf.NewFormatter(template, nil)

	aniscraper.SetPackageLogger(log)
	utils.SetPackageLogger(log)
	webserver.SetPackageLogger(log)
}
