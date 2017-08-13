package main

import (
	lcf "github.com/Robpol86/logrus-custom-formatter"
	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/utils"
)

func SetupLogger(consoleLevel logrus.Level) utils.Logger {
	l := utils.NewStdLogger("")
	l.SetLevel(consoleLevel)

	lcf.WindowsEnableNativeANSI(true)
	template := "%[shortLevelName]s[%04[relativeCreated]d] %-150[message]s%[fields]s\n"
	l.SetFormatter(lcf.NewFormatter(template, nil))
	return l
}
