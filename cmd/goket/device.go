package main

import (
	"strings"

	"github.com/goket-app/goket/pkg/evdevinput"
	"github.com/goket-app/goket/pkg/eventprocessor"
	"github.com/goket-app/goket/pkg/runner"
	"go.uber.org/zap"
)

func handleDevice(logger *zap.Logger, device string, r runner.Runner, eventMap eventprocessor.EventMap, timeout float64) {
	processor := eventprocessor.NewProcessor(logger, eventMap, timeout, r.Channel())
	processor.Start()

	for {
		logger.Info("Opening device")

		dev, err := evdevinput.NewEvdevKeyboardInput(device)
		if err != nil {
			logger.Error("Failed to open device", zap.Error(err))
			return
		}

		for {
			key, err := dev.Read()
			if err != nil {
				logger.Error("Failed to read key", zap.Error(err))
				break
			}

			if key.Down {
				processor.Process(key.KeyName, key.Time)
			}
		}

		err = dev.Close()
		if err != nil {
			logger.Error("Failed to close device", zap.Error(err))
			return
		}
	}
}

func toDeviceList(devices string) []string {
	var result []string

	for _, device := range strings.Split(devices, ",") {
		if device != "" {
			result = append(result, device)
		}
	}
	return result
}
