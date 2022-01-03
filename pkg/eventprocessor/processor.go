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

	currentEvent *Event

	closeChannel chan struct{}
	eventChannel chan *processorEvent
}

type processorEvent struct {
	name string
	when time.Time
}

func NewProcessor(logger *zap.Logger, eventMap EventMap, timeout float64, processEventChannel chan *Event) EventProcessor {
	p := &processor{
		processEventChannel: processEventChannel,
		logger:              logger,

		rootEventMap: eventMap,
		rootTimeout:  timeout,

		eventChannel: make(chan *processorEvent, 16),
	}

	return p
}

func (p *processor) Start() {
	p.Close()
	p.closeChannel = make(chan struct{}, 1)
	go p.processLoop()
}

func (p *processor) Process(name string, when time.Time) {
	p.eventChannel <- &processorEvent{name: name, when: when}
}

func (p *processor) Close() {
	if p.closeChannel != nil {
		p.closeChannel <- struct{}{}
	}
}

func (p *processor) currentEventMap() EventMap {
	if p.currentEvent != nil {
		return p.currentEvent.Children
	}
	return p.rootEventMap
}

func (p *processor) currentTimeoutDuration() time.Duration {
	seconds := p.rootTimeout
	currentEvent := p.currentEvent
	for currentEvent != nil {
		if currentEvent.Timeout > 0 {
			seconds = currentEvent.Timeout
			break
		}
		currentEvent = currentEvent.parent
	}
	return time.Duration(seconds * float64(time.Second))
}

func (p *processor) checkIfTimedOut(when time.Time) {
	// only
	if p.previousEventTimedOut(when) {
		p.findParentCurrentEvent(when)
	}
}

func (p *processor) previousEventTimedOut(when time.Time) bool {
	return p.currentEvent != nil && when.After(p.lastTime.Add(p.currentTimeoutDuration()))
}

// Find parent event that has Stay set to true and has not timed out ; returns nil if no parent found
func (p *processor) findParentCurrentEvent(when time.Time) {
	p.currentEvent = p.currentEvent.parent

	for p.currentEvent != nil {
		if p.currentEvent.Stay && !p.previousEventTimedOut(when) {
			break
		}
		p.currentEvent = p.currentEvent.parent
	}
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
		case event := <-p.eventChannel:
			p.processEvent(event)
		case <-time.After(timeout):
			p.performCurrentEvent()
		}
	}
}

func (p *processor) processEvent(event *processorEvent) {
	p.checkIfTimedOut(event.when)
	p.handleEvent(event.name, event.when)
}

func (p *processor) handleEvent(name string, now time.Time) {
	p.lastTime = now

	newEvent, ok := p.currentEventMap()[name]

	// try keys from root map as well
	if !ok {
		newEvent, ok = p.rootEventMap[name]
	}

	if ok {
		p.currentEvent = newEvent

		if len(newEvent.Children) > 0 {
			p.logger.Info("handling event with children", zap.String("name", name))
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

		p.findParentCurrentEvent(p.lastTime)
	}
}
