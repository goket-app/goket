package eventprocessor

import (
	"time"

	"go.uber.org/zap"
)

type processor struct {
	logger              *zap.Logger
	processEventChannel chan *Event

	lastTime     time.Time
	rootEventMap EventMap
	rootTimeout  float64

	currentEvent    *Event
	currentEventMap EventMap
	currentTimeout  float64

	closeChannel chan struct{}
	eventChannel chan string
}

func NewProcessor(logger *zap.Logger, eventMap EventMap, timeout float64, processEventChannel chan *Event) EventProcessor {
	p := &processor{
		processEventChannel: processEventChannel,
		logger:              logger,

		rootEventMap: eventMap,
		rootTimeout:  timeout,

		closeChannel: make(chan struct{}, 1),
		eventChannel: make(chan string, 16),
	}

	p.resetCurrentEvent()

	go p.processLoop()

	return p
}

func (p *processor) currentTimeoutDuration() time.Duration {
	return time.Duration(p.currentTimeout * float64(time.Second))
}

func (p *processor) Process(name string) {
	p.eventChannel <- name
}

func (p *processor) Close() {
	p.closeChannel <- struct{}{}
}

func (p *processor) checkIfTimedOut() {
	if p.currentEvent != nil && time.Now().After(p.lastTime.Add(p.currentTimeoutDuration())) {
		p.resetCurrentEvent()
	}
}

func (p *processor) resetCurrentEvent() {
	p.logger.Debug("resetting current event")
	p.currentEvent = nil
	p.currentEventMap = p.rootEventMap
	p.currentTimeout = p.rootTimeout

}

func (p *processor) processLoop() {
	var timeout time.Duration
	for {
		if p.currentEvent != nil {
			timeout = p.currentTimeoutDuration()
		} else {
			timeout = 60 * time.Second
		}

		select {
		case <-p.closeChannel:
			close(p.closeChannel)
			close(p.eventChannel)
			return
		case name := <-p.eventChannel:
			p.checkIfTimedOut()
			p.handleEvent(name)
		case <-time.After(timeout):
			p.performCurrentEvent()
		}
	}
}

func (p *processor) handleEvent(name string) {
	p.lastTime = time.Now()

	newEvent, ok := p.currentEventMap[name]

	// try keys from root map as well
	if !ok {
		newEvent, ok = p.rootEventMap[name]
	}

	if ok {
		p.currentEvent = newEvent

		if len(newEvent.Children) > 0 {
			p.logger.Info("handling event with children", zap.String("name", name))

			p.currentEventMap = newEvent.Children
			if newEvent.Timeout > 0 {
				p.currentTimeout = float64(newEvent.Timeout)
			}
		} else {
			p.logger.Info("handling event without children", zap.String("name", name))
			p.performCurrentEvent()
		}
	} else {
		p.logger.Info("invalid event received for current step", zap.String("name", name))
	}
}

func (p *processor) performCurrentEvent() {
	if p.currentEvent != nil {
		if p.currentEvent.Action != "" {
			p.logger.Info("sending event", zap.String("action", p.currentEvent.Action), zap.String("type", p.currentEvent.Type))
			p.processEventChannel <- p.currentEvent
		} else {
			p.logger.Info("ignoring empty event", zap.String("action", p.currentEvent.Action), zap.String("type", p.currentEvent.Type))
		}

		p.resetCurrentEvent()
	}
}
