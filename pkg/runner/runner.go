package runner

import (
	"net/http"

	"github.com/wojciechka/goket/pkg/eventprocessor"
	"go.uber.org/zap"
)

type Runner interface {
	Channel() chan *eventprocessor.Event
	Close()
}

type runner struct {
	logger       *zap.Logger
	eventChannel chan *eventprocessor.Event
	closeChannel chan struct{}
}

func NewRunner(logger *zap.Logger) Runner {
	r := &runner{
		logger:       logger,
		eventChannel: make(chan *eventprocessor.Event, 16),
		closeChannel: make(chan struct{}),
	}

	go r.processLoop()

	return r
}

func (r *runner) Channel() chan *eventprocessor.Event {
	return r.eventChannel
}

func (r *runner) Close() {
	r.closeChannel <- struct{}{}
}

func (r *runner) processLoop() {
	for {
		select {
		case event := <-r.eventChannel:
			r.runEvent(event)
		case <-r.closeChannel:
			return
		}
	}
}

func (r *runner) runEvent(event *eventprocessor.Event) {
	switch event.Type {
	case "", "url":
		go func() {
			logger := r.logger.With(zap.String("url", event.Action))
			logger.Info("invoking url")

			resp, err := http.Get(event.Action)
			if err != nil {
				logger.Error("url failed", zap.Error(err))
				return
			}

			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				logger.Error("url returned error", zap.Int("code", resp.StatusCode))
			} else {
				logger.Info("url request complete", zap.Int("code", resp.StatusCode))
			}
		}()
	default:
		r.logger.Error("unknown event type", zap.String("type", event.Type))
	}
}
