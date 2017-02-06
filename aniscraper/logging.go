package aniscraper

import "github.com/Sirupsen/logrus"

var log *logrus.Logger

func SetPackageLogger(logger *logrus.Logger) {
	log = logger
}
