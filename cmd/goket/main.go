package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/goket-app/goket/pkg/evdevinput"
	"github.com/goket-app/goket/pkg/runner"
	"go.uber.org/zap"
)

var configPath string
var devices string
var timeout float64

func main() {
	flag.StringVar(&configPath, "config", "/etc/goket.json", "path to config file")
	flag.StringVar(&devices, "devices", "", "path to devices")
	flag.Float64Var(&timeout, "timeout", 2.0, "timeout between keypresses, in seconds")

	flag.Parse()

	logger, err := zap.NewProduction()
	defer logger.Sync()

	if err != nil {
		log.Fatal(err)
	}

	deviceList := toDeviceList(devices)
	if len(deviceList) > 0 {
		deviceList = strings.Split(devices, ",")
	} else {
		var err error
		deviceList, err = evdevinput.ListDevices()
		if err != nil {
			logger.Error("Failed to list devices", zap.Error(err))
			return
		}
	}

	config, err := getConfig(configPath)
	if err != nil {
		logger.Error("Failed to read configuration", zap.Error(err))
		return
	}

	r := runner.NewRunner(logger.With(zap.String("component", "runner")))

	for _, device := range deviceList {
		go handleDevice(logger.With(zap.String("component", "device"), zap.String("device", device)), device, r, config.Keys, timeout)
	}

	// wait indefinitely
	for {
		time.Sleep(time.Second * 86400)
	}
}
